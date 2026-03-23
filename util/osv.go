package util

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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

// OSVSeverity represents a CVSS severity entry from the OSV API.
type OSVSeverity struct {
	Type  string `json:"type"`
	Score string `json:"score"` // CVSS vector string, e.g. "CVSS:3.1/AV:N/AC:L/..."
}

// OSVDatabaseSpecific holds distribution-specific fields (e.g. AlmaLinux, Rocky).
type OSVDatabaseSpecific struct {
	Severity string `json:"severity,omitempty"` // e.g. "Important", "Critical"
}

type OSVVuln struct {
	ID               string              `json:"id"`
	Summary          string              `json:"summary,omitempty"`
	Details          string              `json:"details,omitempty"`
	ModifiedAt       time.Time           `json:"modified,omitempty"`
	Published        time.Time           `json:"published,omitempty"`
	Affected         []OSVAffected       `json:"affected,omitempty"`
	Severity         []OSVSeverity       `json:"severity,omitempty"`
	DatabaseSpecific OSVDatabaseSpecific `json:"database_specific,omitempty"`
}

// cvssBaseScoreRe extracts the base score from a CVSS 3.x vector string.
var cvssBaseScoreRe = regexp.MustCompile(`CVSS:3\.\d+/`)

// ExtractSeverityAndScore determines the severity label and numeric CVSS score
// from structured OSV data, falling back to text matching as a last resort.
func (v *OSVVuln) ExtractSeverityAndScore() (severity string, cvssScore float64) {
	// 1. Try to extract CVSS numeric score from severity[] vector
	for _, s := range v.Severity {
		if strings.HasPrefix(s.Score, "CVSS:3") {
			score := parseCVSSScore(s.Score)
			if score > cvssScore {
				cvssScore = score
			}
		}
	}

	// 2. Try database_specific.severity (used by AlmaLinux, Rocky, etc.)
	if v.DatabaseSpecific.Severity != "" {
		dbSev := strings.ToUpper(v.DatabaseSpecific.Severity)
		switch {
		case dbSev == "CRITICAL":
			severity = "CRITICAL"
		case dbSev == "IMPORTANT" || dbSev == "HIGH":
			severity = "HIGH"
		case dbSev == "MODERATE" || dbSev == "MEDIUM":
			severity = "MEDIUM"
		case dbSev == "LOW":
			severity = "LOW"
		}
	}

	// 3. Infer severity from CVSS score if not set by database_specific
	if severity == "" && cvssScore > 0 {
		switch {
		case cvssScore >= 9.0:
			severity = "CRITICAL"
		case cvssScore >= 7.0:
			severity = "HIGH"
		case cvssScore >= 4.0:
			severity = "MEDIUM"
		default:
			severity = "LOW"
		}
	}

	// 4. Fallback: infer from summary/details text
	if severity == "" {
		summaryUpper := strings.ToUpper(v.Summary)
		detailsUpper := strings.ToUpper(v.Details)

		switch {
		case strings.Contains(summaryUpper, "CRITICAL") || strings.Contains(detailsUpper, "CRITICAL"):
			severity = "CRITICAL"
			if cvssScore == 0 {
				cvssScore = 9.5
			}
		case strings.Contains(summaryUpper, "IMPORTANT") || strings.Contains(summaryUpper, "HIGH") || strings.Contains(detailsUpper, "HIGH"):
			severity = "HIGH"
			if cvssScore == 0 {
				cvssScore = 8.0
			}
		case strings.Contains(summaryUpper, "MODERATE") || strings.Contains(summaryUpper, "MEDIUM") || strings.Contains(detailsUpper, "MEDIUM"):
			severity = "MEDIUM"
			if cvssScore == 0 {
				cvssScore = 5.5
			}
		case strings.Contains(summaryUpper, "LOW") || strings.Contains(detailsUpper, "LOW"):
			severity = "LOW"
			if cvssScore == 0 {
				cvssScore = 3.0
			}
		default:
			severity = "UNKNOWN"
		}
	}

	return severity, cvssScore
}

// parseCVSSScore extracts the base score from a CVSS 3.x vector string.
// It computes a rough estimate based on the vector components.
// For a more accurate score, a dedicated CVSS library should be used.
func parseCVSSScore(vector string) float64 {
	// Quick extraction: look for known base metric values and compute approximate score
	// The vector format is: CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H
	parts := strings.Split(vector, "/")
	if len(parts) < 2 {
		return 0
	}

	metrics := make(map[string]string)
	for _, p := range parts[1:] { // skip "CVSS:3.x"
		kv := strings.SplitN(p, ":", 2)
		if len(kv) == 2 {
			metrics[kv[0]] = kv[1]
		}
	}

	// Approximate CVSS 3.x base score calculation
	av := map[string]float64{"N": 0.85, "A": 0.62, "L": 0.55, "P": 0.20}
	ac := map[string]float64{"L": 0.77, "H": 0.44}
	pr := map[string]float64{"N": 0.85, "L": 0.62, "H": 0.27}
	ui := map[string]float64{"N": 0.85, "R": 0.62}
	cia := map[string]float64{"H": 0.56, "L": 0.22, "N": 0.0}

	avVal := av[metrics["AV"]]
	acVal := ac[metrics["AC"]]
	prVal := pr[metrics["PR"]]
	uiVal := ui[metrics["UI"]]
	cVal := cia[metrics["C"]]
	iVal := cia[metrics["I"]]
	aVal := cia[metrics["A"]]

	// ISS (Impact Sub-Score)
	iss := 1 - ((1 - cVal) * (1 - iVal) * (1 - aVal))
	if iss <= 0 {
		return 0
	}

	// Exploitability
	exploitability := 8.22 * avVal * acVal * prVal * uiVal

	// Impact depends on Scope
	var impact float64
	if metrics["S"] == "C" {
		impact = 7.52*(iss-0.029) - 3.25*math.Pow(iss-0.02, 15)
	} else {
		impact = 6.42 * iss
	}

	if impact <= 0 {
		return 0
	}

	var score float64
	if metrics["S"] == "C" {
		score = math.Min(1.08*(impact+exploitability), 10.0)
	} else {
		score = math.Min(impact+exploitability, 10.0)
	}

	// Round up to nearest 0.1
	score = math.Ceil(score*10) / 10

	// Clamp
	if score > 10.0 {
		score = 10.0
	}

	return score
}

// ParseCVSSScoreFromString tries to parse a CVSS score from a full CVSS vector or a plain number.
func ParseCVSSScoreFromString(s string) float64 {
	if strings.HasPrefix(s, "CVSS:3") {
		return parseCVSSScore(s)
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
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
