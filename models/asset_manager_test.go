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

	err = am.UpsertAsset(tx, hostname, machineID, timestamp, sql.NullBool{}, sql.NullString{}, "")
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
	err = am.UpsertAsset(tx, hostname, machineID, timestamp1, sql.NullBool{}, sql.NullString{}, "")
	if err != nil {
		t.Fatalf("First UpsertAsset() failed: %v", err)
	}

	// Second update
	err = am.UpsertAsset(tx, hostname, machineID, timestamp2, sql.NullBool{}, sql.NullString{}, "")
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
	err = am.UpsertAsset(tx, hostname, oldMachineID, timestamp.Add(-2*time.Hour), sql.NullBool{}, sql.NullString{}, "")
	if err != nil {
		t.Fatalf("Failed to insert old asset: %v", err)
	}

	// Insert new asset with same hostname
	err = am.UpsertAsset(tx, hostname, newMachineID, timestamp, sql.NullBool{}, sql.NullString{}, "")
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
	err = am.UpsertAsset(tx, hostname, machineID, timestamp.Add(-3*time.Hour), sql.NullBool{}, sql.NullString{}, "")
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
	err = am.UpsertAsset(tx, hostname, machineID, timestamp, sql.NullBool{}, sql.NullString{}, "")
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
	err = am.UpsertAsset(tx, hostname, machineID, timestamp, sql.NullBool{}, sql.NullString{}, "")
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
	err = am.UpsertAsset(tx, hostname, machineID, timestamp, sql.NullBool{}, sql.NullString{}, "")
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

func TestAssetManager_DeactivateAssetsByMachineID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	am := NewAssetManager(db)
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	machineID := "test-machine-unique-id"
	hostname1 := "test-server-01"
	hostname2 := "test-server-02"
	timestamp := time.Now()

	// 1. Insert first asset (Host1, ID)
	err = am.UpsertAsset(tx, hostname1, machineID, timestamp.Add(-1*time.Hour), sql.NullBool{}, sql.NullString{}, "")
	if err != nil {
		t.Fatalf("Failed to insert asset 1: %v", err)
	}

	// Check active count for machineID
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM assets WHERE machine_id = $1 AND is_active = TRUE", machineID).Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 active asset, got %d", count)
	}

	// 2. Insert second asset with SAME ID but DIFFERENT Hostname (Host2, ID)
	// This should deactivate the previous asset with the same machineID
	err = am.UpsertAsset(tx, hostname2, machineID, timestamp, sql.NullBool{}, sql.NullString{}, "")
	if err != nil {
		t.Fatalf("Failed to insert asset 2: %v", err)
	}

	// Verify Host1 is now INACTIVE
	var isActive bool
	err = tx.QueryRow("SELECT is_active FROM assets WHERE hostname = $1 AND machine_id = $2", hostname1, machineID).Scan(&isActive)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if isActive {
		t.Errorf("Expected asset 1 to be inactive")
	}

	// Verify Host2 is ACTIVE
	err = tx.QueryRow("SELECT is_active FROM assets WHERE hostname = $1 AND machine_id = $2", hostname2, machineID).Scan(&isActive)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if !isActive {
		t.Errorf("Expected asset 2 to be active")
	}

	// 3. Insert third asset with DIFFERENT ID but SAME Hostname (Host2, ID2)
	// This should NOT deactivate the previous asset with the same hostname (Host2, ID)
	machineID2 := "test-machine-unique-id-2"
	err = am.UpsertAsset(tx, hostname2, machineID2, timestamp, sql.NullBool{}, sql.NullString{}, "")
	if err != nil {
		t.Fatalf("Failed to insert asset 3: %v", err)
	}

	// Verify Host2/ID is STILL ACTIVE (Hostname duplication allowed)
	err = tx.QueryRow("SELECT is_active FROM assets WHERE hostname = $1 AND machine_id = $2", hostname2, machineID).Scan(&isActive)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if !isActive {
		t.Errorf("Expected asset 2 (Host2, ID) to remain active")
	}

	// Verify Host2/ID2 is ACTIVE
	err = tx.QueryRow("SELECT is_active FROM assets WHERE hostname = $1 AND machine_id = $2", hostname2, machineID2).Scan(&isActive)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if !isActive {
		t.Errorf("Expected asset 3 (Host2, ID2) to be active")
	}
}
