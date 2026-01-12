package models

import "time"

// ===== Package Comparison Models =====

// AssetPackageSet represents the set of packages installed on a specific asset
type AssetPackageSet struct {
	MachineID string            `json:"machine_id"`
	Hostname  string            `json:"hostname"`
	Packages  map[string]string `json:"packages"` // package_name -> version-release
}

// PackageVersionDiff represents a version difference for a package across assets
type PackageVersionDiff struct {
	MachineID string `json:"machine_id"`
	Hostname  string `json:"hostname"`
	Version   string `json:"version"`
	Release   string `json:"release"`
}

// PackageComparisonResult represents the result of comparing packages across multiple assets
type PackageComparisonResult struct {
	Assets    []AssetPackageSet               `json:"assets"`
	OnlyIn    map[string][]string             `json:"only_in"`   // machine_id -> []package_names unique to that asset
	Different map[string][]PackageVersionDiff `json:"different"` // package_name -> version differences across assets
	Common    []string                        `json:"common"`    // packages present in all assets with same version
}

// ===== Package Freshness Models =====
// Package Freshness measures how current the deployed versions are based on when they were first seen

// PackageFreshnessInfo represents freshness information for a single package
type PackageFreshnessInfo struct {
	Package     string    `json:"package"`
	Version     string    `json:"version"`
	Release     string    `json:"release"`
	FirstSeenAt time.Time `json:"first_seen_at"`
	AgeInDays   int       `json:"age_in_days"`
	AssetCount  int       `json:"asset_count"`
}

// PackageFreshnessReport represents a freshness analysis report
type PackageFreshnessReport struct {
	MachineID     string                 `json:"machine_id,omitempty"`
	Hostname      string                 `json:"hostname,omitempty"`
	AverageAge    float64                `json:"average_age_days"`
	OldestPackage *PackageFreshnessInfo  `json:"oldest_package,omitempty"`
	NewestPackage *PackageFreshnessInfo  `json:"newest_package,omitempty"`
	Packages      []PackageFreshnessInfo `json:"packages,omitempty"`
}

// ===== Package Adoption Models =====
// Package Adoption measures how widely packages have been adopted across assets

// PackageAdoptionInfo represents adoption information for a single package
type PackageAdoptionInfo struct {
	Package        string  `json:"package"`
	ActiveAssets   int     `json:"active_assets"`
	TotalAssets    int     `json:"total_assets"`
	Percentage     float64 `json:"percentage"`
	UpdateCount30d int     `json:"update_count_30d"` // updates in last 30 days
}

// PackageAdoptionReport represents the adoption report
type PackageAdoptionReport struct {
	TotalActiveAssets int                   `json:"total_active_assets"`
	Packages          []PackageAdoptionInfo `json:"packages"`
}

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
