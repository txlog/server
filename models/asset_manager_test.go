package models

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// setupTestDB creates a test database connection
// This requires a PostgreSQL instance running with the txlog database
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

// cleanupTestData removes test data from assets table
func cleanupTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM assets WHERE hostname LIKE 'test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup test data: %v", err)
	}
}

func TestAssetManager_UpsertAsset_NewAsset(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	am := NewAssetManager(db)
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	hostname := "test-server01"
	machineID := "test-machine-001"
	timestamp := time.Now()

	err = am.UpsertAsset(tx, hostname, machineID, timestamp, sql.NullBool{}, sql.NullString{})
	if err != nil {
		t.Errorf("UpsertAsset() failed for new asset: %v", err)
	}

	// Verify asset was created
	var count int
	err = tx.QueryRow(`
		SELECT COUNT(*) FROM assets 
		WHERE hostname = $1 AND machine_id = $2 AND is_active = TRUE
	`, hostname, machineID).Scan(&count)

	if err != nil {
		t.Fatalf("Failed to query asset: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 active asset, got %d", count)
	}
}

func TestAssetManager_UpsertAsset_UpdateExisting(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	am := NewAssetManager(db)
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	hostname := "test-server02"
	machineID := "test-machine-002"
	timestamp1 := time.Now().Add(-1 * time.Hour)
	timestamp2 := time.Now()

	// First insert
	err = am.UpsertAsset(tx, hostname, machineID, timestamp1, sql.NullBool{}, sql.NullString{})
	if err != nil {
		t.Fatalf("First UpsertAsset() failed: %v", err)
	}

	// Second update
	err = am.UpsertAsset(tx, hostname, machineID, timestamp2, sql.NullBool{}, sql.NullString{})
	if err != nil {
		t.Errorf("Second UpsertAsset() failed: %v", err)
	}

	// Verify last_seen was updated
	var lastSeen time.Time
	err = tx.QueryRow(`
		SELECT last_seen FROM assets 
		WHERE hostname = $1 AND machine_id = $2
	`, hostname, machineID).Scan(&lastSeen)

	if err != nil {
		t.Fatalf("Failed to query last_seen: %v", err)
	}

	// Allow 1 second difference for timestamp comparison
	if lastSeen.Unix() != timestamp2.Unix() {
		t.Errorf("Expected last_seen to be updated to %v, got %v", timestamp2, lastSeen)
	}
}

func TestAssetManager_UpsertAsset_ReplacesOldAsset(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	am := NewAssetManager(db)
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	hostname := "test-server03"
	oldMachineID := "test-machine-003-old"
	newMachineID := "test-machine-003-new"
	timestamp := time.Now()

	// Insert old asset
	err = am.UpsertAsset(tx, hostname, oldMachineID, timestamp.Add(-2*time.Hour), sql.NullBool{}, sql.NullString{})
	if err != nil {
		t.Fatalf("Failed to insert old asset: %v", err)
	}

	// Insert new asset with same hostname
	err = am.UpsertAsset(tx, hostname, newMachineID, timestamp, sql.NullBool{}, sql.NullString{})
	if err != nil {
		t.Errorf("Failed to insert new asset: %v", err)
	}

	// Verify old asset is inactive
	var oldIsActive bool
	err = tx.QueryRow(`
		SELECT is_active FROM assets 
		WHERE hostname = $1 AND machine_id = $2
	`, hostname, oldMachineID).Scan(&oldIsActive)

	if err != nil {
		t.Fatalf("Failed to query old asset: %v", err)
	}

	if oldIsActive {
		t.Errorf("Expected old asset to be inactive, but is_active = true")
	}

	// Verify new asset is active
	var newIsActive bool
	err = tx.QueryRow(`
		SELECT is_active FROM assets 
		WHERE hostname = $1 AND machine_id = $2
	`, hostname, newMachineID).Scan(&newIsActive)

	if err != nil {
		t.Fatalf("Failed to query new asset: %v", err)
	}

	if !newIsActive {
		t.Errorf("Expected new asset to be active, but is_active = false")
	}

	// Verify only one asset is active per hostname
	var activeCount int
	err = tx.QueryRow(`
		SELECT COUNT(*) FROM assets 
		WHERE hostname = $1 AND is_active = TRUE
	`, hostname).Scan(&activeCount)

	if err != nil {
		t.Fatalf("Failed to count active assets: %v", err)
	}

	if activeCount != 1 {
		t.Errorf("Expected 1 active asset per hostname, got %d", activeCount)
	}
}

