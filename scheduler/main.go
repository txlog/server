package scheduler

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/txlog/server/database"
)

// StartScheduler initializes and runs a scheduled task for database housekeeping.
// It creates a new scheduler that runs every 2 hours at a random minute/second
// to delete old execution records from the database. The retention period is
// controlled by the EXECUTION_RETENTION_DAYS environment variable (defaults to 7 days).
// The random scheduling helps distribute the load when running multiple instances.
func StartScheduler() {
	s, _ := gocron.NewScheduler()
	defer func() { _ = s.Shutdown() }()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Run every two hours at a random minute/second
	crontab := fmt.Sprintf("%d %d */2 * * *", r.Intn(59), r.Intn(59))
	_, _ = s.NewJob(
		gocron.CronJob(
			crontab,
			true,
		),
		gocron.NewTask(
			func() {
				retentionDays := os.Getenv("EXECUTION_RETENTION_DAYS")
				if retentionDays == "" {
					retentionDays = "7" // default to 7 days if not set
				}
				if regexp.MustCompile(`^[0-9]+$`).MatchString(retentionDays) {
					_, _ = database.Db.Exec("DELETE FROM executions WHERE executed_at < NOW() - INTERVAL '" + retentionDays + " day'")
				}

				fmt.Println("Housekeeping: executions older than " + retentionDays + " days are deleted.")
			},
		),
	)

	s.Start()
	fmt.Println("Scheduler started: " + crontab)
}
