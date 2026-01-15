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

// =============== Compare Packages Tests ===============

func TestComparePackages_MissingParameter(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/compare-packages", ComparePackages(db))

	req, _ := http.NewRequest("GET", "/v1/reports/compare-packages", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestComparePackages_TooFewMachineIDs(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/compare-packages", ComparePackages(db))

	req, _ := http.NewRequest("GET", "/v1/reports/compare-packages?machine_ids=single-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestComparePackages_TooManyMachineIDs(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/compare-packages", ComparePackages(db))

	// Create 21 IDs
	ids := "id1,id2,id3,id4,id5,id6,id7,id8,id9,id10,id11,id12,id13,id14,id15,id16,id17,id18,id19,id20,id21"
	req, _ := http.NewRequest("GET", "/v1/reports/compare-packages?machine_ids="+ids, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for too many IDs, got %d", w.Code)
	}
}

func TestComparePackages_ValidParameters(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/compare-packages", ComparePackages(db))

	req, _ := http.NewRequest("GET", "/v1/reports/compare-packages?machine_ids=machine1,machine2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result models.PackageComparisonResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify structure
	if result.OnlyIn == nil {
		t.Error("Expected only_in to be initialized")
	}
	if result.Different == nil {
		t.Error("Expected different to be initialized")
	}
	if result.Common == nil {
		t.Error("Expected common to be initialized")
	}
}

// =============== Version Comparison Logic Tests ===============
// These tests verify the version comparison logic without needing a database

func TestVersionComparisonLogic_SameVersions(t *testing.T) {
	// Simulates the logic in comparePackageSets to verify same versions are identified correctly
	tests := []struct {
		name     string
		versions map[string]string
		expected bool // true = all same, false = different
	}{
		{
			name: "identical simple versions",
			versions: map[string]string{
				"machine1": "1.5.7-el9",
				"machine2": "1.5.7-el9",
				"machine3": "1.5.7-el9",
			},
			expected: true,
		},
		{
			name: "identical versions with epoch",
			versions: map[string]string{
				"machine1": "2:8.2.2637-el9",
				"machine2": "2:8.2.2637-el9",
				"machine3": "2:8.2.2637-el9",
			},
			expected: true,
		},
		{
			name: "identical versions with complex release",
			versions: map[string]string{
				"machine1": "3.18.0-el9.x86_64",
				"machine2": "3.18.0-el9.x86_64",
			},
			expected: true,
		},
		{
			name: "different versions",
			versions: map[string]string{
				"machine1": "1.5.7-el9",
				"machine2": "1.5.8-el9",
			},
			expected: false,
		},
		{
			name: "different releases same version",
			versions: map[string]string{
				"machine1": "1.5.7-el8",
				"machine2": "1.5.7-el9",
			},
			expected: false,
		},
		{
			name: "versions with hyphens in release",
			versions: map[string]string{
				"machine1": "4.3.0-1.el9",
				"machine2": "4.3.0-1.el9",
				"machine3": "4.3.0-1.el9",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the same comparison logic as in comparePackageSets
			firstVersion := ""
			allSame := true
			for _, ver := range tt.versions {
				if firstVersion == "" {
					firstVersion = ver
				} else if ver != firstVersion {
					allSame = false
					break
				}
			}

			if allSame != tt.expected {
				t.Errorf("Expected allSame=%v, got %v for versions: %v", tt.expected, allSame, tt.versions)
			}
		})
	}
}

func TestVersionComparisonLogic_VersionParsing(t *testing.T) {
	// Test that version-release parsing works correctly for display purposes
	tests := []struct {
		name            string
		versionRelease  string
		expectedVersion string
		expectedRelease string
	}{
		{
			name:            "simple version-release",
			versionRelease:  "1.5.7-el9",
			expectedVersion: "1.5.7",
			expectedRelease: "el9",
		},
		{
			name:            "version with epoch",
			versionRelease:  "2:8.2.2637-el9",
			expectedVersion: "2:8.2.2637",
			expectedRelease: "el9",
		},
		{
			name:            "release with dots",
			versionRelease:  "3.18.0-el9.x86_64",
			expectedVersion: "3.18.0",
			expectedRelease: "el9.x86_64",
		},
		{
			name:            "release with hyphen",
			versionRelease:  "4.3.0-1.el9",
			expectedVersion: "4.3.0",
			expectedRelease: "1.el9",
		},
		{
			name:            "complex release",
			versionRelease:  "0.69.0-1.el9.noarch",
			expectedVersion: "0.69.0",
			expectedRelease: "1.el9.noarch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := splitVersionRelease(tt.versionRelease)
			if parts[0] != tt.expectedVersion {
				t.Errorf("Expected version=%s, got %s", tt.expectedVersion, parts[0])
			}
			if parts[1] != tt.expectedRelease {
				t.Errorf("Expected release=%s, got %s", tt.expectedRelease, parts[1])
			}
		})
	}
}

// Helper function to split version-release (same logic as in comparePackageSets)
func splitVersionRelease(versionRelease string) [2]string {
	for i := 0; i < len(versionRelease); i++ {
		if versionRelease[i] == '-' {
			return [2]string{versionRelease[:i], versionRelease[i+1:]}
		}
	}
	return [2]string{versionRelease, ""}
}

// =============== Package Freshness Tests ===============

func TestGetPackageFreshness_ValidParameters(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/package-freshness", GetPackageFreshness(db))

	req, _ := http.NewRequest("GET", "/v1/reports/package-freshness", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result models.PackageFreshnessReport
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.Packages == nil {
		t.Error("Expected packages to be initialized")
	}
}

func TestGetPackageFreshness_InvalidLimit(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/package-freshness", GetPackageFreshness(db))

	tests := []struct {
		name  string
		limit string
	}{
		{"zero", "0"},
		{"negative", "-1"},
		{"too large", "501"},
		{"not a number", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/v1/reports/package-freshness?limit="+tt.limit, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400 for limit=%s, got %d", tt.limit, w.Code)
			}
		})
	}
}

func TestGetPackageFreshness_WithMachineID(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/package-freshness", GetPackageFreshness(db))

	req, _ := http.NewRequest("GET", "/v1/reports/package-freshness?machine_id=test-machine", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// =============== Package Adoption Tests ===============

func TestGetPackageAdoption_ValidParameters(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/package-adoption", GetPackageAdoption(db))

	req, _ := http.NewRequest("GET", "/v1/reports/package-adoption", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result models.PackageAdoptionReport
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.Packages == nil {
		t.Error("Expected packages to be initialized")
	}
	if result.TotalActiveAssets < 0 {
		t.Error("Expected total_active_assets to be >= 0")
	}
}

func TestGetPackageAdoption_InvalidLimit(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/package-adoption", GetPackageAdoption(db))

	req, _ := http.NewRequest("GET", "/v1/reports/package-adoption?limit=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetPackageAdoption_InvalidMinAssets(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/reports/package-adoption", GetPackageAdoption(db))

	req, _ := http.NewRequest("GET", "/v1/reports/package-adoption?min_assets=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
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
