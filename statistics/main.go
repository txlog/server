package statistics

import (
	"github.com/txlog/server/database"
	"github.com/txlog/server/logger"
)

// countStat runs a query that returns two counts — the last 30 days and the 30
// days before those — and stores the recent count under name in the statistics
// table, together with the percentage change between the two periods. A
// previous period of zero leaves the percentage at zero rather than dividing by
// it. Errors are logged and end the update; the row keeps its previous value.
func countStat(name, query string) {
	var thisMonth, previousMonth int
	if err := database.Db.QueryRow(query).Scan(&thisMonth, &previousMonth); err != nil {
		logger.Error("Error querying statistics: " + err.Error())
		return
	}

	var percentage float64
	if previousMonth > 0 {
		percentage = float64(thisMonth-previousMonth) / float64(previousMonth) * 100
	}

	_, err := database.Db.Exec(`
	        INSERT INTO statistics (name, value, percentage, updated_at)
	        VALUES ($1, $2, $3, NOW())
	        ON CONFLICT (name) DO UPDATE
	        SET value = $2, percentage = $3, updated_at = NOW()`,
		name, thisMonth, percentage)
	if err != nil {
		logger.Error("Error inserting statistics: " + err.Error())
	}
}

// CountExecutions stores the number of agent executions recorded in the last 30
// days under "executions-30-days", with the change against the 30 days before.
func CountExecutions() {
	countStat("executions-30-days", `
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
	      `)
}

// CountInstalledPackages stores the number of packages installed in the last 30
// days under "installed-packages-30-days", with the change against the 30 days
// before.
func CountInstalledPackages() {
	countStat("installed-packages-30-days", `
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
      `)
}

// CountUpgradedPackages stores the number of packages upgraded in the last 30
// days under "upgraded-packages-30-days", with the change against the 30 days
// before.
func CountUpgradedPackages() {
	countStat("upgraded-packages-30-days", `
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
      `)
}
