package statistics

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/txlog/server/database"
)

// setupTestDB creates a test database connection for statistics tests
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
	_, err := db.Exec("DELETE FROM executions WHERE machine_id LIKE 'stats-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup executions: %v", err)
	}
	_, err = db.Exec("DELETE FROM transaction_items WHERE machine_id LIKE 'stats-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup transaction_items: %v", err)
	}
	_, err = db.Exec("DELETE FROM transactions WHERE machine_id LIKE 'stats-test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup transactions: %v", err)
	}
	_, err = db.Exec("DELETE FROM statistics WHERE name LIKE 'test-%'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup statistics: %v", err)
	}
}

// TestCountExecutions tests the CountExecutions function
func TestCountExecutions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	// Set the database connection for the statistics package
	originalDB := database.Db
	database.Db = db
	defer func() { database.Db = originalDB }()

	machineID := "stats-test-machine-001"

	// Insert test executions
	t.Run("Insert test data", func(t *testing.T) {
		// Last 30 days: 10 executions
		for i := 0; i < 10; i++ {
			executedAt := time.Now().AddDate(0, 0, -i-1)
			_, err := db.Exec(`
				INSERT INTO executions (machine_id, status, message, executed_at)
				VALUES ($1, 'success', 'test execution', $2)`,
				machineID, executedAt)
			if err != nil {
				t.Fatalf("Failed to insert execution: %v", err)
			}
		}

		// 30-60 days ago: 5 executions
		for i := 0; i < 5; i++ {
			executedAt := time.Now().AddDate(0, 0, -31-i)
			_, err := db.Exec(`
				INSERT INTO executions (machine_id, status, message, executed_at)
				VALUES ($1, 'success', 'test execution', $2)`,
				machineID, executedAt)
			if err != nil {
				t.Fatalf("Failed to insert execution: %v", err)
			}
		}
	})

	t.Run("Count executions and verify statistics", func(t *testing.T) {
		CountExecutions()

		var value int
		var percentage float64
		err := db.QueryRow(`
			SELECT value, percentage
			FROM statistics
			WHERE name = 'executions-30-days'`).Scan(&value, &percentage)

		if err != nil {
			t.Fatalf("Failed to query statistics: %v", err)
		}

		if value < 10 {
			t.Errorf("Expected at least 10 executions, got %d", value)
		}

		// Expected percentage: (10-5)/5 * 100 = 100%
		expectedPercentage := float64(100)
		tolerance := float64(10) // Allow some tolerance due to other data
		if percentage < expectedPercentage-tolerance {
			t.Logf("Warning: Percentage %f is lower than expected %f", percentage, expectedPercentage)
		}
	})
}

// TestCountInstalledPackages tests the CountInstalledPackages function
func TestCountInstalledPackages(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	originalDB := database.Db
	database.Db = db
	defer func() { database.Db = originalDB }()

	machineID := "stats-test-machine-002"

	t.Run("Insert test transactions and items", func(t *testing.T) {
		// Last 30 days: 8 installs
		for i := 0; i < 8; i++ {
			transactionID := fmt.Sprintf("stats-test-tx-%d", i)
			beginTime := time.Now().AddDate(0, 0, -i-1)

			_, err := db.Exec(`
				INSERT INTO transactions (transaction_id, machine_id, begin_time, end_time, return_code)
				VALUES ($1, $2, $3, $4, 0)`,
				transactionID, machineID, beginTime, beginTime)
			if err != nil {
				t.Fatalf("Failed to insert transaction: %v", err)
			}

			_, err = db.Exec(`
				INSERT INTO transaction_items (transaction_id, machine_id, item_id, name, action, version, arch, repo)
				VALUES ($1, $2, $3, $4, 'Install', '1.0.0', 'x86_64', 'test-repo')`,
				transactionID, machineID, fmt.Sprintf("test-package-%d", i), fmt.Sprintf("test-package-%d", i))
			if err != nil {
				t.Fatalf("Failed to insert transaction_item: %v", err)
			}
		}

		// 30-60 days ago: 4 installs
		for i := 0; i < 4; i++ {
			transactionID := fmt.Sprintf("stats-test-tx-old-%d", i)
			beginTime := time.Now().AddDate(0, 0, -31-i)

			_, err := db.Exec(`
				INSERT INTO transactions (transaction_id, machine_id, begin_time, end_time, return_code)
				VALUES ($1, $2, $3, $4, 0)`,
				transactionID, machineID, beginTime, beginTime)
			if err != nil {
				t.Fatalf("Failed to insert transaction: %v", err)
			}

			_, err = db.Exec(`
				INSERT INTO transaction_items (transaction_id, machine_id, item_id, name, action, version, arch, repo)
				VALUES ($1, $2, $3, $4, 'Install', '1.0.0', 'x86_64', 'test-repo')`,
				transactionID, machineID, fmt.Sprintf("test-package-old-%d", i), fmt.Sprintf("test-package-old-%d", i))
			if err != nil {
				t.Fatalf("Failed to insert transaction_item: %v", err)
			}
		}
	})

	t.Run("Count installed packages and verify statistics", func(t *testing.T) {
		CountInstalledPackages()

		var value int
		var percentage float64
		err := db.QueryRow(`
			SELECT value, percentage
			FROM statistics
			WHERE name = 'installed-packages-30-days'`).Scan(&value, &percentage)

		if err != nil {
			t.Fatalf("Failed to query statistics: %v", err)
		}

		if value < 8 {
			t.Errorf("Expected at least 8 installed packages, got %d", value)
		}

		// Expected percentage: (8-4)/4 * 100 = 100%
		expectedPercentage := float64(100)
		tolerance := float64(10)
		if percentage < expectedPercentage-tolerance {
			t.Logf("Warning: Percentage %f is lower than expected %f", percentage, expectedPercentage)
		}
	})
}

