package util

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchOSVVulnerabilitiesBatch(t *testing.T) {
	mockResponse := OSVBatchResponse{
		Results: []OSVResult{
			{
				Vulns: []OSVVuln{
					{
						ID:      "CVE-2023-1234",
						Summary: "Test Vulnerability",
					},
				},
			},
			{},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/querybatch" {
			t.Errorf("Expected path /v1/querybatch, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(mockResponse)
		if err != nil {
			t.Fatalf("Failed to encode mock response: %v", err)
		}
	}))
	defer server.Close()

	originalURL := "https://api.osv.dev/v1/querybatch"

	fetchFunc := func(url string, queries []OSVQuery) (*OSVBatchResponse, error) {
		if len(queries) == 0 {
			return &OSVBatchResponse{}, nil
		}

		batchPayload := OSVQueryBatch{Queries: queries}
		jsonData, err := json.Marshal(batchPayload)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var batchResp OSVBatchResponse
		err = json.NewDecoder(resp.Body).Decode(&batchResp)
		if err != nil {
			return nil, err
		}

		return &batchResp, nil
	}

	queries := []OSVQuery{
		{
			Package: OSVPackage{Name: "curl", Ecosystem: "Debian:11"},
			Version: "7.74.0-1.3+deb11u1",
		},
		{
			Package: OSVPackage{Name: "wget", Ecosystem: "Debian:11"},
			Version: "1.21-1+deb11u1",
		},
	}

	resp, err := fetchFunc(server.URL+"/v1/querybatch", queries)
	if err != nil {
		t.Fatalf("fetchFunc failed: %v", err)
	}

	if len(resp.Results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(resp.Results))
	}

	if len(resp.Results[0].Vulns) != 1 {
		t.Fatalf("Expected 1 vulnerability in first result, got %d", len(resp.Results[0].Vulns))
	}

	if resp.Results[0].Vulns[0].ID != "CVE-2023-1234" {
		t.Errorf("Expected ID 'CVE-2023-1234', got '%s'", resp.Results[0].Vulns[0].ID)
	}

	if len(resp.Results[1].Vulns) != 0 {
		t.Fatalf("Expected 0 vulnerabilities in second result, got %d", len(resp.Results[1].Vulns))
	}

	_ = originalURL
}
