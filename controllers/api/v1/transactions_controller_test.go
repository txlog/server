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
