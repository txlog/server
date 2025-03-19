package scheduler

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/go-co-op/gocron/v2"
	"github.com/txlog/server/database"
)

// StartScheduler initializes and starts a cron scheduler for database
// maintenance tasks. It configures a job to delete execution records older than
// a specified number of days, defined by the CRON_RETENTION_DAYS environment
// variable. If the variable is not set or is invalid, it defaults to 7 days.
// The cron expression for the job is read from the CRON_RETENTION_EXPRESSION
// environment variable. The scheduler runs until the application shuts down, at
// which point it is gracefully stopped.
func StartScheduler() {
	s, _ := gocron.NewScheduler()
	defer func() { _ = s.Shutdown() }()

	_, _ = s.NewJob(
		gocron.CronJob(os.Getenv("CRON_RETENTION_EXPRESSION"), false),
		gocron.NewTask(
			func() {
				retentionDays := os.Getenv("CRON_RETENTION_DAYS")
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
}

// executeWithLock attempts to acquire a lock for the given task name and, if
// successful, executes the task. It uses a mutex to prevent concurrent
// execution of the same task across multiple instances. If the lock cannot be
// acquired, it logs that the task is already running. After acquiring the lock,
// it defers the release of the lock to ensure it's always released when the
// function exits. It simulates a task execution with a 5-second sleep.
//
// Parameters:
//   - taskName: A string representing the name of the task to execute. This name is used as the
//     key for the mutex.
//
// Example:
//
//	executeHousekeeping("my_important_task")
func executeHousekeeping(taskName string) {
	if acquireLock(taskName) {
		defer releaseLock(taskName)
		log.Printf("Executando tarefa: %s", taskName)

		retentionDays := os.Getenv("CRON_RETENTION_DAYS")
		if retentionDays == "" {
			retentionDays = "7" // default to 7 days if not set
		}
		if regexp.MustCompile(`^[0-9]+$`).MatchString(retentionDays) {
			_, _ = database.Db.Exec("DELETE FROM executions WHERE executed_at < NOW() - INTERVAL '" + retentionDays + " day'")
		}

		fmt.Println("Housekeeping: executions older than " + retentionDays + " days are deleted.")

		log.Printf("Tarefa %s concluída", taskName)
	} else {
		log.Printf("Tarefa %s já em execução em outra instância", taskName)
	}
}

// acquireLock attempts to acquire a lock for a given task name. It starts a
// database transaction, checks if a lock entry exists for the task, creates one
// if it doesn't, and attempts to update the lock status to 'locked'. It returns
// true if the lock was successfully acquired, and false otherwise. If any error
// occurs during the process, it logs the error and rolls back the transaction.
func acquireLock(taskName string) bool {
	tx, err := database.Db.Begin()
	if err != nil {
		log.Printf("Falha ao iniciar transação: %v", err)
		return false
	}
	defer tx.Rollback()

	var locked bool
	err = tx.QueryRow(`SELECT locked FROM cron_locks WHERE task_name = $1`, taskName).Scan(&locked)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Falha ao verificar lock: %v", err)
		return false
	}

	if err == sql.ErrNoRows {
		_, err = tx.Exec(`INSERT INTO cron_locks (task_name, locked) VALUES ($1, FALSE)`, taskName)
		if err != nil {
			log.Printf("Falha ao criar registro de lock: %v", err)
			return false
		}
		locked = false
	}

	if locked {
		return false
	}

	_, err = tx.Exec(`UPDATE cron_locks SET locked = TRUE, locked_at = NOW() WHERE task_name = $1`, taskName)
	if err != nil {
		log.Printf("Falha ao adquirir lock: %v", err)
		return false
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Falha ao commitar transação: %v", err)
		return false
	}

	return true
}

// releaseLock releases a lock in the database for a given task name. It updates
// the cron_locks table, setting the locked column to FALSE and locked_at to
// NULL for the specified task. If an error occurs during the database
// operation, it logs the error.
func releaseLock(taskName string) {
	_, err := database.Db.Exec(`UPDATE cron_locks SET locked = FALSE, locked_at = NULL WHERE task_name = $1`, taskName)
	if err != nil {
		log.Printf("Falha ao liberar lock: %v", err)
	}
}
