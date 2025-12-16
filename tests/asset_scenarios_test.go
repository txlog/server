package tests

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/txlog/server/models"
)

func TestMultipleAssetReplacements(t *testing.T) {
	db := setupIntegrationTestDB(t)
	defer db.Close()
	defer cleanupIntegrationTestData(t, db)

	am := models.NewAssetManager(db)
	hostname := "integration-test-replacements"

	// Create sequence of machines replacing each other
	machines := []string{
		"integration-test-machine-r01",
		"integration-test-machine-r02",
		"integration-test-machine-r03",
	}

	for i, machineID := range machines {
		tx, _ := db.Begin()
		defer tx.Rollback()

		timestamp := time.Now().Add(time.Duration(i) * time.Hour)
		err := am.UpsertAsset(tx, hostname, machineID, timestamp, sql.NullBool{}, sql.NullString{}, "")
		if err != nil {
			t.Fatalf("Failed to create asset %d: %v", i, err)
		}

		if err := tx.Commit(); err != nil {
			t.Fatalf("Failed to commit %d: %v", i, err)
		}
	}

	// Verify only the last one is active
	activeAsset, err := am.GetActiveAsset(hostname)
	if err != nil {
		t.Fatalf("Failed to get active asset: %v", err)
	}

	if activeAsset.MachineID != machines[2] {
		t.Errorf("Expected active machine_id %s, got %s", machines[2], activeAsset.MachineID)
	}

	// Verify all previous ones are inactive
	for i := 0; i < 2; i++ {
		asset, err := am.GetAssetByMachineID(machines[i])
		if err != nil {
			t.Fatalf("Failed to get asset %s: %v", machines[i], err)
		}

		if asset.IsActive {
			t.Errorf("Expected asset %s to be inactive", machines[i])
		}

		if asset.DeactivatedAt == nil {
			t.Errorf("Expected asset %s to have deactivated_at set", machines[i])
		}
	}
}

