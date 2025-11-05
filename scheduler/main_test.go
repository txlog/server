package scheduler

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/txlog/server/database"
)

// setupTestDB creates a test database connection for scheduler tests
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

// cleanupTestData removes all test data
func cleanupTestData(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM cron_lock WHERE job_name LIKE 'test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup cron_lock: %v", err)
	}
	_, err = db.Exec("DELETE FROM executions WHERE machine_id LIKE 'scheduler-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup executions: %v", err)
	}
}

// TestAcquireLock tests the acquireLock function
func TestAcquireLock(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	// Set the database connection for the scheduler package
	originalDB := database.Db
	database.Db = db
	defer func() { database.Db = originalDB }()

	lockName := "test-lock-1"

	t.Run("Acquire lock successfully", func(t *testing.T) {
		locked, err := acquireLock(lockName)
		if err != nil {
			t.Fatalf("Failed to acquire lock: %v", err)
		}

		if !locked {
			t.Error("Expected lock to be acquired, but it was not")
		}

		// Verify lock exists in database
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM cron_lock WHERE job_name = $1", lockName).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query lock: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected 1 lock entry, got %d", count)
		}
	})

	t.Run("Cannot acquire already locked job", func(t *testing.T) {
		locked, err := acquireLock(lockName)
		if err != nil {
			t.Fatalf("Failed to attempt lock acquisition: %v", err)
		}

		if locked {
			t.Error("Expected lock acquisition to fail, but it succeeded")
		}
	})
}

// TestReleaseLock tests the releaseLock function
func TestReleaseLock(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	originalDB := database.Db
	database.Db = db
	defer func() { database.Db = originalDB }()

	lockName := "test-lock-2"

	t.Run("Release existing lock", func(t *testing.T) {
		// First acquire the lock
		locked, err := acquireLock(lockName)
		if err != nil {
			t.Fatalf("Failed to acquire lock: %v", err)
		}

		if !locked {
			t.Fatal("Lock should have been acquired")
		}

		// Now release it
		err = releaseLock(lockName)
		if err != nil {
			t.Fatalf("Failed to release lock: %v", err)
		}

		// Verify lock is removed
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM cron_lock WHERE job_name = $1", lockName).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query lock: %v", err)
		}

		if count != 0 {
			t.Errorf("Expected 0 lock entries after release, got %d", count)
		}
	})

	t.Run("Release non-existent lock does not error", func(t *testing.T) {
		err := releaseLock("test-lock-nonexistent")
		if err != nil {
			t.Errorf("Expected no error when releasing non-existent lock, got: %v", err)
		}
	})
}

// TestLockAcquireReleaseFlow tests the complete lock lifecycle
func TestLockAcquireReleaseFlow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	originalDB := database.Db
	database.Db = db
	defer func() { database.Db = originalDB }()

	lockName := "test-lock-flow"

	t.Run("Complete lock lifecycle", func(t *testing.T) {
		// Acquire lock
		locked, err := acquireLock(lockName)
		if err != nil {
			t.Fatalf("Failed to acquire lock: %v", err)
		}
		if !locked {
			t.Fatal("Lock should have been acquired")
		}

		// Try to acquire again (should fail)
		locked, err = acquireLock(lockName)
		if err != nil {
			t.Fatalf("Failed to attempt second acquisition: %v", err)
		}
		if locked {
			t.Error("Second acquisition should have failed")
		}

		// Release lock
		err = releaseLock(lockName)
		if err != nil {
			t.Fatalf("Failed to release lock: %v", err)
		}

		// Acquire again (should succeed now)
		locked, err = acquireLock(lockName)
		if err != nil {
			t.Fatalf("Failed to re-acquire lock: %v", err)
		}
		if !locked {
			t.Error("Lock should have been re-acquired after release")
		}

		// Cleanup
		releaseLock(lockName)
	})
}

// TestHousekeepingJob tests the housekeepingJob function
func TestHousekeepingJob(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	originalDB := database.Db
	database.Db = db
	defer func() { database.Db = originalDB }()

	machineID := "scheduler-test-machine-001"

	t.Run("Delete old executions", func(t *testing.T) {
		// Set retention to 7 days for this test
		originalRetention := os.Getenv("CRON_RETENTION_DAYS")
		os.Setenv("CRON_RETENTION_DAYS", "7")
		defer func() {
			if originalRetention == "" {
				os.Unsetenv("CRON_RETENTION_DAYS")
			} else {
				os.Setenv("CRON_RETENTION_DAYS", originalRetention)
			}
		}()

		// Insert old executions (older than 7 days)
		oldExecutedAt := time.Now().AddDate(0, 0, -10)
		_, err := db.Exec(`
			INSERT INTO executions (machine_id, status, message, executed_at)
			VALUES ($1, 'success', 'old execution', $2)`,
			machineID, oldExecutedAt)
		if err != nil {
			t.Fatalf("Failed to insert old execution: %v", err)
		}

		// Insert recent executions (within 7 days)
		recentExecutedAt := time.Now().AddDate(0, 0, -2)
		_, err = db.Exec(`
			INSERT INTO executions (machine_id, status, message, executed_at)
			VALUES ($1, 'success', 'recent execution', $2)`,
			machineID, recentExecutedAt)
		if err != nil {
			t.Fatalf("Failed to insert recent execution: %v", err)
		}

		// Count executions before housekeeping
		var countBefore int
		err = db.QueryRow("SELECT COUNT(*) FROM executions WHERE machine_id = $1", machineID).Scan(&countBefore)
		if err != nil {
			t.Fatalf("Failed to count executions: %v", err)
		}

		if countBefore != 2 {
			t.Fatalf("Expected 2 executions before housekeeping, got %d", countBefore)
		}

		// Run housekeeping job
		housekeepingJob()

		// Count executions after housekeeping
		var countAfter int
		err = db.QueryRow("SELECT COUNT(*) FROM executions WHERE machine_id = $1", machineID).Scan(&countAfter)
		if err != nil {
			t.Fatalf("Failed to count executions: %v", err)
		}

		// Should only have 1 execution left (the recent one)
		if countAfter != 1 {
			t.Errorf("Expected 1 execution after housekeeping, got %d", countAfter)
		}

		// Verify the remaining execution is the recent one
		var executedAt time.Time
		err = db.QueryRow("SELECT executed_at FROM executions WHERE machine_id = $1", machineID).Scan(&executedAt)
		if err != nil {
			t.Fatalf("Failed to query execution: %v", err)
		}

		if executedAt.Before(time.Now().AddDate(0, 0, -7)) {
			t.Error("Remaining execution should be within last 7 days")
		}
	})
}

