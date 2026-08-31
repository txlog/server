package scheduler

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/mileusna/crontab"
	"github.com/txlog/server/statistics"
)

// numericRegex is precompiled at package level to avoid recompilation on each housekeeping invocation
var numericRegex = regexp.MustCompile(`^[0-9]+$`)

// StartScheduler initializes and starts the scheduler system with periodic jobs:
//   - A housekeeping job that runs according to CRON_RETENTION_EXPRESSION
//     environment variable
//   - A statistics job that runs according to CRON_STATS_EXPRESSION environment
//     variable
//   - A materialized view refresh job that runs every 5 minutes
//
// The scheduler uses crontab for job scheduling and execution.
func StartScheduler(db *sql.DB) {
	ctab := crontab.New()
	ctab.MustAddJob(os.Getenv("CRON_RETENTION_EXPRESSION"), func() { housekeepingJob(db) })
	ctab.MustAddJob(os.Getenv("CRON_STATS_EXPRESSION"), func() { statsJob(db) })
	ctab.MustAddJob("0 * * * *", latestVersionJob)
	ctab.MustAddJob("*/5 * * * *", func() { refreshMaterializedViewsJob(db) })

	cronOsv := os.Getenv("CRON_OSV_EXPRESSION")
	if cronOsv == "" {
		cronOsv = "0 4 * * *"
	}
	ctab.MustAddJob(cronOsv, func() { UpdateVulnerabilitiesJob(db) })

	latestVersionJob()              // Run for the first time
	refreshMaterializedViewsJob(db) // Run for the first time
	slog.Info("Scheduler: started.")
}

