package util

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type OSVQueryBatch struct {
	Queries []OSVQuery `json:"queries"`
}

type OSVQuery struct {
	Package OSVPackage `json:"package"`
	Version string     `json:"version"`
}

type OSVPackage struct {
	Name      string `json:"name"`
	Ecosystem string `json:"ecosystem,omitempty"`
}

type OSVBatchResponse struct {
	Results []OSVResult `json:"results"`
}

type OSVResult struct {
	Vulns []OSVVuln `json:"vulns,omitempty"`
}

type OSVAffected struct {
	Package OSVPackage `json:"package,omitempty"`
}

type OSVVuln struct {
	ID         string        `json:"id"`
	Summary    string        `json:"summary,omitempty"`
	Details    string        `json:"details,omitempty"`
	ModifiedAt time.Time     `json:"modified,omitempty"`
	Published  time.Time     `json:"published,omitempty"`
	Affected   []OSVAffected `json:"affected,omitempty"`
}

func FetchOSVVulnerabilitiesBatch(queries []OSVQuery) (*OSVBatchResponse, error) {
	if len(queries) == 0 {
		return &OSVBatchResponse{}, nil
	}

	batchPayload := OSVQueryBatch{Queries: queries}
	jsonData, err := json.Marshal(batchPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.osv.dev/v1/querybatch", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
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

func FetchOSVVulnerabilityDetails(id string) (*OSVVuln, error) {
	url := "https://api.osv.dev/v1/vulns/" + id
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil // Might be unfound or error
	}

	var vuln OSVVuln
	err = json.NewDecoder(resp.Body).Decode(&vuln)
	if err != nil {
		return nil, err
	}

	return &vuln, nil
}