func TestConcurrentAssetUpdates(t *testing.T) {
	db := setupIntegrationTestDB(t)
	defer db.Close()
	defer cleanupIntegrationTestData(t, db)

	am := models.NewAssetManager(db)
	hostname := "integration-test-concurrent"
	machineID := "integration-test-machine-c01"

	// Simulate concurrent updates (in sequence but testing the logic)
	for i := 0; i < 5; i++ {
		tx, _ := db.Begin()

		timestamp := time.Now().Add(time.Duration(i) * time.Minute)
		err := am.UpsertAsset(tx, hostname, machineID, timestamp, sql.NullBool{}, sql.NullString{}, "")
		if err != nil {
			tx.Rollback()
			t.Fatalf("Failed to upsert asset iteration %d: %v", i, err)
		}

		if err := tx.Commit(); err != nil {
			t.Fatalf("Failed to commit iteration %d: %v", i, err)
		}
	}

	// Verify we still have only one active asset
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM assets 
		WHERE hostname = $1 AND is_active = TRUE
	`, hostname).Scan(&count)

	if err != nil {
		t.Fatalf("Failed to count active assets: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 active asset, got %d", count)
	}
}

func TestAssetReactivation(t *testing.T) {
	db := setupIntegrationTestDB(t)
	defer db.Close()
	defer cleanupIntegrationTestData(t, db)

	am := models.NewAssetManager(db)
	hostname1 := "integration-test-reactivation1"
	hostname2 := "integration-test-reactivation2"
	machineID := "integration-test-machine-reactivate"

	t.Run("Create asset with hostname1", func(t *testing.T) {
		tx, _ := db.Begin()
		defer tx.Rollback()

		err := am.UpsertAsset(tx, hostname1, machineID, time.Now(), sql.NullBool{}, sql.NullString{}, "")
		if err != nil {
			t.Fatalf("Failed to create initial asset: %v", err)
		}

		if err := tx.Commit(); err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}
	})

	t.Run("Move asset to hostname2", func(t *testing.T) {
		tx, _ := db.Begin()
		defer tx.Rollback()

		err := am.UpsertAsset(tx, hostname2, machineID, time.Now(), sql.NullBool{}, sql.NullString{}, "")
		if err != nil {
			t.Fatalf("Failed to move asset: %v", err)
		}

		if err := tx.Commit(); err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}

		// Verify first hostname1 is inactive
		asset1, err := am.GetAssetByMachineID(machineID)
		if err == nil && asset1.Hostname == hostname1 {
			if asset1.IsActive {
				t.Error("Expected asset with hostname1 to be inactive")
			}
		}

		// Verify hostname2 is active
		asset2, err := am.GetActiveAsset(hostname2)
		if err != nil {
			t.Fatalf("Failed to get active asset: %v", err)
		}

		if asset2.MachineID != machineID {
			t.Errorf("Expected machine_id %s, got %s", machineID, asset2.MachineID)
		}
	})

	t.Run("Move asset back to hostname1", func(t *testing.T) {
		tx, _ := db.Begin()
		defer tx.Rollback()

		err := am.UpsertAsset(tx, hostname1, machineID, time.Now(), sql.NullBool{}, sql.NullString{}, "")
		if err != nil {
			t.Fatalf("Failed to reactivate asset: %v", err)
		}

		if err := tx.Commit(); err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}

		// Verify hostname1 is active again
		asset1, err := am.GetActiveAsset(hostname1)
		if err != nil {
			t.Fatalf("Failed to get active asset: %v", err)
		}

		if asset1.MachineID != machineID {
			t.Errorf("Expected machine_id %s, got %s", machineID, asset1.MachineID)
		}

		if !asset1.IsActive {
			t.Error("Expected asset to be active")
		}
	})
}

func TestAssetHistoryPreservation(t *testing.T) {
	db := setupIntegrationTestDB(t)
	defer db.Close()
	defer cleanupIntegrationTestData(t, db)

	am := models.NewAssetManager(db)
	hostname := "integration-test-history"

	machines := []struct {
		id        string
		timestamp time.Time
	}{
		{"integration-test-machine-h01", time.Now().Add(-48 * time.Hour)},
		{"integration-test-machine-h02", time.Now().Add(-24 * time.Hour)},
		{"integration-test-machine-h03", time.Now()},
	}

	// Create all assets
	for _, m := range machines {
		tx, _ := db.Begin()
		err := am.UpsertAsset(tx, hostname, m.id, m.timestamp, sql.NullBool{}, sql.NullString{}, "")
		if err != nil {
			tx.Rollback()
			t.Fatalf("Failed to create asset %s: %v", m.id, err)
		}
		tx.Commit()
	}

	// Verify we have complete history
	var totalCount int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM assets WHERE hostname = $1
	`, hostname).Scan(&totalCount)

	if err != nil {
		t.Fatalf("Failed to count assets: %v", err)
	}

	if totalCount != 3 {
		t.Errorf("Expected 3 assets in history, got %d", totalCount)
	}

	// Verify only one is active
	var activeCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM assets 
		WHERE hostname = $1 AND is_active = TRUE
	`, hostname).Scan(&activeCount)

	if err != nil {
		t.Fatalf("Failed to count active assets: %v", err)
	}

	if activeCount != 1 {
		t.Errorf("Expected 1 active asset, got %d", activeCount)
	}

	// Verify timestamps are preserved
	for _, m := range machines {
		asset, err := am.GetAssetByMachineID(m.id)
		if err != nil {
			t.Fatalf("Failed to get asset %s: %v", m.id, err)
		}

		// Allow 1 second tolerance for timestamp comparison
		timeDiff := asset.FirstSeen.Sub(m.timestamp)
		if timeDiff < -1*time.Second || timeDiff > 1*time.Second {
			t.Errorf("Asset %s: expected first_seen ~%v, got %v", m.id, m.timestamp, asset.FirstSeen)
		}
	}
}

func TestAssetDatabaseConstraints(t *testing.T) {
	db := setupIntegrationTestDB(t)
	defer db.Close()
	defer cleanupIntegrationTestData(t, db)

	t.Run("Cannot have multiple active assets per hostname", func(t *testing.T) {
		hostname := "integration-test-constraint"
		machineID1 := "integration-test-machine-c1"
		machineID2 := "integration-test-machine-c2"

		// Create first asset
		am := models.NewAssetManager(db)
		tx, _ := db.Begin()
		err := am.UpsertAsset(tx, hostname, machineID1, time.Now(), sql.NullBool{}, sql.NullString{}, "")
		if err != nil {
			tx.Rollback()
			t.Fatalf("Failed to create first asset: %v", err)
		}
		tx.Commit()

		// Create second asset (should deactivate first)
		tx, _ = db.Begin()
		err = am.UpsertAsset(tx, hostname, machineID2, time.Now(), sql.NullBool{}, sql.NullString{}, "")
		if err != nil {
			tx.Rollback()
			t.Fatalf("Failed to create second asset: %v", err)
		}
		tx.Commit()

		// Verify only one is active
		var activeCount int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM assets 
			WHERE hostname = $1 AND is_active = TRUE
		`, hostname).Scan(&activeCount)

		if err != nil {
			t.Fatalf("Failed to count active assets: %v", err)
		}

		if activeCount != 1 {
			t.Errorf("Expected exactly 1 active asset, got %d", activeCount)
		}
	})
}

