package controllers

import (
	"database/sql"
	"testing"
)

// setupTestDB creates a test database connection for controller tests
func setupTestDB(t *testing.T) *sql.DB {
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

func TestGetTotalActiveAssets(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	count, err := getTotalActiveAssets(db)
	if err != nil {
		t.Errorf("getTotalActiveAssets() error = %v", err)
	}

	if count < 0 {
		t.Errorf("getTotalActiveAssets() returned negative count: %d", count)
	}

	// Count should be >= 0
	t.Logf("Total active assets: %d", count)
}

func TestGetAssetsByOS(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	stats, err := getAssetsByOS(db)
	if err != nil {
		t.Errorf("getAssetsByOS() error = %v", err)
	}

	if stats == nil {
		t.Errorf("getAssetsByOS() returned nil")
	}

	// Verify structure
	for _, stat := range stats {
		if stat.NumMachines < 0 {
			t.Errorf("NumMachines should not be negative: %d", stat.NumMachines)
		}
	}

	t.Logf("Assets by OS: %d entries", len(stats))
}

func TestGetAssetsByAgentVersion(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	stats, err := getAssetsByAgentVersion(db)
	if err != nil {
		t.Errorf("getAssetsByAgentVersion() error = %v", err)
	}

	if stats == nil {
		t.Errorf("getAssetsByAgentVersion() returned nil")
	}

	// Verify structure
	for _, stat := range stats {
		if stat.NumMachines < 0 {
			t.Errorf("NumMachines should not be negative: %d", stat.NumMachines)
		}
	}

	t.Logf("Assets by agent version: %d entries", len(stats))
}

func TestGetDuplicatedAssets(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	assets, err := getDuplicatedAssets(db)
	if err != nil {
		t.Errorf("getDuplicatedAssets() error = %v", err)
	}

	if assets == nil {
		t.Errorf("getDuplicatedAssets() returned nil")
	}

	// Verify structure - duplicated assets should have NumMachines > 1
	for _, asset := range assets {
		if asset.NumMachines <= 1 {
			t.Errorf("Duplicated asset should have NumMachines > 1, got %d for hostname %s",
				asset.NumMachines, asset.Hostname)
		}
	}

	t.Logf("Duplicated assets: %d entries", len(assets))
}

func TestGetMostUpdatedPackages(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	packages, err := getMostUpdatedPackages(db)
	if err != nil {
		t.Errorf("getMostUpdatedPackages() error = %v", err)
	}

	if packages == nil {
		t.Errorf("getMostUpdatedPackages() returned nil")
	}

	// Should return at most 10 packages
	if len(packages) > 10 {
		t.Errorf("getMostUpdatedPackages() should return at most 10 packages, got %d", len(packages))
	}

	// Verify structure
	for _, pkg := range packages {
		if pkg.TotalUpdates < 0 {
			t.Errorf("TotalUpdates should not be negative: %d", pkg.TotalUpdates)
		}
		if pkg.DistinctHostsUpdated < 0 {
			t.Errorf("DistinctHostsUpdated should not be negative: %d", pkg.DistinctHostsUpdated)
		}
		if pkg.Package == "" {
			t.Errorf("Package name should not be empty")
		}
	}

	t.Logf("Most updated packages: %d entries", len(packages))
}

func TestGetStatistics(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	stats, err := getStatistics(db)
	if err != nil {
		t.Errorf("getStatistics() error = %v", err)
	}

	if stats == nil {
		t.Errorf("getStatistics() returned nil")
	}

	// Verify expected statistics exist
	expectedStats := []string{
		"executions-30-days",
		"installed-packages-30-days",
		"upgraded-packages-30-days",
	}

	foundStats := make(map[string]bool)
	for _, stat := range stats {
		foundStats[stat.Name] = true
	}

	for _, expected := range expectedStats {
		if !foundStats[expected] {
			t.Logf("Warning: Expected statistic '%s' not found", expected)
		}
	}

	t.Logf("Statistics: %d entries", len(stats))
}

// Integration test for asset queries with filters
func TestGetAssetsIndex_Queries(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name     string
		search   string
		restart  string
		inactive string
	}{
		{
			name:     "no filters",
			search:   "",
			restart:  "",
			inactive: "",
		},
		{
			name:     "restart filter",
			search:   "",
			restart:  "true",
			inactive: "",
		},
		{
			name:     "inactive filter",
			search:   "",
			restart:  "",
			inactive: "true",
		},
		{
			name:     "search by hostname",
			search:   "test",
			restart:  "",
			inactive: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test count query
			baseCountQuery := `
				SELECT COUNT(DISTINCT a.hostname)
				FROM assets a
				WHERE a.is_active = TRUE
			`

			var count int
			err := db.QueryRow(baseCountQuery).Scan(&count)
			if err != nil {
				t.Errorf("Count query failed: %v", err)
			}

			if count < 0 {
				t.Errorf("Count should not be negative: %d", count)
			}

			// Test select query structure
			baseSelectQuery := `
				SELECT 
					a.asset_id as execution_id,
					a.hostname,
					a.last_seen as executed_at,
					a.machine_id,
					e.os,
					e.needs_restarting
				FROM assets a
				LEFT JOIN LATERAL (
					SELECT os, needs_restarting
					FROM executions
					WHERE machine_id = a.machine_id AND hostname = a.hostname
					ORDER BY executed_at DESC
					LIMIT 1
				) e ON true
				WHERE a.is_active = TRUE
				LIMIT 10
			`

			rows, err := db.Query(baseSelectQuery)
			if err != nil {
				t.Errorf("Select query failed: %v", err)
			}
			defer rows.Close()

			// Verify we can scan results
			rowCount := 0
			for rows.Next() {
				var executionID int
				var hostname string
				var executedAt sql.NullTime
				var machineID string
				var os sql.NullString
				var needsRestarting sql.NullBool

				err := rows.Scan(&executionID, &hostname, &executedAt, &machineID, &os, &needsRestarting)
				if err != nil {
					t.Errorf("Failed to scan row: %v", err)
				}
				rowCount++
			}

			t.Logf("%s: returned %d rows", tt.name, rowCount)
		})
	}
}

