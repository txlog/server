package tests

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/txlog/server/models"
)

// setupIntegrationTestDB creates a test database for integration tests
func setupIntegrationTestDB(t *testing.T) *sql.DB {
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=txlog_test sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skip("Skipping integration test: PostgreSQL not available")
	}

	if err := db.Ping(); err != nil {
		t.Skip("Skipping integration test: Cannot connect to PostgreSQL")
	}

	return db
}

// cleanupIntegrationTestData removes all test data
func cleanupIntegrationTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM assets WHERE machine_id LIKE 'integration-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup assets: %v", err)
	}
}

// TestFullAssetLifecycle tests the complete lifecycle of an asset
func TestFullAssetLifecycle(t *testing.T) {
	db := setupIntegrationTestDB(t)
	defer db.Close()
	defer cleanupIntegrationTestData(t, db)

	am := models.NewAssetManager(db)

	hostname := "integration-test-web01"
	machineID1 := "integration-test-machine-001"
	machineID2 := "integration-test-machine-002"

	t.Run("Create initial asset", func(t *testing.T) {
		tx, _ := db.Begin()
		defer tx.Rollback()

		timestamp := time.Now()
		err := am.UpsertAsset(tx, hostname, machineID1, timestamp, sql.NullBool{}, sql.NullString{}, "")
		if err != nil {
			t.Fatalf("Failed to create asset: %v", err)
		}

		if err := tx.Commit(); err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}

		// Verify asset is active
		asset, err := am.GetActiveAsset(hostname)
		if err != nil {
			t.Fatalf("Failed to get active asset: %v", err)
		}

		if asset.MachineID != machineID1 {
			t.Errorf("Expected machine_id %s, got %s", machineID1, asset.MachineID)
		}

		if !asset.IsActive {
			t.Errorf("Expected asset to be active")
		}
	})

	t.Run("Replace asset with new machine_id", func(t *testing.T) {
		tx, _ := db.Begin()
		defer tx.Rollback()

		timestamp := time.Now()
		err := am.UpsertAsset(tx, hostname, machineID2, timestamp, sql.NullBool{}, sql.NullString{}, "")
		if err != nil {
			t.Fatalf("Failed to create replacement asset: %v", err)
		}

		if err := tx.Commit(); err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}

		// Verify new asset is active
		asset, err := am.GetActiveAsset(hostname)
		if err != nil {
			t.Fatalf("Failed to get active asset: %v", err)
		}

		if asset.MachineID != machineID2 {
			t.Errorf("Expected new machine_id %s, got %s", machineID2, asset.MachineID)
		}

		// Verify old asset is inactive
		oldAsset, err := am.GetAssetByMachineID(machineID1)
		if err != nil {
			t.Fatalf("Failed to get old asset: %v", err)
		}

		if oldAsset.IsActive {
			t.Errorf("Expected old asset to be inactive")
		}
	})
}
