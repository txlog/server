package v1

import (
	"bytes"
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

func setupAPITestDB(t *testing.T) *sql.DB {
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

func cleanupAPITestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM executions WHERE hostname LIKE 'api-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup executions: %v", err)
	}
	_, err = db.Exec("DELETE FROM assets WHERE hostname LIKE 'api-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup assets: %v", err)
	}
}

func TestPostExecutions_CreatesAsset(t *testing.T) {
	db := setupAPITestDB(t)
	defer db.Close()
	defer cleanupAPITestData(t, db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/executions", PostExecutions(db))

	hostname := "api-test-server01"
	machineID := "api-test-machine-001"
	timestamp := time.Now()

	execution := models.Execution{
		Hostname:              hostname,
		MachineID:             machineID,
		ExecutedAt:            &timestamp,
		Success:               true,
		TransactionsProcessed: 5,
		TransactionsSent:      3,
		AgentVersion:          "1.0.0",
		OS:                    "Linux",
	}

	body, _ := json.Marshal(execution)
	req, _ := http.NewRequest("POST", "/v1/executions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify asset was created
	am := models.NewAssetManager(db)
	asset, err := am.GetActiveAsset(hostname)
	if err != nil {
		t.Fatalf("Failed to get asset: %v", err)
	}

	if asset.Hostname != hostname {
		t.Errorf("Expected hostname %s, got %s", hostname, asset.Hostname)
	}

	if asset.MachineID != machineID {
		t.Errorf("Expected machine_id %s, got %s", machineID, asset.MachineID)
	}

	if !asset.IsActive {
		t.Errorf("Expected asset to be active")
	}
}

func TestPostExecutions_UpdatesAsset(t *testing.T) {
	db := setupAPITestDB(t)
	defer db.Close()
	defer cleanupAPITestData(t, db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/executions", PostExecutions(db))

	hostname := "api-test-server02"
	machineID := "api-test-machine-002"
	timestamp1 := time.Now().Add(-1 * time.Hour)
	timestamp2 := time.Now()

	// First execution
	execution1 := models.Execution{
		Hostname:              hostname,
		MachineID:             machineID,
		ExecutedAt:            &timestamp1,
		Success:               true,
		TransactionsProcessed: 5,
		TransactionsSent:      3,
	}

	body1, _ := json.Marshal(execution1)
	req1, _ := http.NewRequest("POST", "/v1/executions", bytes.NewBuffer(body1))
	req1.Header.Set("Content-Type", "application/json")

	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("First execution failed with status %d", w1.Code)
	}

	// Second execution
	execution2 := models.Execution{
		Hostname:              hostname,
		MachineID:             machineID,
		ExecutedAt:            &timestamp2,
		Success:               true,
		TransactionsProcessed: 10,
		TransactionsSent:      5,
	}

	body2, _ := json.Marshal(execution2)
	req2, _ := http.NewRequest("POST", "/v1/executions", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Second execution failed with status %d", w2.Code)
	}

	// Verify asset was updated
	am := models.NewAssetManager(db)
	asset, err := am.GetActiveAsset(hostname)
	if err != nil {
		t.Fatalf("Failed to get asset: %v", err)
	}

	// Check that last_seen is close to timestamp2
	timeDiff := asset.LastSeen.Sub(timestamp2)
	if timeDiff < -1*time.Second || timeDiff > 1*time.Second {
		t.Errorf("Expected last_seen to be ~%v, got %v (diff: %v)", timestamp2, asset.LastSeen, timeDiff)
	}
}

func TestPostExecutions_ReplacesAsset(t *testing.T) {
	db := setupAPITestDB(t)
	defer db.Close()
	defer cleanupAPITestData(t, db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/executions", PostExecutions(db))

	hostname := "api-test-server03"
	oldMachineID := "api-test-machine-003-old"
	newMachineID := "api-test-machine-003-new"
	timestamp := time.Now()

	// First execution with old machine
	execution1 := models.Execution{
		Hostname:              hostname,
		MachineID:             oldMachineID,
		ExecutedAt:            &timestamp,
		Success:               true,
		TransactionsProcessed: 5,
		TransactionsSent:      3,
	}

	body1, _ := json.Marshal(execution1)
	req1, _ := http.NewRequest("POST", "/v1/executions", bytes.NewBuffer(body1))
	req1.Header.Set("Content-Type", "application/json")

	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("First execution failed with status %d", w1.Code)
	}

	// Second execution with new machine
	execution2 := models.Execution{
		Hostname:              hostname,
		MachineID:             newMachineID,
		ExecutedAt:            &timestamp,
		Success:               true,
		TransactionsProcessed: 10,
		TransactionsSent:      5,
	}

	body2, _ := json.Marshal(execution2)
	req2, _ := http.NewRequest("POST", "/v1/executions", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Second execution failed with status %d", w2.Code)
	}

	// Verify new asset is active
	am := models.NewAssetManager(db)
	asset, err := am.GetActiveAsset(hostname)
	if err != nil {
		t.Fatalf("Failed to get active asset: %v", err)
	}

	if asset.MachineID != newMachineID {
		t.Errorf("Expected active machine_id %s, got %s", newMachineID, asset.MachineID)
	}

	// Verify old asset is inactive
	oldAsset, err := am.GetAssetByMachineID(oldMachineID)
	if err != nil {
		t.Fatalf("Failed to get old asset: %v", err)
	}

	if oldAsset.IsActive {
		t.Errorf("Expected old asset to be inactive")
	}

	if oldAsset.DeactivatedAt == nil {
		t.Errorf("Expected old asset to have deactivated_at set")
	}
}

func TestPostExecutions_InvalidJSON(t *testing.T) {
	db := setupAPITestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/executions", PostExecutions(db))

	req, _ := http.NewRequest("POST", "/v1/executions", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestPostExecutions_WithNullTimestamp(t *testing.T) {
	db := setupAPITestDB(t)
	defer db.Close()
	defer cleanupAPITestData(t, db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/executions", PostExecutions(db))

	hostname := "api-test-server04"
	machineID := "api-test-machine-004"

	execution := models.Execution{
		Hostname:              hostname,
		MachineID:             machineID,
		ExecutedAt:            nil,
		Success:               true,
		TransactionsProcessed: 5,
		TransactionsSent:      3,
	}

	body, _ := json.Marshal(execution)
	req, _ := http.NewRequest("POST", "/v1/executions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify asset was created with approximate timestamp
	am := models.NewAssetManager(db)
	asset, err := am.GetActiveAsset(hostname)
	if err != nil {
		t.Fatalf("Failed to get asset: %v", err)
	}

	if asset.MachineID != machineID {
		t.Errorf("Expected machine_id %s, got %s", machineID, asset.MachineID)
	}

	// Check that the asset has a recent timestamp
	timeDiff := time.Since(asset.LastSeen)
	if timeDiff > 5*time.Second {
		t.Errorf("Expected last_seen to be recent, got %v ago", timeDiff)
	}
}
