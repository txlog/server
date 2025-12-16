package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetVersions_ReturnsVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	expectedVersion := "1.0.0"
	router.GET("/v1/version", GetVersions(expectedVersion))

	req, _ := http.NewRequest("GET", "/v1/version", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["version"] != expectedVersion {
		t.Errorf("Expected version %s, got %s", expectedVersion, response["version"])
	}
}

func TestGetVersions_ReturnsCorrectContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/v1/version", GetVersions("1.0.0"))

	req, _ := http.NewRequest("GET", "/v1/version", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	expectedContentType := "application/json; charset=utf-8"
	if contentType != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
	}
}

func TestGetVersions_WithDifferentVersionFormats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name            string
		version         string
		expectedVersion string
	}{
		{"semver", "1.2.3", "1.2.3"},
		{"semver with prerelease", "1.2.3-beta.1", "1.2.3-beta.1"},
		{"semver with metadata", "1.2.3+build.456", "1.2.3+build.456"},
		{"dev version", "0.0-dev", "0.0-dev"},
		{"empty version", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/v1/version", GetVersions(tc.version))

			req, _ := http.NewRequest("GET", "/v1/version", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}

			var response map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response["version"] != tc.expectedVersion {
				t.Errorf("Expected version %s, got %s", tc.expectedVersion, response["version"])
			}
		})
	}
}

func TestGetVersions_ClosureCapture(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Verify that the closure captures the version at creation time
	version := "1.0.0"
	handler := GetVersions(version)
	router.GET("/v1/version", handler)

	// The handler should use the captured version, not any subsequent changes
	// (this is a conceptual test - version is captured by value)

	req, _ := http.NewRequest("GET", "/v1/version", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["version"] != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", response["version"])
	}
}