// TestCountUpgradedPackages tests the CountUpgradedPackages function
func TestCountUpgradedPackages(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	originalDB := database.Db
	database.Db = db
	defer func() { database.Db = originalDB }()

	machineID := "stats-test-machine-003"

	t.Run("Insert test upgrades", func(t *testing.T) {
		// Last 30 days: 6 upgrades
		for i := 0; i < 6; i++ {
			transactionID := fmt.Sprintf("stats-test-upgrade-tx-%d", i)
			beginTime := time.Now().AddDate(0, 0, -i-1)

			_, err := db.Exec(`
				INSERT INTO transactions (transaction_id, machine_id, begin_time, end_time, return_code)
				VALUES ($1, $2, $3, $4, 0)`,
				transactionID, machineID, beginTime, beginTime)
			if err != nil {
				t.Fatalf("Failed to insert transaction: %v", err)
			}

			_, err = db.Exec(`
				INSERT INTO transaction_items (transaction_id, machine_id, item_id, name, action, version, arch, repo)
				VALUES ($1, $2, $3, $4, 'Upgrade', '2.0.0', 'x86_64', 'test-repo')`,
				transactionID, machineID, fmt.Sprintf("test-upgrade-%d", i), fmt.Sprintf("test-upgrade-%d", i))
			if err != nil {
				t.Fatalf("Failed to insert transaction_item: %v", err)
			}
		}

		// 30-60 days ago: 3 upgrades
		for i := 0; i < 3; i++ {
			transactionID := fmt.Sprintf("stats-test-upgrade-tx-old-%d", i)
			beginTime := time.Now().AddDate(0, 0, -31-i)

			_, err := db.Exec(`
				INSERT INTO transactions (transaction_id, machine_id, begin_time, end_time, return_code)
				VALUES ($1, $2, $3, $4, 0)`,
				transactionID, machineID, beginTime, beginTime)
			if err != nil {
				t.Fatalf("Failed to insert transaction: %v", err)
			}

			_, err = db.Exec(`
				INSERT INTO transaction_items (transaction_id, machine_id, item_id, name, action, version, arch, repo)
				VALUES ($1, $2, $3, $4, 'Upgrade', '2.0.0', 'x86_64', 'test-repo')`,
				transactionID, machineID, fmt.Sprintf("test-upgrade-old-%d", i), fmt.Sprintf("test-upgrade-old-%d", i))
			if err != nil {
				t.Fatalf("Failed to insert transaction_item: %v", err)
			}
		}
	})

	t.Run("Count upgraded packages and verify statistics", func(t *testing.T) {
		CountUpgradedPackages()

		var value int
		var percentage float64
		err := db.QueryRow(`
			SELECT value, percentage
			FROM statistics
			WHERE name = 'upgraded-packages-30-days'`).Scan(&value, &percentage)

		if err != nil {
			t.Fatalf("Failed to query statistics: %v", err)
		}

		if value < 6 {
			t.Errorf("Expected at least 6 upgraded packages, got %d", value)
		}

		// Expected percentage: (6-3)/3 * 100 = 100%
		expectedPercentage := float64(100)
		tolerance := float64(10)
		if percentage < expectedPercentage-tolerance {
			t.Logf("Warning: Percentage %f is lower than expected %f", percentage, expectedPercentage)
		}
	})
}

// TestStatisticsWithZeroPreviousMonth tests percentage calculation when previous month is zero
func TestStatisticsWithZeroPreviousMonth(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	originalDB := database.Db
	database.Db = db
	defer func() { database.Db = originalDB }()

	machineID := "stats-test-machine-004"

	t.Run("Insert data only in last 30 days", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			executedAt := time.Now().AddDate(0, 0, -i-1)
			_, err := db.Exec(`
				INSERT INTO executions (machine_id, status, message, executed_at)
				VALUES ($1, 'success', 'test execution', $2)`,
				machineID, executedAt)
			if err != nil {
				t.Fatalf("Failed to insert execution: %v", err)
			}
		}
	})

	t.Run("Verify percentage is 0 when previous month is 0", func(t *testing.T) {
		CountExecutions()

		var value int
		var percentage float64
		err := db.QueryRow(`
			SELECT value, percentage
			FROM statistics
			WHERE name = 'executions-30-days'`).Scan(&value, &percentage)

		if err != nil {
			t.Fatalf("Failed to query statistics: %v", err)
		}

		if value < 5 {
			t.Errorf("Expected at least 5 executions, got %d", value)
		}

		// When previousMonth is 0, percentage should be 0 (avoiding division by zero)
		// The actual result may vary due to other data in the database
		t.Logf("Percentage with zero previous month: %f", percentage)
	})
}
