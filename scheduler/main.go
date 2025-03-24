package scheduler

import (
	"os"
	"regexp"

	"github.com/mileusna/crontab"
	"github.com/txlog/server/database"
	logger "github.com/txlog/server/logger"
)

// StartScheduler initializes and starts the scheduler system with two periodic
// jobs:
// - A housekeeping job that runs according to CRON_RETENTION_EXPRESSION
// environment variable
// - A statistics job that runs according to CRON_STATS_EXPRESSION environment
// variable
// The scheduler uses crontab for job scheduling and execution.
func StartScheduler() {
	ctab := crontab.New()
	ctab.MustAddJob(os.Getenv("CRON_RETENTION_EXPRESSION"), housekeepingJob)
	ctab.MustAddJob(os.Getenv("CRON_STATS_EXPRESSION"), statsJob)

	logger.Info("Scheduler: started.")
}

// statsJob executes a periodic task to collect and store system statistics.
// It manages concurrent access using a lock mechanism and performs the following:
//
// 1. Collects server counts for the current 30-day period and previous 30-day period
// 2. Calculates percentage change between periods
// 3. Stores results in the statistics table
//
// The function uses a "stats" lock to prevent concurrent execution across multiple instances.
// If another instance is already running this job, it will exit early.
//
// The function handles database operations and logs any errors that occur during execution.
// Results are stored in the statistics table with the key "servers-30-days".
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

	// servers-30-days: 350
	var thisMonth, previousMonth int
	err = database.Db.QueryRow(`
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

	// executions-30-days: 632 8%
	// installed-packages-30-days: 1.355 0%
	// upgraded-packages-30-days: 656 4%

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
