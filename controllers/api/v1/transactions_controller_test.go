package v1

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/txlog/server/models"
)

func setupTransactionsTestDB(t *testing.T) *sql.DB {
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

func cleanupTransactionsTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM transaction_items WHERE machine_id LIKE 'tx-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup transaction_items: %v", err)
	}
	_, err = db.Exec("DELETE FROM transactions WHERE machine_id LIKE 'tx-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup transactions: %v", err)
	}
}

func TestGetTransactionIDs_InvalidRawData(t *testing.T) {
	db := setupTransactionsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/transactions/ids", GetTransactionIDs(db))

	// Test with no body - this should still work but return empty results
	req, _ := http.NewRequest("GET", "/v1/transactions/ids", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty body results in empty JSON unmarshal, which is valid
	// But machine_id and hostname will be empty, so query returns empty
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetTransactionIDs_InvalidJSON(t *testing.T) {
	db := setupTransactionsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/transactions/ids", GetTransactionIDs(db))

	req, _ := http.NewRequest("GET", "/v1/transactions/ids", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetTransactionIDs_EmptyResult(t *testing.T) {
	db := setupTransactionsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/transactions/ids", GetTransactionIDs(db))

	body := models.Transaction{
		MachineID: "nonexistent-machine",
		Hostname:  "nonexistent-host",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("GET", "/v1/transactions/ids", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []int
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 0 {
		t.Errorf("Expected empty array, got %d items", len(response))
	}
}

func TestGetTransactionIDs_ReturnsIDs(t *testing.T) {
	db := setupTransactionsTestDB(t)
	defer db.Close()
	defer cleanupTransactionsTestData(t, db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/transactions/ids", GetTransactionIDs(db))

	machineID := "tx-test-machine-001"
	hostname := "tx-test-hostname"

	// Insert test transactions
	for i := 1; i <= 3; i++ {
		_, err := db.Exec(`
			INSERT INTO transactions (transaction_id, machine_id, hostname, actions, altered, "user", return_code, release_version, command_line, comment, scriptlet_output)
			VALUES ($1, $2, $3, 'Install', '1', 'root', '0', '8.5', 'dnf install test', '', '')`,
			i*10, machineID, hostname)
		if err != nil {
			t.Fatalf("Failed to insert test transaction %d: %v", i, err)
		}
	}

	body := models.Transaction{
		MachineID: machineID,
		Hostname:  hostname,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("GET", "/v1/transactions/ids", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response []int
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 3 {
		t.Errorf("Expected 3 transaction IDs, got %d", len(response))
	}

	// Verify they are in ascending order
	expectedIDs := []int{10, 20, 30}
	for i, id := range response {
		if id != expectedIDs[i] {
			t.Errorf("Expected transaction ID %d at index %d, got %d", expectedIDs[i], i, id)
		}
	}
}

func TestGetTransactionIDs_FiltersCorrectly(t *testing.T) {
	db := setupTransactionsTestDB(t)
	defer db.Close()
	defer cleanupTransactionsTestData(t, db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/transactions/ids", GetTransactionIDs(db))

	machineID1 := "tx-test-machine-filter-1"
	machineID2 := "tx-test-machine-filter-2"
	hostname := "tx-test-hostname-filter"

	// Insert transactions for machine 1
	_, err := db.Exec(`
		INSERT INTO transactions (transaction_id, machine_id, hostname, actions, altered, "user", return_code, release_version, command_line, comment, scriptlet_output)
		VALUES (100, $1, $2, 'Install', '1', 'root', '0', '8.5', 'dnf install test', '', '')`,
		machineID1, hostname)
	if err != nil {
		t.Fatalf("Failed to insert transaction for machine1: %v", err)
	}

	// Insert transactions for machine 2
	_, err = db.Exec(`
		INSERT INTO transactions (transaction_id, machine_id, hostname, actions, altered, "user", return_code, release_version, command_line, comment, scriptlet_output)
		VALUES (200, $1, $2, 'Install', '1', 'root', '0', '8.5', 'dnf install test', '', '')`,
		machineID2, hostname)
	if err != nil {
		t.Fatalf("Failed to insert transaction for machine2: %v", err)
	}

	// Query for machine 1 only
	body := models.Transaction{
		MachineID: machineID1,
		Hostname:  hostname,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("GET", "/v1/transactions/ids", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []int
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 1 {
		t.Errorf("Expected 1 transaction ID, got %d", len(response))
	}

	if len(response) > 0 && response[0] != 100 {
		t.Errorf("Expected transaction ID 100, got %d", response[0])
	}
}

// Tests for POST /v1/transactions

func cleanupPostTransactionsTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM transaction_items WHERE machine_id LIKE 'post-tx-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup transaction_items: %v", err)
	}
	_, err = db.Exec("DELETE FROM transactions WHERE machine_id LIKE 'post-tx-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup transactions: %v", err)
	}
	_, err = db.Exec("DELETE FROM assets WHERE machine_id LIKE 'post-tx-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup assets: %v", err)
	}
}

func TestPostTransactions_InvalidRawData(t *testing.T) {
	db := setupTransactionsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/transactions", PostTransactions(db))

	req, _ := http.NewRequest("POST", "/v1/transactions", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty body results in empty JSON unmarshal error
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPostTransactions_InvalidJSON(t *testing.T) {
	db := setupTransactionsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/transactions", PostTransactions(db))

	req, _ := http.NewRequest("POST", "/v1/transactions", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPostTransactions_CreatesTransaction(t *testing.T) {
	db := setupTransactionsTestDB(t)
	defer db.Close()
	defer cleanupPostTransactionsTestData(t, db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/transactions", PostTransactions(db))

	machineID := "post-tx-test-machine-001"
	hostname := "post-tx-test-hostname"
	transactionID := "1001"

	body := models.Transaction{
		TransactionID:  transactionID,
		MachineID:      machineID,
		Hostname:       hostname,
		Actions:        "Install",
		Altered:        "1",
		User:           "root",
		ReturnCode:     "0",
		ReleaseVersion: "8.5",
		CommandLine:    "dnf install test-package",
		Comment:        "",
		Items: []models.TransactionItem{
			{
				Action:  "Install",
				Name:    "test-package",
				Version: "1.0.0",
				Release: "1.el8",
				Epoch:   "0",
				Arch:    "x86_64",
				Repo:    "baseos",
			},
		},
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/v1/transactions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Verify transaction was created
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM transactions WHERE machine_id = $1 AND transaction_id = $2", machineID, transactionID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query transactions: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 transaction, got %d", count)
	}

	// Verify items were created
	err = db.QueryRow("SELECT COUNT(*) FROM transaction_items WHERE machine_id = $1 AND transaction_id = $2", machineID, transactionID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query transaction_items: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 transaction item, got %d", count)
	}
}

func TestPostTransactions_DuplicateTransaction(t *testing.T) {
	db := setupTransactionsTestDB(t)
	defer db.Close()
	defer cleanupPostTransactionsTestData(t, db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/transactions", PostTransactions(db))

	machineID := "post-tx-test-machine-002"
	hostname := "post-tx-test-hostname-dup"
	transactionID := "1002"

	body := models.Transaction{
		TransactionID:  transactionID,
		MachineID:      machineID,
		Hostname:       hostname,
		Actions:        "Install",
		Altered:        "1",
		User:           "root",
		ReturnCode:     "0",
		ReleaseVersion: "8.5",
		CommandLine:    "dnf install test-package",
	}
	jsonBody, _ := json.Marshal(body)

	// First request - should succeed
	req1, _ := http.NewRequest("POST", "/v1/transactions", bytes.NewBuffer(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("First request failed with status %d", w1.Code)
	}

	// Second request with same transaction_id and machine_id - should return "already exists"
	req2, _ := http.NewRequest("POST", "/v1/transactions", bytes.NewBuffer(jsonBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w2.Code)
	}

	var response map[string]string
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Transaction already exists" {
		t.Errorf("Expected 'Transaction already exists' message, got '%s'", response["message"])
	}
}

func TestPostTransactions_WithMultipleItems(t *testing.T) {
	db := setupTransactionsTestDB(t)
	defer db.Close()
	defer cleanupPostTransactionsTestData(t, db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/transactions", PostTransactions(db))

	machineID := "post-tx-test-machine-003"
	hostname := "post-tx-test-hostname-multi"
	transactionID := "1003"

	body := models.Transaction{
		TransactionID:  transactionID,
		MachineID:      machineID,
		Hostname:       hostname,
		Actions:        "Install",
		Altered:        "5",
		User:           "root",
		ReturnCode:     "0",
		ReleaseVersion: "8.5",
		CommandLine:    "dnf install pkg1 pkg2 pkg3 pkg4 pkg5",
		Items: []models.TransactionItem{
			{Action: "Install", Name: "pkg1", Version: "1.0", Release: "1.el8", Epoch: "0", Arch: "x86_64", Repo: "baseos"},
			{Action: "Install", Name: "pkg2", Version: "2.0", Release: "1.el8", Epoch: "0", Arch: "x86_64", Repo: "baseos"},
			{Action: "Install", Name: "pkg3", Version: "3.0", Release: "1.el8", Epoch: "0", Arch: "x86_64", Repo: "appstream"},
			{Action: "Install", Name: "pkg4", Version: "4.0", Release: "1.el8", Epoch: "0", Arch: "noarch", Repo: "appstream"},
			{Action: "Install", Name: "pkg5", Version: "5.0", Release: "1.el8", Epoch: "0", Arch: "x86_64", Repo: "epel"},
		},
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/v1/transactions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Verify all items were created
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM transaction_items WHERE machine_id = $1 AND transaction_id = $2", machineID, transactionID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query transaction_items: %v", err)
	}
	if count != 5 {
		t.Errorf("Expected 5 transaction items, got %d", count)
	}
}
