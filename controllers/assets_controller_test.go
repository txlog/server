package controllers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/txlog/server/models"
)

func setupAssetsTestDB(t *testing.T) *sql.DB {
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

func cleanupAssetsControllerTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM executions WHERE hostname LIKE 'assets-ctrl-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup executions: %v", err)
	}
	_, err = db.Exec("DELETE FROM assets WHERE hostname LIKE 'assets-ctrl-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup assets: %v", err)
	}
	_, err = db.Exec("DELETE FROM transaction_items WHERE transaction_id IN (SELECT transaction_id FROM transactions WHERE machine_id LIKE 'assets-ctrl-test-%')")
	if err != nil {
		t.Logf("Warning: Failed to cleanup transaction_items: %v", err)
	}
	_, err = db.Exec("DELETE FROM transactions WHERE machine_id LIKE 'assets-ctrl-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup transactions: %v", err)
	}
}

func createTestAssetWithData(t *testing.T, db *sql.DB, hostname, machineID, os string, needsRestart bool, lastSeen time.Time) {
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Insert execution
	_, err = tx.Exec(`
		INSERT INTO executions (
			machine_id, hostname, executed_at, success,
			transactions_processed, transactions_sent, agent_version, os, needs_restarting
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		machineID, hostname, lastSeen, true, 5, 3, "1.0.0", os, needsRestart,
	)
	if err != nil {
		t.Fatalf("Failed to insert execution: %v", err)
	}

	// Insert asset with needs_restarting information
	am := models.NewAssetManager(db)
	var needsRestartingNull sql.NullBool
	if needsRestart {
		needsRestartingNull = sql.NullBool{Bool: needsRestart, Valid: true}
	}
	err = am.UpsertAsset(tx, hostname, machineID, lastSeen, needsRestartingNull, sql.NullString{}, "")
	if err != nil {
		t.Fatalf("Failed to insert asset: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}
}

func TestGetAssetsIndex_NoFilters(t *testing.T) {
	db := setupAssetsTestDB(t)
	defer db.Close()
	defer cleanupAssetsControllerTestData(t, db)

	// Create test assets
	now := time.Now()
	createTestAssetWithData(t, db, "assets-ctrl-test-server01", "assets-ctrl-test-001", "Linux", false, now)
	createTestAssetWithData(t, db, "assets-ctrl-test-server02", "assets-ctrl-test-002", "Windows", false, now)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.LoadHTMLGlob("../../templates/*.html")
	router.GET("/assets", GetAssetsIndex(db))

	req, _ := http.NewRequest("GET", "/assets", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}
}

func TestGetAssetsIndex_SearchByHostname(t *testing.T) {
	db := setupAssetsTestDB(t)
	defer db.Close()
	defer cleanupAssetsControllerTestData(t, db)

	now := time.Now()
	createTestAssetWithData(t, db, "assets-ctrl-test-findme", "assets-ctrl-test-003", "Linux", false, now)
	createTestAssetWithData(t, db, "assets-ctrl-test-other", "assets-ctrl-test-004", "Linux", false, now)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.LoadHTMLGlob("../../templates/*.html")
	router.GET("/assets", GetAssetsIndex(db))

	req, _ := http.NewRequest("GET", "/assets?search=findme", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetAssetsIndex_SearchByMachineID(t *testing.T) {
	db := setupAssetsTestDB(t)
	defer db.Close()
	defer cleanupAssetsControllerTestData(t, db)

	now := time.Now()
	machineID := "12345678901234567890123456789012" // 32 characters
	createTestAssetWithData(t, db, "assets-ctrl-test-machine", machineID, "Linux", false, now)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.LoadHTMLGlob("../../templates/*.html")
	router.GET("/assets", GetAssetsIndex(db))

	req, _ := http.NewRequest("GET", "/assets?search="+machineID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetAssetsIndex_RestartFilter(t *testing.T) {
	db := setupAssetsTestDB(t)
	defer db.Close()
	defer cleanupAssetsControllerTestData(t, db)

	now := time.Now()
	createTestAssetWithData(t, db, "assets-ctrl-test-restart", "assets-ctrl-test-005", "Linux", true, now)
	createTestAssetWithData(t, db, "assets-ctrl-test-norestart", "assets-ctrl-test-006", "Linux", false, now)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.LoadHTMLGlob("../../templates/*.html")
	router.GET("/assets", GetAssetsIndex(db))

	req, _ := http.NewRequest("GET", "/assets?restart=true", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetAssetsIndex_InactiveFilter(t *testing.T) {
	db := setupAssetsTestDB(t)
	defer db.Close()
	defer cleanupAssetsControllerTestData(t, db)

	now := time.Now()
	oldTime := now.Add(-20 * 24 * time.Hour) // 20 days ago

	createTestAssetWithData(t, db, "assets-ctrl-test-active", "assets-ctrl-test-007", "Linux", false, now)
	createTestAssetWithData(t, db, "assets-ctrl-test-inactive", "assets-ctrl-test-008", "Linux", false, oldTime)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.LoadHTMLGlob("../../templates/*.html")
	router.GET("/assets", GetAssetsIndex(db))

	req, _ := http.NewRequest("GET", "/assets?inactive=true", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetAssetsIndex_Pagination(t *testing.T) {
	db := setupAssetsTestDB(t)
	defer db.Close()
	defer cleanupAssetsControllerTestData(t, db)

	now := time.Now()
	createTestAssetWithData(t, db, "assets-ctrl-test-page01", "assets-ctrl-test-p01", "Linux", false, now)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.LoadHTMLGlob("../../templates/*.html")
	router.GET("/assets", GetAssetsIndex(db))

	req, _ := http.NewRequest("GET", "/assets?page=2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestDeleteMachineID_Success(t *testing.T) {
	db := setupAssetsTestDB(t)
	defer db.Close()
	defer cleanupAssetsControllerTestData(t, db)

	machineID := "assets-ctrl-test-delete-001"
	now := time.Now()
	createTestAssetWithData(t, db, "assets-ctrl-test-todelete", machineID, "Linux", false, now)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/assets/:machine_id", DeleteMachineID(db))

	req, _ := http.NewRequest("DELETE", "/assets/"+machineID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %d", w.Code)
	}

	// Verify asset was deleted
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM assets WHERE machine_id = $1", machineID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count assets: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 assets, found %d", count)
	}
}

func TestDeleteMachineID_EmptyMachineID(t *testing.T) {
	db := setupAssetsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.LoadHTMLGlob("../../templates/*.html")
	router.DELETE("/assets/:machine_id", DeleteMachineID(db))

	req, _ := http.NewRequest("DELETE", "/assets/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should result in 404 since route won't match
	if w.Code == http.StatusOK {
		t.Error("Expected non-OK status for empty machine_id")
	}
}

func TestGetMachineID_Success(t *testing.T) {
	db := setupAssetsTestDB(t)
	defer db.Close()
	defer cleanupAssetsControllerTestData(t, db)

	machineID := "assets-ctrl-test-view-001"
	now := time.Now()
	createTestAssetWithData(t, db, "assets-ctrl-test-view", machineID, "Linux", false, now)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.LoadHTMLGlob("../../templates/*.html")
	router.GET("/assets/:machine_id", GetMachineID(db))

	req, _ := http.NewRequest("GET", "/assets/"+machineID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestGetMachineID_NotFound(t *testing.T) {
	db := setupAssetsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.LoadHTMLGlob("../../templates/*.html")
	router.GET("/assets/:machine_id", GetMachineID(db))

	req, _ := http.NewRequest("GET", "/assets/nonexistent-machine-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError && w.Code != http.StatusNotFound {
		t.Errorf("Expected error status for non-existent machine_id, got %d", w.Code)
	}
}

func TestGetMachineID_WithMultipleAssetsSameHostname(t *testing.T) {
	db := setupAssetsTestDB(t)
	defer db.Close()
	defer cleanupAssetsControllerTestData(t, db)

	hostname := "assets-ctrl-test-multiple"
	machineID1 := "assets-ctrl-test-multi-001"
	machineID2 := "assets-ctrl-test-multi-002"
	now := time.Now()

	createTestAssetWithData(t, db, hostname, machineID1, "Linux", false, now.Add(-1*time.Hour))
	createTestAssetWithData(t, db, hostname, machineID2, "Linux", false, now)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.LoadHTMLGlob("../../templates/*.html")
	router.GET("/assets/:machine_id", GetMachineID(db))

	req, _ := http.NewRequest("GET", "/assets/"+machineID2, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}
}

func TestDeleteMachineID_Isolation(t *testing.T) {
	db := setupAssetsTestDB(t)
	defer db.Close()
	defer cleanupAssetsControllerTestData(t, db)

	machineID1 := "assets-ctrl-test-iso-001"
	machineID2 := "assets-ctrl-test-iso-002"
	hostname := "assets-ctrl-test-iso-host"
	now := time.Now()
	txID := 99999

	// Create assets and executions
	createTestAssetWithData(t, db, hostname, machineID1, "Linux", false, now)
	createTestAssetWithData(t, db, hostname, machineID2, "Linux", false, now)

	// Insert transactions with SAME transaction_id but DIFFERENT machine_id
	_, err := db.Exec(`
		INSERT INTO transactions (transaction_id, machine_id, hostname, begin_time, end_time)
		VALUES ($1, $2, $3, $4, $5), ($1, $6, $3, $4, $5)`,
		txID, machineID1, hostname, now, now, machineID2,
	)
	if err != nil {
		t.Fatalf("Failed to insert transactions: %v", err)
	}

	// Insert transaction items for both
	_, err = db.Exec(`
		INSERT INTO transaction_items (transaction_id, machine_id, package, version, release, arch)
		VALUES
		($1, $2, 'pkg-1', '1.0', '1', 'x86_64'),
		($1, $3, 'pkg-1', '1.0', '1', 'x86_64')`,
		txID, machineID1, machineID2,
	)
	if err != nil {
		t.Fatalf("Failed to insert transaction items: %v", err)
	}

	// Delete machineID1
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/assets/:machine_id", DeleteMachineID(db))

	req, _ := http.NewRequest("DELETE", "/assets/"+machineID1, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %d", w.Code)
	}

	// Verify machineID1 items are gone
	var count1 int
	err = db.QueryRow("SELECT COUNT(*) FROM transaction_items WHERE machine_id = $1", machineID1).Scan(&count1)
	if err != nil {
		t.Fatalf("Failed to count items for machine 1: %v", err)
	}
	if count1 != 0 {
		t.Errorf("Expected 0 items for machine 1, found %d", count1)
	}

	// Verify machineID2 items STILL EXIST
	var count2 int
	err = db.QueryRow("SELECT COUNT(*) FROM transaction_items WHERE machine_id = $1", machineID2).Scan(&count2)
	if err != nil {
		t.Fatalf("Failed to count items for machine 2: %v", err)
	}
	if count2 != 1 {
		t.Errorf("Expected 1 item for machine 2, found %d. This indicates the bug regression!", count2)
	}
}
