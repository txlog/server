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
