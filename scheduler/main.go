package scheduler

import (
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/mileusna/crontab"
	"github.com/txlog/server/database"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/statistics"
)

// StartScheduler initializes and starts the scheduler system with periodic jobs:
//   - A housekeeping job that runs according to CRON_RETENTION_EXPRESSION
//     environment variable
//   - A statistics job that runs according to CRON_STATS_EXPRESSION environment
//     variable
//   - A materialized view refresh job that runs every 5 minutes
//
// The scheduler uses crontab for job scheduling and execution.
func StartScheduler() {
	ctab := crontab.New()
	ctab.MustAddJob(os.Getenv("CRON_RETENTION_EXPRESSION"), housekeepingJob)
	ctab.MustAddJob(os.Getenv("CRON_STATS_EXPRESSION"), statsJob)
	ctab.MustAddJob("0 * * * *", latestVersionJob)
	ctab.MustAddJob("*/5 * * * *", refreshMaterializedViewsJob)

	latestVersionJob()            // Run for the first time
	refreshMaterializedViewsJob() // Run for the first time
	logger.Info("Scheduler: started.")
}

func latestVersionJob() {
	resp, err := http.Get("https://txlog.rda.run/server/version")
	if err != nil {
		logger.Error("Error fetching latest version: " + err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading response body: " + err.Error())
		return
	}

	version := strings.TrimSpace(string(body))
	os.Setenv("LATEST_VERSION", version)
	logger.Info("Latest version updated: " + version)
}

// refreshMaterializedViewsJob refreshes the materialized views used for performance optimization.
// It uses a distributed lock mechanism to ensure only one instance runs at a time.
// Currently refreshes:
//   - mv_package_listing: Pre-computed package listing data for the /packages endpoint
//
// Note: The /assets endpoint no longer uses a materialized view since the os column
// is now stored directly in the assets table and updated in real-time.
//
// This job should run frequently (every 5 minutes) to keep the data relatively fresh
// while avoiding the expensive CTEs on each request.
func refreshMaterializedViewsJob() {
	lockName := "refresh-materialized-views"

	locked, err := acquireLock(lockName)
	if err != nil {
		logger.Error("Error acquiring lock for materialized view refresh: " + err.Error())
		return
	}

	if !locked {
		// Another instance is already refreshing
		return
	}

	defer releaseLock(lockName)

	// Refresh the package listing materialized view
	// Using CONCURRENTLY to allow reads during refresh (requires UNIQUE index)
	_, err = database.Db.Exec(`REFRESH MATERIALIZED VIEW CONCURRENTLY mv_package_listing`)
	if err != nil {
		// If CONCURRENTLY fails (e.g., first run or no unique index), try without it
		_, err = database.Db.Exec(`REFRESH MATERIALIZED VIEW mv_package_listing`)
		if err != nil {
			// View might not exist yet (migration not applied)
			// This is expected on first deployment, so we only log at debug level
			logger.Debug("Could not refresh mv_package_listing: " + err.Error())
			return
		}
	}

	logger.Debug("Materialized views refreshed successfully.")
}

// statsJob executes statistical tasks for the system while ensuring only one instance
// runs at a time using a distributed lock mechanism.
//
// The function performs the following operations:
// 1. Attempts to acquire a lock named "stats"
// 2. If lock acquisition fails or another instance is running, exits early
// 3. Counts executions, installed packages, and upgraded packages for the last 30 days
// 4. Automatically releases the lock when the function completes
func statsJob() {
	logger.Info("Statistics: executing task...")

	lockName := "stats"

	locked, err := acquireLock(lockName)
	if err != nil {
		logger.Error("Error acquiring lock: " + err.Error())
		return
	}

	if !locked {
		logger.Info("Another instance is running this job.")
		return
	}

	defer releaseLock(lockName)

	statistics.CountExecutions()
	statistics.CountInstalledPackages()
	statistics.CountUpgradedPackages()

	logger.Info("Statistics updated.")
}

// housekeepingJob performs database cleanup by deleting old execution records.
// It uses a distributed lock mechanism to ensure only one instance runs at a time.
// The retention period is configured via CRON_RETENTION_DAYS environment variable
// (defaults to 7 days if not set). Records older than the retention period are
// deleted from the executions table. The function logs its progress and any errors
// encountered during the process.
func housekeepingJob() {
	logger.Info("Housekeeping: executing task...")

	lockName := "retention-days"

	locked, err := acquireLock(lockName)
	if err != nil {
		logger.Error("Error acquiring lock: " + err.Error())
		return
	}

	if !locked {
		logger.Info("Another instance is running this job.")
		return
	}

	defer releaseLock(lockName)

	retentionDays := os.Getenv("CRON_RETENTION_DAYS")
	if retentionDays == "" {
		retentionDays = "7" // default to 7 days if not set
	}
	if regexp.MustCompile(`^[0-9]+$`).MatchString(retentionDays) {
		_, _ = database.Db.Exec("DELETE FROM executions WHERE executed_at < NOW() - INTERVAL $1 day", retentionDays)
	}

	logger.Info("Housekeeping: executions older than " + retentionDays + " days are deleted.")
}

// acquireLock attempts to obtain a lock for a given job name in the cron_lock table.
// It uses a PostgreSQL INSERT with ON CONFLICT DO NOTHING to ensure atomic locking.
//
// Parameters:
//   - lockName: The name of the job to lock
//
// Returns:
//   - bool: true if the lock was acquired, false if it already exists
//   - error: An error object if the database operation fails
func acquireLock(lockName string) (bool, error) {
	res, err := database.Db.Exec(`INSERT INTO cron_lock (job_name, locked_at) VALUES ($1, NOW()) ON CONFLICT (job_name) DO NOTHING`, lockName)
	if err != nil {
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

// releaseLock removes a lock entry from the cron_lock table for a given job name.
// It helps in cleaning up the lock after a job has completed or when releasing a lock is necessary.
//
// Parameters:
//   - lockName: string representing the name of the job whose lock needs to be released
//
// Returns:
//   - error: returns any error that occurred during the lock release operation,
//     nil if successful
func releaseLock(lockName string) error {
	_, err := database.Db.Exec(`DELETE FROM cron_lock WHERE job_name = $1`, lockName)
	return err
}
