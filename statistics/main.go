package statistics

import (
	"github.com/txlog/server/database"
	"github.com/txlog/server/logger"
)

// CountServers calculates and stores statistics about server usage over time.
// It queries the database to count distinct servers (machine_ids) that executed
// commands in the last 30 days compared to the previous 30 day period (30-60 days ago).
// The function calculates the percentage change between these two periods.
//
// The results are stored in the statistics table with the name "servers-30-days",
// including the current month's count and the percentage change.
//
// If an error occurs during database operations, it logs the error and returns
// without updating statistics.
func CountServers() {
	var thisMonth, previousMonth int
	err := database.Db.QueryRow(`
	        WITH last30days AS (
	          SELECT DISTINCT machine_id
	          FROM executions
	          WHERE executed_at >= NOW() - INTERVAL '30 days'
	        ),

	        last60days AS (
	          SELECT DISTINCT machine_id
	          FROM executions
	          WHERE executed_at >= NOW() - INTERVAL '60 days' AND executed_at < NOW() - INTERVAL '30 days'
	        )

	        SELECT
	          (SELECT COUNT(*) FROM last30days) AS this_month,
	          (SELECT COUNT(*) FROM last60days) AS previous_month;
	      `).Scan(&thisMonth, &previousMonth)

	if err != nil {
		logger.Error("Error querying statistics: " + err.Error())
		return
	}

	var percentage float64
	if previousMonth > 0 {
		percentage = float64(thisMonth-previousMonth) / float64(previousMonth) * 100
	}

	_, err = database.Db.Exec(`
	        INSERT INTO statistics (name, value, percentage, updated_at)
	        VALUES ($1, $2, $3, NOW())
	        ON CONFLICT (name) DO UPDATE
	        SET value = $2, percentage = $3, updated_at = NOW()`,
		"servers-30-days", thisMonth, percentage)

	if err != nil {
		logger.Error("Error inserting statistics: " + err.Error())
		return
	}
}

// CountExecutions calculates execution statistics over a 60-day period and stores the results.
// It queries the database to count executions in two periods:
// - Current period: last 30 days
// - Previous period: 30-60 days ago
//
// It then calculates the percentage change between these periods and stores:
// - The count for the current 30-day period
// - The percentage change compared to the previous period
//
// The results are stored in the statistics table under the name "executions-30-days".
// If a record already exists, it updates the existing values.
//
// Any database errors encountered during the process are logged using the logger.Error function.
func CountExecutions() {
	var thisMonth, previousMonth int
	err := database.Db.QueryRow(`
	        WITH last30days AS (
	          SELECT id
	          FROM executions
	          WHERE executed_at >= NOW() - INTERVAL '30 days'
	        ),

	        last60days AS (
	          SELECT id
	          FROM executions
	          WHERE executed_at >= NOW() - INTERVAL '60 days' AND executed_at < NOW() - INTERVAL '30 days'
	        )

	        SELECT
	          (SELECT COUNT(*) FROM last30days) AS this_month,
	          (SELECT COUNT(*) FROM last60days) AS previous_month;
	      `).Scan(&thisMonth, &previousMonth)

	if err != nil {
		logger.Error("Error querying statistics: " + err.Error())
		return
	}

	var percentage float64
	if previousMonth > 0 {
		percentage = float64(thisMonth-previousMonth) / float64(previousMonth) * 100
	}

	_, err = database.Db.Exec(`
	        INSERT INTO statistics (name, value, percentage, updated_at)
	        VALUES ($1, $2, $3, NOW())
	        ON CONFLICT (name) DO UPDATE
	        SET value = $2, percentage = $3, updated_at = NOW()`,
		"executions-30-days", thisMonth, percentage)

	if err != nil {
		logger.Error("Error inserting statistics: " + err.Error())
		return
	}
}

