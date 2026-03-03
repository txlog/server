package models

import "time"

// ===== Transaction Anomaly Models =====

// AnomalyType represents the type of anomaly detected
type AnomalyType string

const (
	AnomalyHighVolume  AnomalyType = "high_volume"
	AnomalyRapidChange AnomalyType = "rapid_change"
	AnomalyDowngrade   AnomalyType = "downgrade"
)

// AnomalySeverity represents the severity level of an anomaly
type AnomalySeverity string

const (
	SeverityLow    AnomalySeverity = "low"
	SeverityMedium AnomalySeverity = "medium"
	SeverityHigh   AnomalySeverity = "high"
)

// TransactionAnomaly represents a detected anomaly in transaction patterns
type TransactionAnomaly struct {
	Type        AnomalyType     `json:"type"`
	MachineID   string          `json:"machine_id"`
	Hostname    string          `json:"hostname"`
	DetectedAt  time.Time       `json:"detected_at"`
	Description string          `json:"description"`
	Severity    AnomalySeverity `json:"severity"`
	Details     any             `json:"details,omitempty"`
}

// HighVolumeDetails provides details for high volume anomalies
type HighVolumeDetails struct {
	TransactionID   int    `json:"transaction_id"`
	PackageCount    int    `json:"package_count"`
	TransactionTime string `json:"transaction_time"`
}

// RapidChangeDetails provides details for rapid change anomalies
type RapidChangeDetails struct {
	Package     string   `json:"package"`
	ChangeCount int      `json:"change_count"`
	TimeWindow  string   `json:"time_window"`
	Actions     []string `json:"actions"`
}

// DowngradeDetails provides details for downgrade anomalies
type DowngradeDetails struct {
	Package     string `json:"package"`
	FromVersion string `json:"from_version"`
	ToVersion   string `json:"to_version"`
}

// AnomalyReport represents the anomaly detection report
type AnomalyReport struct {
	TimeWindow string               `json:"time_window"`
	Anomalies  []TransactionAnomaly `json:"anomalies"`
	Summary    AnomalySummary       `json:"summary"`
}

// AnomalySummary provides a summary count of anomalies by type and severity
type AnomalySummary struct {
	TotalCount    int            `json:"total_count"`
	ByType        map[string]int `json:"by_type"`
	BySeverity    map[string]int `json:"by_severity"`
	AffectedHosts int            `json:"affected_hosts"`
}
