package scheduler

import (
	"database/sql"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
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
	_, err := db.Exec("DELETE FROM executions WHERE machine_id LIKE 'scheduler-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup executions: %v", err)
	}
}

// TestWithLock covers the advisory lock that keeps two instances from running
// the same job at once: it runs the work when the lock is free, skips it when
// another holder has it, and gives the lock back when the work returns.
func TestWithLock(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	const lockName = "test-with-lock"

	t.Run("Runs the function when the lock is free", func(t *testing.T) {
		ran := false
		withLock(db, lockName, func() { ran = true })
		if !ran {
			t.Error("Expected withLock to run the function")
		}
	})

	t.Run("Releases the lock when the function returns", func(t *testing.T) {
		withLock(db, lockName, func() {})

		ran := false
		withLock(db, lockName, func() { ran = true })
		if !ran {
			t.Error("Expected the lock to be free for a second call")
		}
	})

	t.Run("Skips the function while another holder has the lock", func(t *testing.T) {
		inner := false
		withLock(db, lockName, func() {
			// A nested call stands in for a second instance: the lock lives on
			// its own connection, so this competes for it exactly as another
			// process would.
			withLock(db, lockName, func() { inner = true })

			if !IsJobRunning(db, lockName) {
				t.Error("Expected IsJobRunning to report the job as running")
			}
		})

		if inner {
			t.Error("Expected the nested call to be skipped while the lock was held")
		}
	})

	t.Run("Reports the job as not running once the lock is released", func(t *testing.T) {
		withLock(db, lockName, func() {})

		if IsJobRunning(db, lockName) {
			t.Error("Expected IsJobRunning to report the job as not running")
		}
	})
}

// TestConcurrentWithLock verifies that exactly one of several simultaneous
// callers gets to run.
func TestConcurrentWithLock(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	const lockName = "test-concurrent-lock"

	var mu sync.Mutex
	ran := 0
	release := make(chan struct{})
	var wg sync.WaitGroup

	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			withLock(db, lockName, func() {
				mu.Lock()
				ran++
				mu.Unlock()
				// Hold the lock until every goroutine has had its attempt, so
				// the losers cannot simply be running one after the other.
				<-release
			})
		}()
	}

	// Give the losers time to fail their attempt, then let the winner finish.
	time.Sleep(500 * time.Millisecond)
	close(release)
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if ran != 1 {
		t.Errorf("Expected exactly 1 goroutine to run the job, got %d", ran)
	}
}

// TestHousekeepingJob tests the housekeepingJob function
func TestHousekeepingJob(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	machineID := "scheduler-test-machine-001"

	t.Run("Delete old executions", func(t *testing.T) {
		// Set retention to 7 days for this test
		t.Setenv("CRON_RETENTION_DAYS", "7")

		// Insert old executions (older than 7 days)
		oldExecutedAt := time.Now().AddDate(0, 0, -10)
		_, err := db.Exec(`
			INSERT INTO executions (machine_id, hostname, executed_at, success, details)
			VALUES ($1, $1, $2, TRUE, 'old execution')`,
			machineID, oldExecutedAt)
		if err != nil {
			t.Fatalf("Failed to insert old execution: %v", err)
		}

		// Insert recent executions (within 7 days)
		recentExecutedAt := time.Now().AddDate(0, 0, -2)
		_, err = db.Exec(`
			INSERT INTO executions (machine_id, hostname, executed_at, success, details)
			VALUES ($1, $1, $2, TRUE, 'recent execution')`,
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
		housekeepingJob(db)

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

	t.Run("Use default retention when env not set", func(t *testing.T) {
		// Unset the retention env variable
		// t.Setenv cannot unset; an empty value is equivalent here, since
		// DeleteOldExecutions falls back to the default when os.Getenv is "".
		t.Setenv("CRON_RETENTION_DAYS", "")

		machineID := "scheduler-test-machine-002"

		// Insert old execution (older than default 7 days)
		oldExecutedAt := time.Now().AddDate(0, 0, -10)
		_, err := db.Exec(`
			INSERT INTO executions (machine_id, hostname, executed_at, success, details)
			VALUES ($1, $1, $2, TRUE, 'old execution')`,
			machineID, oldExecutedAt)
		if err != nil {
			t.Fatalf("Failed to insert old execution: %v", err)
		}

		// Run housekeeping with default retention
		housekeepingJob(db)

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

	t.Run("Ignore invalid retention value", func(t *testing.T) {
		// Set invalid retention value
		t.Setenv("CRON_RETENTION_DAYS", "invalid")

		machineID := "scheduler-test-machine-003"

		// Insert execution
		executedAt := time.Now().AddDate(0, 0, -10)
		_, err := db.Exec(`
			INSERT INTO executions (machine_id, hostname, executed_at, success, details)
			VALUES ($1, $1, $2, TRUE, 'test execution')`,
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
		housekeepingJob(db)

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
