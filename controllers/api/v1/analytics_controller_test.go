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

func setupAnalyticsTestDB(t *testing.T) *sql.DB {
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

// =============== Anomaly Detection Tests ===============

func TestGetAnomalies_ValidParameters(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/anomalies", GetAnomalies(db))

	req, _ := http.NewRequest("GET", "/v1/reports/anomalies", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result models.AnomalyReport
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.Anomalies == nil {
		t.Error("Expected anomalies to be initialized")
	}
	if result.TimeWindow == "" {
		t.Error("Expected time_window to be set")
	}
}

func TestGetAnomalies_InvalidDays(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/anomalies", GetAnomalies(db))

	tests := []struct {
		name string
		days string
	}{
		{"zero", "0"},
		{"negative", "-1"},
		{"too large", "91"},
		{"not a number", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/v1/reports/anomalies?days="+tt.days, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400 for days=%s, got %d", tt.days, w.Code)
			}
		})
	}
}

func TestGetAnomalies_InvalidSeverity(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/anomalies", GetAnomalies(db))

	req, _ := http.NewRequest("GET", "/v1/reports/anomalies?severity=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetAnomalies_ValidSeverityFilters(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/anomalies", GetAnomalies(db))

	severities := []string{"low", "medium", "high"}

	for _, severity := range severities {
		t.Run(severity, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/v1/reports/anomalies?severity="+severity, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for severity=%s, got %d", severity, w.Code)
			}
		})
	}
}

func TestGetAnomalies_ResponseStructure(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/anomalies", GetAnomalies(db))

	req, _ := http.NewRequest("GET", "/v1/reports/anomalies?days=7", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var rawResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &rawResponse); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify required fields exist
	requiredFields := []string{"time_window", "anomalies", "summary"}
	for _, field := range requiredFields {
		if _, exists := rawResponse[field]; !exists {
			t.Errorf("Expected field '%s' to exist in response", field)
		}
	}

	// Verify summary structure
	summary, ok := rawResponse["summary"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected summary to be an object")
	}

	summaryFields := []string{"total_count", "by_type", "by_severity", "affected_hosts"}
	for _, field := range summaryFields {
		if _, exists := summary[field]; !exists {
			t.Errorf("Expected field '%s' to exist in summary", field)
		}
	}
}