func latestVersionJob() {
	resp, err := http.Get("https://txlog.rda.run/server/version")
	if err != nil {
		slog.Error("Error fetching latest version: " + err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Error reading response body: " + err.Error())
		return
	}

	version := strings.TrimSpace(string(body))
	if err := os.Setenv("LATEST_VERSION", version); err != nil {
		slog.Error("Failed to store latest version: " + err.Error())
		return
	}
	slog.Info("Latest version updated: " + version)
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
func refreshMaterializedViewsJob(db *sql.DB) {
	withLock(db, "refresh-materialized-views", func() {
		// CONCURRENTLY allows reads during the refresh but needs a unique index,
		// and the view may not exist at all before its migration has been
		// applied — expected on a first deployment, hence the debug level.
		views := []string{
			"mv_package_listing",
			"mv_dashboard_os_stats",
			"mv_dashboard_agent_stats",
			"mv_dashboard_most_updated",
		}
		for _, view := range views {
			if _, err := db.Exec(`REFRESH MATERIALIZED VIEW CONCURRENTLY ` + view); err != nil {
				if _, err := db.Exec(`REFRESH MATERIALIZED VIEW ` + view); err != nil {
					slog.Debug("Could not refresh " + view + ": " + err.Error())
				}
			}
		}

		slog.Debug("Materialized views refreshed successfully.")
	})
}

// statsJob executes statistical tasks for the system while ensuring only one instance
// runs at a time using a distributed lock mechanism.
//
// The function performs the following operations:
// 1. Attempts to acquire a lock named "stats"
// 2. If lock acquisition fails or another instance is running, exits early
// 3. Counts executions, installed packages, and upgraded packages for the last 30 days
// 4. Automatically releases the lock when the function completes
func statsJob(db *sql.DB) {
	withLock(db, "stats", func() {
		slog.Info("Statistics: executing task...")

		statistics.CountExecutions()
		statistics.CountInstalledPackages()
		statistics.CountUpgradedPackages()

		slog.Info("Statistics updated.")
	})
}

// housekeepingJob performs database cleanup by deleting old execution records.
// It uses a distributed lock mechanism to ensure only one instance runs at a time.
// The retention period is configured via CRON_RETENTION_DAYS environment variable
// (defaults to 7 days if not set). Records older than the retention period are
// deleted from the executions table. The function logs its progress and any errors
// encountered during the process.
func housekeepingJob(db *sql.DB) {
	withLock(db, "retention-days", func() { housekeep(db) })
}

func housekeep(db *sql.DB) {
	slog.Info("Housekeeping: executing task...")

	retentionDays := os.Getenv("CRON_RETENTION_DAYS")
	if retentionDays == "" {
		retentionDays = "7" // default to 7 days if not set
	}
	if numericRegex.MatchString(retentionDays) {
		// An interval literal cannot take a placeholder: "INTERVAL $1 day" is a
		// syntax error, so build the interval from the validated numeric string.
		if _, err := db.Exec("DELETE FROM executions WHERE executed_at < NOW() - ($1 || ' days')::interval", retentionDays); err != nil {
			slog.Error("Error deleting executions past the retention period: " + err.Error())
		}
	}

	// D11: Cleanup orphan transaction_items and transactions from inactive assets
	_, err := db.Exec(`
		DELETE FROM transaction_items ti
		WHERE NOT EXISTS (
			SELECT 1 FROM assets a
			WHERE a.machine_id = ti.machine_id AND a.is_active = TRUE
		)
		AND ti.machine_id IN (
			SELECT machine_id FROM assets WHERE is_active = FALSE AND deactivated_at < NOW() - INTERVAL '90 days'
		)
	`)
	if err != nil {
		slog.Error("Housekeeping: error cleaning orphan transaction_items: " + err.Error())
	}

	_, err = db.Exec(`
		DELETE FROM transactions t
		WHERE NOT EXISTS (
			SELECT 1 FROM assets a
			WHERE a.machine_id = t.machine_id AND a.is_active = TRUE
		)
		AND t.machine_id IN (
			SELECT machine_id FROM assets WHERE is_active = FALSE AND deactivated_at < NOW() - INTERVAL '90 days'
		)
	`)
	if err != nil {
		slog.Error("Housekeeping: error cleaning orphan transactions: " + err.Error())
	}

	slog.Info("Housekeeping: executions older than " + retentionDays + " days are deleted.")
}

// lockKey turns a job name into the 64-bit key a PostgreSQL advisory lock takes.
// Any stable mapping would do; SHA-256 makes a collision between two job names
// not worth thinking about. The top bit is cleared so the key is never negative,
// which keeps the arithmetic in IsJobRunning inside bigint range.
func lockKey(name string) int64 {
	sum := sha256.Sum256([]byte(name))
	return int64(binary.BigEndian.Uint64(sum[:8]) & math.MaxInt64)
}

// withLock runs fn while holding the PostgreSQL advisory lock for the named job,
// and does nothing at all when another instance already holds it. This is what
// keeps two replicas from running the same job against the same database.
//
// The lock is taken on a connection pinned for the duration, because an advisory
// lock belongs to the session that took it: unlocking over a different pooled
// connection would silently fail. Tying it to a session is also what makes the
// lock self-cleaning — if this process dies mid-job, PostgreSQL drops the lock
// when the backend goes away, so there is no stale lock for the next run to trip
// over and none to reap.
func withLock(db *sql.DB, name string, fn func()) {
	ctx := context.Background()

	conn, err := db.Conn(ctx)
	if err != nil {
		slog.Error("Failed to open a connection for the " + name + " lock: " + err.Error())
		return
	}
	defer conn.Close()

	key := lockKey(name)
	var locked bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", key).Scan(&locked); err != nil {
		slog.Error("Failed to acquire the " + name + " lock: " + err.Error())
		return
	}
	if !locked {
		slog.Debug("Another instance is running the " + name + " job.")
		return
	}

	defer func() {
		if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", key); err != nil {
			slog.Error("Failed to release the " + name + " lock: " + err.Error())
		}
	}()

	fn()
}

// IsJobRunning reports whether any instance currently holds the advisory lock
// for the named job. It reads pg_locks instead of trying to take the lock, so
// asking the question has no effect on the answer.
//
// pg_locks splits the 64-bit key across classid and objid, which is why the two
// halves are recombined here.
func IsJobRunning(db *sql.DB, name string) bool {
	var running bool
	err := db.QueryRow(`
		SELECT EXISTS (
		  SELECT 1 FROM pg_locks
		  WHERE locktype = 'advisory'
		    AND (classid::bigint << 32) | (objid::bigint) = $1
		)`, lockKey(name)).Scan(&running)
	if err != nil {
		slog.Error("Failed to check whether the " + name + " job is running: " + err.Error())
		return false
	}
	return running
}