// CountInstalledPackages queries and stores statistics about package installations.
// It calculates the number of package upgrades in the last 30 days (thisMonth)
// and the 30 days before that (previousMonth). It then computes the percentage
// change between these two periods.
//
// The function performs two main operations:
// 1. Queries the transaction_items and transactions tables to count package upgrades
// 2. Stores the results in the statistics table with the name "installed-packages-30-days"
//
// If previousMonth is 0, the percentage change is not calculated to avoid division by zero.
// Any database errors are logged using the logger.Error function.
//
// The function does not return any values but updates the statistics table
// with the current count and percentage change.
func CountInstalledPackages() {
	var thisMonth, previousMonth int
	err := database.Db.QueryRow(`
      WITH last30days AS (
        SELECT ti.item_id
        FROM transaction_items ti
        JOIN transactions t ON t.transaction_id = ti.transaction_id
        WHERE t.begin_time >= NOW() - INTERVAL '30 days'
        AND ti.action = 'Install'
      ),

      last60days AS (
        SELECT ti.item_id
        FROM transaction_items ti
        JOIN transactions t ON t.transaction_id = ti.transaction_id
        WHERE t.begin_time >= NOW() - INTERVAL '60 days'
        AND t.begin_time < NOW() - INTERVAL '30 days'
        AND ti.action = 'Install'
      )

      SELECT
        (SELECT COUNT(*) FROM last30days) AS this_month,
        (SELECT COUNT(*) FROM last60days) AS previous_month;
      `).Scan(&thisMonth, &previousMonth)

	if err != nil {
		logger.Error("Error querying statistics: " + err.Error())
		return
	}

	var percentage float64
	if previousMonth > 0 {
		percentage = float64(thisMonth-previousMonth) / float64(previousMonth) * 100
	}

	_, err = database.Db.Exec(`
	        INSERT INTO statistics (name, value, percentage, updated_at)
	        VALUES ($1, $2, $3, NOW())
	        ON CONFLICT (name) DO UPDATE
	        SET value = $2, percentage = $3, updated_at = NOW()`,
		"installed-packages-30-days", thisMonth, percentage)

	if err != nil {
		logger.Error("Error inserting statistics: " + err.Error())
		return
	}
}

// CountUpgradedPackages queries and stores statistics about package upgrades in the database.
// It calculates the number of package upgrades in the last 30 days compared to the previous 30 days,
// computes the percentage change, and stores these values in the statistics table.
//
// The function performs two main operations:
// 1. Queries the number of upgrades for two consecutive 30-day periods
// 2. Stores the results in the statistics table with the name "upgraded-packages-30-days"
//
// If there's an error during the database operations, it logs the error and returns without updating statistics.
// The percentage change is calculated only if there were upgrades in the previous period.
func CountUpgradedPackages() {
	var thisMonth, previousMonth int
	err := database.Db.QueryRow(`
      WITH last30days AS (
        SELECT ti.item_id
        FROM transaction_items ti
        JOIN transactions t ON t.transaction_id = ti.transaction_id
        WHERE t.begin_time >= NOW() - INTERVAL '30 days'
        AND ti.action = 'Upgrade'
      ),

      last60days AS (
        SELECT ti.item_id
        FROM transaction_items ti
        JOIN transactions t ON t.transaction_id = ti.transaction_id
        WHERE t.begin_time >= NOW() - INTERVAL '60 days'
        AND t.begin_time < NOW() - INTERVAL '30 days'
        AND ti.action = 'Upgrade'
      )

      SELECT
        (SELECT COUNT(*) FROM last30days) AS this_month,
        (SELECT COUNT(*) FROM last60days) AS previous_month;
      `).Scan(&thisMonth, &previousMonth)

	if err != nil {
		logger.Error("Error querying statistics: " + err.Error())
		return
	}

	var percentage float64
	if previousMonth > 0 {
		percentage = float64(thisMonth-previousMonth) / float64(previousMonth) * 100
	}

	_, err = database.Db.Exec(`
	        INSERT INTO statistics (name, value, percentage, updated_at)
	        VALUES ($1, $2, $3, NOW())
	        ON CONFLICT (name) DO UPDATE
	        SET value = $2, percentage = $3, updated_at = NOW()`,
		"upgraded-packages-30-days", thisMonth, percentage)

	if err != nil {
		logger.Error("Error inserting statistics: " + err.Error())
		return
	}
}