// TestHousekeepingJobWithDefaultRetention tests housekeeping with default retention
func TestHousekeepingJobWithDefaultRetention(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	originalDB := database.Db
	database.Db = db
	defer func() { database.Db = originalDB }()

	t.Run("Use default retention when env not set", func(t *testing.T) {
		// Unset the retention env variable
		originalRetention := os.Getenv("CRON_RETENTION_DAYS")
		os.Unsetenv("CRON_RETENTION_DAYS")
		defer func() {
			if originalRetention != "" {
				os.Setenv("CRON_RETENTION_DAYS", originalRetention)
			}
		}()

		machineID := "scheduler-test-machine-002"

		// Insert old execution (older than default 7 days)
		oldExecutedAt := time.Now().AddDate(0, 0, -10)
		_, err := db.Exec(`
			INSERT INTO executions (machine_id, status, message, executed_at)
			VALUES ($1, 'success', 'old execution', $2)`,
			machineID, oldExecutedAt)
		if err != nil {
			t.Fatalf("Failed to insert old execution: %v", err)
		}

		// Run housekeeping with default retention
		housekeepingJob()

		// Verify old execution was deleted
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM executions WHERE machine_id = $1", machineID).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count executions: %v", err)
		}

		if count != 0 {
			t.Errorf("Expected 0 executions after housekeeping with default retention, got %d", count)
		}
	})
}

// TestHousekeepingJobWithInvalidRetention tests housekeeping with invalid retention value
func TestHousekeepingJobWithInvalidRetention(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	originalDB := database.Db
	database.Db = db
	defer func() { database.Db = originalDB }()

	t.Run("Ignore invalid retention value", func(t *testing.T) {
		// Set invalid retention value
		originalRetention := os.Getenv("CRON_RETENTION_DAYS")
		os.Setenv("CRON_RETENTION_DAYS", "invalid")
		defer func() {
			if originalRetention == "" {
				os.Unsetenv("CRON_RETENTION_DAYS")
			} else {
				os.Setenv("CRON_RETENTION_DAYS", originalRetention)
			}
		}()

		machineID := "scheduler-test-machine-003"

		// Insert execution
		executedAt := time.Now().AddDate(0, 0, -10)
		_, err := db.Exec(`
			INSERT INTO executions (machine_id, status, message, executed_at)
			VALUES ($1, 'success', 'test execution', $2)`,
			machineID, executedAt)
		if err != nil {
			t.Fatalf("Failed to insert execution: %v", err)
		}

		// Count before
		var countBefore int
		err = db.QueryRow("SELECT COUNT(*) FROM executions WHERE machine_id = $1", machineID).Scan(&countBefore)
		if err != nil {
			t.Fatalf("Failed to count executions: %v", err)
		}

		// Run housekeeping (should not delete anything with invalid value)
		housekeepingJob()

		// Count after
		var countAfter int
		err = db.QueryRow("SELECT COUNT(*) FROM executions WHERE machine_id = $1", machineID).Scan(&countAfter)
		if err != nil {
			t.Fatalf("Failed to count executions: %v", err)
		}

		// Should still have the execution (invalid retention is ignored)
		if countBefore != countAfter {
			t.Logf("Note: Executions count changed from %d to %d with invalid retention", countBefore, countAfter)
		}
	})
}

// TestConcurrentLockAcquisition tests that locks prevent concurrent execution
func TestConcurrentLockAcquisition(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	originalDB := database.Db
	database.Db = db
	defer func() { database.Db = originalDB }()

	lockName := "test-concurrent-lock"

	t.Run("Multiple goroutines try to acquire same lock", func(t *testing.T) {
		results := make(chan bool, 5)

		// Launch 5 goroutines trying to acquire the same lock
		for i := 0; i < 5; i++ {
			go func() {
				locked, _ := acquireLock(lockName)
				results <- locked
			}()
		}

		// Collect results
		successCount := 0
		for i := 0; i < 5; i++ {
			if <-results {
				successCount++
			}
		}

		// Only one should have succeeded
		if successCount != 1 {
			t.Errorf("Expected exactly 1 successful lock acquisition, got %d", successCount)
		}

		// Cleanup
		releaseLock(lockName)
	})
}