func TestAssetManager_UpsertAsset_ReactivatesInactive(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	am := NewAssetManager(db)
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	hostname := "test-server04"
	machineID := "test-machine-004"
	timestamp := time.Now()

	// Insert asset
	err = am.UpsertAsset(tx, hostname, machineID, timestamp.Add(-3*time.Hour), sql.NullBool{}, sql.NullString{})
	if err != nil {
		t.Fatalf("Failed to insert asset: %v", err)
	}

	// Manually deactivate it
	_, err = tx.Exec(`
		UPDATE assets SET is_active = FALSE, deactivated_at = NOW()
		WHERE hostname = $1 AND machine_id = $2
	`, hostname, machineID)
	if err != nil {
		t.Fatalf("Failed to deactivate asset: %v", err)
	}

	// Reactivate by upserting again
	err = am.UpsertAsset(tx, hostname, machineID, timestamp, sql.NullBool{}, sql.NullString{})
	if err != nil {
		t.Errorf("Failed to reactivate asset: %v", err)
	}

	// Verify asset is active again
	var isActive bool
	var deactivatedAt sql.NullTime
	err = tx.QueryRow(`
		SELECT is_active, deactivated_at FROM assets 
		WHERE hostname = $1 AND machine_id = $2
	`, hostname, machineID).Scan(&isActive, &deactivatedAt)

	if err != nil {
		t.Fatalf("Failed to query asset: %v", err)
	}

	if !isActive {
		t.Errorf("Expected asset to be reactivated, but is_active = false")
	}

	if deactivatedAt.Valid {
		t.Errorf("Expected deactivated_at to be NULL, got %v", deactivatedAt.Time)
	}
}

func TestAssetManager_GetActiveAsset(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	am := NewAssetManager(db)
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	hostname := "test-server05"
	machineID := "test-machine-005"
	timestamp := time.Now()

	// Insert asset
	err = am.UpsertAsset(tx, hostname, machineID, timestamp, sql.NullBool{}, sql.NullString{})
	if err != nil {
		t.Fatalf("Failed to insert asset: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Get active asset
	asset, err := am.GetActiveAsset(hostname)
	if err != nil {
		t.Errorf("GetActiveAsset() failed: %v", err)
	}

	if asset.Hostname != hostname {
		t.Errorf("Expected hostname %s, got %s", hostname, asset.Hostname)
	}

	if asset.MachineID != machineID {
		t.Errorf("Expected machine_id %s, got %s", machineID, asset.MachineID)
	}

	if !asset.IsActive {
		t.Errorf("Expected asset to be active")
	}
}

func TestAssetManager_GetAssetByMachineID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	am := NewAssetManager(db)
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	hostname := "test-server06"
	machineID := "test-machine-006"
	timestamp := time.Now()

	// Insert asset
	err = am.UpsertAsset(tx, hostname, machineID, timestamp, sql.NullBool{}, sql.NullString{})
	if err != nil {
		t.Fatalf("Failed to insert asset: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Get asset by machine_id
	asset, err := am.GetAssetByMachineID(machineID)
	if err != nil {
		t.Errorf("GetAssetByMachineID() failed: %v", err)
	}

	if asset.Hostname != hostname {
		t.Errorf("Expected hostname %s, got %s", hostname, asset.Hostname)
	}

	if asset.MachineID != machineID {
		t.Errorf("Expected machine_id %s, got %s", machineID, asset.MachineID)
	}
}

func TestAssetManager_DeactivateOldAssets(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	am := NewAssetManager(db)
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	hostname := "test-server07"
	oldMachineID1 := "test-machine-007-old1"
	oldMachineID2 := "test-machine-007-old2"
	newMachineID := "test-machine-007-new"
	timestamp := time.Now()

	// Insert multiple old assets
	err = am.UpsertAsset(tx, hostname, oldMachineID1, timestamp.Add(-3*time.Hour), sql.NullBool{}, sql.NullString{})
	if err != nil {
		t.Fatalf("Failed to insert old asset 1: %v", err)
	}

	err = am.UpsertAsset(tx, hostname, oldMachineID2, timestamp.Add(-2*time.Hour), sql.NullBool{}, sql.NullString{})
	if err != nil {
		t.Fatalf("Failed to insert old asset 2: %v", err)
	}

	// Insert new asset
	err = am.UpsertAsset(tx, hostname, newMachineID, timestamp, sql.NullBool{}, sql.NullString{})
	if err != nil {
		t.Fatalf("Failed to insert new asset: %v", err)
	}

	// Verify all old assets are inactive
	var inactiveCount int
	err = tx.QueryRow(`
		SELECT COUNT(*) FROM assets 
		WHERE hostname = $1 AND is_active = FALSE
	`, hostname).Scan(&inactiveCount)

	if err != nil {
		t.Fatalf("Failed to count inactive assets: %v", err)
	}

	if inactiveCount != 2 {
		t.Errorf("Expected 2 inactive assets, got %d", inactiveCount)
	}

	// Verify only one is active
	var activeCount int
	err = tx.QueryRow(`
		SELECT COUNT(*) FROM assets 
		WHERE hostname = $1 AND is_active = TRUE
	`, hostname).Scan(&activeCount)

	if err != nil {
		t.Fatalf("Failed to count active assets: %v", err)
	}

	if activeCount != 1 {
		t.Errorf("Expected 1 active asset, got %d", activeCount)
	}
}
