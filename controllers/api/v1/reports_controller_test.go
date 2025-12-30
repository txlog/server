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
)

func setupReportsTestDB(t *testing.T) *sql.DB {
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

func cleanupReportsTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM transaction_items WHERE machine_id LIKE 'report-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup transaction_items: %v", err)
	}
	_, err = db.Exec("DELETE FROM transactions WHERE machine_id LIKE 'report-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup transactions: %v", err)
	}
	_, err = db.Exec("DELETE FROM executions WHERE machine_id LIKE 'report-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup executions: %v", err)
	}
	_, err = db.Exec("DELETE FROM assets WHERE machine_id LIKE 'report-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup assets: %v", err)
	}
}

func TestGetMonthlyReport_MissingParameters(t *testing.T) {
	db := setupReportsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/monthly", GetMonthlyReport(db))

	tests := []struct {
		name     string
		url      string
		expected int
	}{
		{"missing both parameters", "/v1/reports/monthly", http.StatusBadRequest},
		{"missing year", "/v1/reports/monthly?month=12", http.StatusBadRequest},
		{"missing month", "/v1/reports/monthly?year=2025", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, w.Code)
			}
		})
	}
}

func TestGetMonthlyReport_InvalidParameters(t *testing.T) {
	db := setupReportsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/monthly", GetMonthlyReport(db))

	tests := []struct {
		name     string
		url      string
		expected int
	}{
		{"invalid month - too low", "/v1/reports/monthly?month=0&year=2025", http.StatusBadRequest},
		{"invalid month - too high", "/v1/reports/monthly?month=13&year=2025", http.StatusBadRequest},
		{"invalid month - not a number", "/v1/reports/monthly?month=abc&year=2025", http.StatusBadRequest},
		{"invalid year - too low", "/v1/reports/monthly?month=12&year=1999", http.StatusBadRequest},
		{"invalid year - too high", "/v1/reports/monthly?month=12&year=2101", http.StatusBadRequest},
		{"invalid year - not a number", "/v1/reports/monthly?month=12&year=abc", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, w.Code)
			}
		})
	}
}

func TestGetMonthlyReport_ValidParameters(t *testing.T) {
	db := setupReportsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/monthly", GetMonthlyReport(db))

	req, _ := http.NewRequest("GET", "/v1/reports/monthly?month=12&year=2025", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response MonthlyReportResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Month != 12 {
		t.Errorf("Expected month 12, got %d", response.Month)
	}

	if response.Year != 2025 {
		t.Errorf("Expected year 2025, got %d", response.Year)
	}

	if response.AssetCount < 0 {
		t.Errorf("Expected asset count >= 0, got %d", response.AssetCount)
	}

	if response.Packages == nil {
		t.Errorf("Expected packages to be an array, got nil")
	}
}

func TestGetMonthlyReport_ResponseStructure(t *testing.T) {
	db := setupReportsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/monthly", GetMonthlyReport(db))

	req, _ := http.NewRequest("GET", "/v1/reports/monthly?month=1&year=2025", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	// Test JSON structure
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &rawResponse); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify required fields exist
	requiredFields := []string{"month", "year", "asset_count", "packages"}
	for _, field := range requiredFields {
		if _, exists := rawResponse[field]; !exists {
			t.Errorf("Expected field '%s' to exist in response", field)
		}
	}
}

func TestGetMonthlyReport_WithTestData(t *testing.T) {
	db := setupReportsTestDB(t)
	defer db.Close()
	defer cleanupReportsTestData(t, db)

	// Insert test data
	machineID := "report-test-machine-001"
	hostname := "report-test-server01"
	now := time.Now()

	// Insert asset
	_, err := db.Exec(`
		INSERT INTO assets (machine_id, hostname, is_active, first_seen, last_seen)
		VALUES ($1, $2, TRUE, $3, $3)
	`, machineID, hostname, now)
	if err != nil {
		t.Fatalf("Failed to insert asset: %v", err)
	}

	// Insert execution with OS
	_, err = db.Exec(`
		INSERT INTO executions (machine_id, hostname, executed_at, success, os, agent_version)
		VALUES ($1, $2, $3, TRUE, 'Red Hat Enterprise Linux 9.3', '1.0.0')
	`, machineID, hostname, now)
	if err != nil {
		t.Fatalf("Failed to insert execution: %v", err)
	}

	// Insert transaction
	transactionID := 1
	_, err = db.Exec(`
		INSERT INTO transactions (transaction_id, machine_id, hostname, begin_time, end_time)
		VALUES ($1, $2, $3, $4, $5)
	`, transactionID, machineID, hostname, now, now)
	if err != nil {
		t.Fatalf("Failed to insert transaction: %v", err)
	}

	// Insert transaction item
	_, err = db.Exec(`
		INSERT INTO transaction_items (transaction_id, machine_id, action, package, version, release, arch)
		VALUES ($1, $2, 'Upgraded', 'test-package', '1.0.0', '1.el9', 'x86_64')
	`, transactionID, machineID)
	if err != nil {
		t.Fatalf("Failed to insert transaction_item: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/monthly", GetMonthlyReport(db))

	// Query for current month/year
	month := int(now.Month())
	year := now.Year()

	req, _ := http.NewRequest("GET", "/v1/reports/monthly?month="+string(rune('0'+month/10))+string(rune('0'+month%10))+"&year="+string(rune('0'+year/1000%10))+string(rune('0'+year/100%10))+string(rune('0'+year/10%10))+string(rune('0'+year%10)), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response MonthlyReportResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.AssetCount < 1 {
		t.Errorf("Expected at least 1 asset, got %d", response.AssetCount)
	}

	// Verify packages contain our test package
	foundPackage := false
	for _, pkg := range response.Packages {
		if pkg.PackageRPM == "test-package-1.0.0-1.el9.x86_64" {
			foundPackage = true
			if pkg.AssetsAffected != 1 {
				t.Errorf("Expected 1 asset affected, got %d", pkg.AssetsAffected)
			}
			if pkg.OSVersion != "Red Hat Enterprise Linux 9.3" {
				t.Errorf("Expected OS 'Red Hat Enterprise Linux 9.3', got '%s'", pkg.OSVersion)
			}
			break
		}
	}

	if !foundPackage {
		t.Errorf("Expected to find test-package in response")
	}
}

func TestGetMonthlyReport_EmptyPeriod(t *testing.T) {
	db := setupReportsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/monthly", GetMonthlyReport(db))

	// Query for a future date where there should be no data
	req, _ := http.NewRequest("GET", "/v1/reports/monthly?month=1&year=2099", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response MonthlyReportResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response.Packages) != 0 {
		t.Errorf("Expected 0 packages for future date, got %d", len(response.Packages))
	}
}

func TestGetMonthlyReport_AllMonths(t *testing.T) {
	db := setupReportsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/monthly", GetMonthlyReport(db))

	// Test all valid months
	for month := 1; month <= 12; month++ {
		t.Run("month_"+string(rune('0'+month/10))+string(rune('0'+month%10)), func(t *testing.T) {
			url := "/v1/reports/monthly?month=" + string(rune('0'+month/10)) + string(rune('0'+month%10)) + "&year=2025"
			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for month %d, got %d", month, w.Code)
			}
		})
	}
}
