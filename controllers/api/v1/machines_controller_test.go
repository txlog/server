package v1

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/txlog/server/models"
)

func setupMachinesTestDB(t *testing.T) *sql.DB {
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=txlog_test sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skip("Skipping test: PostgreSQL not available")
	}

	if err := db.Ping(); err != nil {
		t.Skip("Skipping test: Cannot connect to PostgreSQL")
	}

	return db
}

func cleanupMachinesTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM executions WHERE hostname LIKE 'machines-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup executions: %v", err)
	}
	_, err = db.Exec("DELETE FROM assets WHERE hostname LIKE 'machines-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup assets: %v", err)
	}
}

func createTestAssetWithExecution(t *testing.T, db *sql.DB, hostname, machineID, os, agentVersion string) {
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	timestamp := time.Now()

	// Insert execution first
	_, err = tx.Exec(`
		INSERT INTO executions (
			machine_id, hostname, executed_at, success, 
			transactions_processed, transactions_sent, agent_version, os
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		machineID, hostname, timestamp, true, 5, 3, agentVersion, os,
	)
	if err != nil {
		t.Fatalf("Failed to insert execution: %v", err)
	}

	// Insert asset
	am := models.NewAssetManager(db)
	err = am.UpsertAsset(tx, hostname, machineID, timestamp, sql.NullBool{}, sql.NullString{}, "")
	if err != nil {
		t.Fatalf("Failed to insert asset: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}
}

func TestGetMachines_ReturnsActiveAssets(t *testing.T) {
	db := setupMachinesTestDB(t)
	defer db.Close()
	defer cleanupMachinesTestData(t, db)

	// Create test assets
	createTestAssetWithExecution(t, db, "machines-test-server01", "machines-test-001", "Linux", "1.0.0")
	createTestAssetWithExecution(t, db, "machines-test-server02", "machines-test-002", "Windows", "1.0.0")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/machines", GetMachines(db))

	req, _ := http.NewRequest("GET", "/v1/machines", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var machines []MachineID
	err := json.Unmarshal(w.Body.Bytes(), &machines)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(machines) < 2 {
		t.Errorf("Expected at least 2 machines, got %d", len(machines))
	}

	// Verify our test machines are in the response
	foundServer01 := false
	foundServer02 := false
	for _, m := range machines {
		if m.Hostname == "machines-test-server01" && m.MachineID == "machines-test-001" {
			foundServer01 = true
		}
		if m.Hostname == "machines-test-server02" && m.MachineID == "machines-test-002" {
			foundServer02 = true
		}
	}

	if !foundServer01 {
		t.Error("Expected to find machines-test-server01 in response")
	}
	if !foundServer02 {
		t.Error("Expected to find machines-test-server02 in response")
	}
}

func TestGetMachines_FilterByOS(t *testing.T) {
	db := setupMachinesTestDB(t)
	defer db.Close()
	defer cleanupMachinesTestData(t, db)

	// Create test assets with different OS
	createTestAssetWithExecution(t, db, "machines-test-linux01", "machines-test-linux-001", "Linux", "1.0.0")
	createTestAssetWithExecution(t, db, "machines-test-windows01", "machines-test-win-001", "Windows", "1.0.0")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/machines", GetMachines(db))

	req, _ := http.NewRequest("GET", "/v1/machines?os=Linux", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var machines []MachineID
	err := json.Unmarshal(w.Body.Bytes(), &machines)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify only Linux machines are returned
	for _, m := range machines {
		if m.Hostname == "machines-test-windows01" {
			t.Error("Windows machine should not be in Linux filter results")
		}
	}

	// Verify Linux machine is present
	foundLinux := false
	for _, m := range machines {
		if m.Hostname == "machines-test-linux01" {
			foundLinux = true
			break
		}
	}

	if !foundLinux {
		t.Error("Expected to find Linux machine in filtered results")
	}
}

func TestGetMachines_FilterByAgentVersion(t *testing.T) {
	db := setupMachinesTestDB(t)
	defer db.Close()
	defer cleanupMachinesTestData(t, db)

	// Create test assets with different agent versions
	createTestAssetWithExecution(t, db, "machines-test-v1", "machines-test-v1-001", "Linux", "1.0.0")
	createTestAssetWithExecution(t, db, "machines-test-v2", "machines-test-v2-001", "Linux", "2.0.0")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/machines", GetMachines(db))

	req, _ := http.NewRequest("GET", "/v1/machines?agent_version=1.0.0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var machines []MachineID
	err := json.Unmarshal(w.Body.Bytes(), &machines)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify only v1.0.0 machines are returned
	for _, m := range machines {
		if m.Hostname == "machines-test-v2" {
			t.Error("v2.0.0 machine should not be in v1.0.0 filter results")
		}
	}

	// Verify v1.0.0 machine is present
	foundV1 := false
	for _, m := range machines {
		if m.Hostname == "machines-test-v1" {
			foundV1 = true
			break
		}
	}

	if !foundV1 {
		t.Error("Expected to find v1.0.0 machine in filtered results")
	}
}

func TestGetMachines_OnlyActiveAssets(t *testing.T) {
	db := setupMachinesTestDB(t)
	defer db.Close()
	defer cleanupMachinesTestData(t, db)

	hostname := "machines-test-inactive"
	oldMachineID := "machines-test-inactive-old"
	newMachineID := "machines-test-inactive-new"

	// Create old asset
	createTestAssetWithExecution(t, db, hostname, oldMachineID, "Linux", "1.0.0")

	// Create new asset (this will deactivate the old one)
	createTestAssetWithExecution(t, db, hostname, newMachineID, "Linux", "1.0.0")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/machines", GetMachines(db))

	req, _ := http.NewRequest("GET", "/v1/machines", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var machines []MachineID
	err := json.Unmarshal(w.Body.Bytes(), &machines)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify only new machine is in results
	foundOld := false
	foundNew := false
	for _, m := range machines {
		if m.MachineID == oldMachineID {
			foundOld = true
		}
		if m.MachineID == newMachineID {
			foundNew = true
		}
	}

	if foundOld {
		t.Error("Old inactive machine should not be in results")
	}

	if !foundNew {
		t.Error("New active machine should be in results")
	}
}

func TestGetMachines_CombinedFilters(t *testing.T) {
	db := setupMachinesTestDB(t)
	defer db.Close()
	defer cleanupMachinesTestData(t, db)

	// Create test assets with various combinations
	createTestAssetWithExecution(t, db, "machines-test-combo01", "machines-test-combo-001", "Linux", "1.0.0")
	createTestAssetWithExecution(t, db, "machines-test-combo02", "machines-test-combo-002", "Linux", "2.0.0")
	createTestAssetWithExecution(t, db, "machines-test-combo03", "machines-test-combo-003", "Windows", "1.0.0")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/machines", GetMachines(db))

	req, _ := http.NewRequest("GET", "/v1/machines?os=Linux&agent_version=1.0.0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var machines []MachineID
	err := json.Unmarshal(w.Body.Bytes(), &machines)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should only find machines-test-combo01
	foundCombo01 := false
	for _, m := range machines {
		if m.Hostname == "machines-test-combo01" {
			foundCombo01 = true
		}
		if m.Hostname == "machines-test-combo02" {
			t.Error("combo02 (Linux v2.0.0) should not match filter")
		}
		if m.Hostname == "machines-test-combo03" {
			t.Error("combo03 (Windows v1.0.0) should not match filter")
		}
	}

	if !foundCombo01 {
		t.Error("Expected to find combo01 (Linux v1.0.0) in results")
	}
}
