package v1

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/txlog/server/models"
)

func setupItemsTestDB(t *testing.T) *sql.DB {
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

func cleanupItemsTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM transaction_items WHERE machine_id LIKE 'items-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup transaction_items: %v", err)
	}
	_, err = db.Exec("DELETE FROM transactions WHERE machine_id LIKE 'items-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup transactions: %v", err)
	}
}

func TestGetItems_RequiresMachineID(t *testing.T) {
	db := setupItemsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/items", GetItems(db))

	req, _ := http.NewRequest("GET", "/v1/items", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetItems_NoTransactionsFound(t *testing.T) {
	db := setupItemsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/items", GetItems(db))

	req, _ := http.NewRequest("GET", "/v1/items?machine_id=nonexistent-machine", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetItems_ReturnsTransactionWithItems(t *testing.T) {
	db := setupItemsTestDB(t)
	defer db.Close()
	defer cleanupItemsTestData(t, db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/items", GetItems(db))

	machineID := "items-test-machine-001"
	transactionID := 1

	// Insert test transaction
	_, err := db.Exec(`
		INSERT INTO transactions (transaction_id, machine_id, hostname, actions, altered, "user", return_code, release_version, command_line, comment, scriptlet_output)
		VALUES ($1, $2, 'test-hostname', 'Install', '1', 'root', '0', '8.5', 'dnf install test', '', '')`,
		transactionID, machineID)
	if err != nil {
		t.Fatalf("Failed to insert test transaction: %v", err)
	}

	// Insert test items
	_, err = db.Exec(`
		INSERT INTO transaction_items (transaction_id, machine_id, action, package, version, release, epoch, arch, repo, from_repo)
		VALUES ($1, $2, 'Install', 'test-package', '1.0.0', '1.el8', '0', 'x86_64', 'baseos', '')`,
		transactionID, machineID)
	if err != nil {
		t.Fatalf("Failed to insert test item: %v", err)
	}

	req, _ := http.NewRequest("GET", "/v1/items?machine_id="+machineID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response models.Transaction
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Hostname != "test-hostname" {
		t.Errorf("Expected hostname 'test-hostname', got '%s'", response.Hostname)
	}

	if len(response.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(response.Items))
	} else {
		if response.Items[0].Name != "test-package" {
			t.Errorf("Expected package 'test-package', got '%s'", response.Items[0].Name)
		}
	}
}

func TestGetItemIDs_RequiresMachineID(t *testing.T) {
	db := setupItemsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/items/ids", GetItemIDs(db))

	req, _ := http.NewRequest("GET", "/v1/items/ids", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetItemIDs_NoTransactionsFound(t *testing.T) {
	db := setupItemsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/items/ids", GetItemIDs(db))

	req, _ := http.NewRequest("GET", "/v1/items/ids?machine_id=nonexistent-machine", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetItemIDs_ReturnsItemIDs(t *testing.T) {
	db := setupItemsTestDB(t)
	defer db.Close()
	defer cleanupItemsTestData(t, db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/items/ids", GetItemIDs(db))

	machineID := "items-test-machine-002"
	transactionID := 2

	// Insert test transaction
	_, err := db.Exec(`
		INSERT INTO transactions (transaction_id, machine_id, hostname, actions, altered, "user", return_code, release_version, command_line, comment, scriptlet_output)
		VALUES ($1, $2, 'test-hostname', 'Install', '1', 'root', '0', '8.5', 'dnf install test', '', '')`,
		transactionID, machineID)
	if err != nil {
		t.Fatalf("Failed to insert test transaction: %v", err)
	}

	// Insert test items
	_, err = db.Exec(`
		INSERT INTO transaction_items (transaction_id, machine_id, action, package, version, release, epoch, arch, repo, from_repo)
		VALUES ($1, $2, 'Install', 'test-package-1', '1.0.0', '1.el8', '0', 'x86_64', 'baseos', '')`,
		transactionID, machineID)
	if err != nil {
		t.Fatalf("Failed to insert test item 1: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO transaction_items (transaction_id, machine_id, action, package, version, release, epoch, arch, repo, from_repo)
		VALUES ($1, $2, 'Install', 'test-package-2', '2.0.0', '1.el8', '0', 'x86_64', 'appstream', '')`,
		transactionID, machineID)
	if err != nil {
		t.Fatalf("Failed to insert test item 2: %v", err)
	}

	req, _ := http.NewRequest("GET", "/v1/items/ids?machine_id="+machineID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response []int
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 item IDs, got %d", len(response))
	}
}

func TestGetItems_SpecificTransactionID(t *testing.T) {
	db := setupItemsTestDB(t)
	defer db.Close()
	defer cleanupItemsTestData(t, db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/items", GetItems(db))

	machineID := "items-test-machine-003"
	transactionID1 := 10
	transactionID2 := 20

	// Insert test transactions
	_, err := db.Exec(`
		INSERT INTO transactions (transaction_id, machine_id, hostname, actions, altered, "user", return_code, release_version, command_line, comment, scriptlet_output)
		VALUES ($1, $2, 'hostname-tx1', 'Install', '1', 'root', '0', '8.5', 'dnf install pkg1', '', '')`,
		transactionID1, machineID)
	if err != nil {
		t.Fatalf("Failed to insert test transaction 1: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO transactions (transaction_id, machine_id, hostname, actions, altered, "user", return_code, release_version, command_line, comment, scriptlet_output)
		VALUES ($1, $2, 'hostname-tx2', 'Install', '2', 'root', '0', '8.5', 'dnf install pkg2', '', '')`,
		transactionID2, machineID)
	if err != nil {
		t.Fatalf("Failed to insert test transaction 2: %v", err)
	}

	// Insert items for both transactions
	_, err = db.Exec(`
		INSERT INTO transaction_items (transaction_id, machine_id, action, package, version, release, epoch, arch, repo, from_repo)
		VALUES ($1, $2, 'Install', 'package-from-tx1', '1.0.0', '1.el8', '0', 'x86_64', 'baseos', '')`,
		transactionID1, machineID)
	if err != nil {
		t.Fatalf("Failed to insert item for tx1: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO transaction_items (transaction_id, machine_id, action, package, version, release, epoch, arch, repo, from_repo)
		VALUES ($1, $2, 'Install', 'package-from-tx2', '2.0.0', '1.el8', '0', 'x86_64', 'appstream', '')`,
		transactionID2, machineID)
	if err != nil {
		t.Fatalf("Failed to insert item for tx2: %v", err)
	}

	// Request specific transaction
	req, _ := http.NewRequest("GET", "/v1/items?machine_id="+machineID+"&transaction_id=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response models.Transaction
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Hostname != "hostname-tx1" {
		t.Errorf("Expected hostname 'hostname-tx1', got '%s'", response.Hostname)
	}

	if len(response.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(response.Items))
	} else {
		if response.Items[0].Name != "package-from-tx1" {
			t.Errorf("Expected package 'package-from-tx1', got '%s'", response.Items[0].Name)
		}
	}
}