func TestGetActiveAsset_NoAsset(t *testing.T) {
	db := setupIntegrationTestDB(t)
	defer db.Close()

	am := models.NewAssetManager(db)
	_, err := am.GetActiveAsset("nonexistent-hostname")

	if err == nil {
		t.Error("Expected error when getting non-existent asset")
	}

	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestGetAssetByMachineID_NoAsset(t *testing.T) {
	db := setupIntegrationTestDB(t)
	defer db.Close()

	am := models.NewAssetManager(db)
	_, err := am.GetAssetByMachineID("nonexistent-machine-id")

	if err == nil {
		t.Error("Expected error when getting non-existent asset")
	}

	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestAssetTimestampEdgeCases(t *testing.T) {
	db := setupIntegrationTestDB(t)
	defer db.Close()
	defer cleanupIntegrationTestData(t, db)

	am := models.NewAssetManager(db)

	t.Run("Update with older timestamp", func(t *testing.T) {
		hostname := "integration-test-timestamp"
		machineID := "integration-test-machine-ts"

		// Create with current time
		tx, _ := db.Begin()
		now := time.Now()
		err := am.UpsertAsset(tx, hostname, machineID, now, sql.NullBool{}, sql.NullString{}, "")
		if err != nil {
			tx.Rollback()
			t.Fatalf("Failed to create asset: %v", err)
		}
		tx.Commit()

		// Update with older timestamp (should still update last_seen)
		tx, _ = db.Begin()
		older := now.Add(-1 * time.Hour)
		err = am.UpsertAsset(tx, hostname, machineID, older, sql.NullBool{}, sql.NullString{}, "")
		if err != nil {
			tx.Rollback()
			t.Fatalf("Failed to update asset: %v", err)
		}
		tx.Commit()

		// Verify last_seen was updated (to the older time)
		asset, err := am.GetActiveAsset(hostname)
		if err != nil {
			t.Fatalf("Failed to get asset: %v", err)
		}

		// The last_seen should be the most recent upsert (older timestamp)
		timeDiff := asset.LastSeen.Sub(older)
		if timeDiff < -1*time.Second || timeDiff > 1*time.Second {
			t.Errorf("Expected last_seen to be updated to %v, got %v", older, asset.LastSeen)
		}
	})
}