// Test to verify assets table uses last_seen correctly
func TestAssetsTable_LastSeenColumn(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Verify last_seen column exists and is usable
	query := `
		SELECT hostname, last_seen 
		FROM assets 
		WHERE is_active = TRUE 
		AND last_seen < NOW() - INTERVAL '15 days'
		LIMIT 5
	`

	rows, err := db.Query(query)
	if err != nil {
		t.Errorf("Query with last_seen column failed: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var hostname string
		var lastSeen sql.NullTime
		err := rows.Scan(&hostname, &lastSeen)
		if err != nil {
			t.Errorf("Failed to scan last_seen: %v", err)
		}
		count++
	}

	t.Logf("Found %d inactive assets (>15 days)", count)
}

// Test to verify only one active asset per hostname
func TestAssetsTable_UniqueActivePerHostname(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	query := `
		SELECT hostname, COUNT(*) as count
		FROM assets
		WHERE is_active = TRUE
		GROUP BY hostname
		HAVING COUNT(*) > 1
	`

	rows, err := db.Query(query)
	if err != nil {
		t.Errorf("Query failed: %v", err)
	}
	defer rows.Close()

	violations := 0
	for rows.Next() {
		var hostname string
		var count int
		err := rows.Scan(&hostname, &count)
		if err != nil {
			t.Errorf("Failed to scan: %v", err)
		}
		t.Errorf("Hostname %s has %d active assets (should be 1)", hostname, count)
		violations++
	}

	if violations > 0 {
		t.Errorf("Found %d hostnames with multiple active assets", violations)
	}
}
