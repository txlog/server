package controllers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/models"
)

type MachineID struct {
	Hostname  string     `json:"hostname"`
	MachineID string     `json:"machine_id"`
	BeginTime *time.Time `json:"begin_time"`
}

func GetMachines(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		os := c.Query("os")
		agentVersion := c.Query("agent_version")

		var rows *sql.Rows
		var err error

		query := `
    SELECT
      hostname,
      machine_id
    FROM (
      SELECT
        hostname,
        agent_version,
        os,
        machine_id,
        ROW_NUMBER() OVER(PARTITION BY hostname ORDER BY executed_at DESC) as rn
      FROM
        executions
    ) sub
    WHERE
      sub.rn = 1`

		var params []interface{}
		var paramCount int

		if os != "" {
			logger.Debug("os: " + os)
			if os == "Undefined OS" {
				os = ""
			}
			paramCount++
			query += ` AND os = $` + strconv.Itoa(paramCount)
			params = append(params, os)
		}

		if agentVersion != "" {
			logger.Debug("agent_version: " + agentVersion)
			if agentVersion == "with undefined version" {
				agentVersion = ""
			}
			paramCount++
			query += ` AND agent_version = $` + strconv.Itoa(paramCount)
			params = append(params, agentVersion)
		}

		if len(params) > 0 {
			rows, err = database.Query(query+" ORDER BY hostname", params...)
		} else {
			rows, err = database.Query(query + " ORDER BY hostname")
		}

		if err != nil {
			logger.Error("Error querying assets: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		machines := []MachineID{}
		for rows.Next() {
			var machine MachineID
			err := rows.Scan(
				&machine.Hostname,
				&machine.MachineID,
			)
			if err != nil {
				logger.Error("Error iterating assets: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
				return
			}
			machines = append(machines, machine)
		}

		c.JSON(http.StatusOK, machines)
	}
}

// GetMachineIDs List the machine_id of the given hostname
//
//	@Summary		List machine IDs
//	@Description	List machine IDs
//	@Tags			machines
//	@Accept			json
//	@Produce		json
//	@Param			hostname	query		string	false	"Hostname"
//	@Success		200			{object}	interface{}
//	@Failure		400			{string}	string	"Invalid execution data"
//	@Failure		400			{string}	string	"Invalid JSON input"
//	@Failure		500			{string}	string	"Database error"
//	@Router			/v1/machines/ids [get]
func GetMachineIDs(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostname := c.Query("hostname")

		if hostname == "" {
			c.AbortWithStatusJSON(400, "hostname is required")
			return
		}

		var rows *sql.Rows
		var err error

		rows, err = database.Query(`
      SELECT hostname, machine_id, begin_time
      FROM transactions
      WHERE transaction_id IN (
        SELECT MIN(transaction_id)
        FROM transactions
        GROUP BY machine_id
      ) AND hostname = $1
      ORDER BY begin_time DESC`,
			hostname,
		)

		if err != nil {
			logger.Error("Error querying machine_id: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		machines := []MachineID{}
		for rows.Next() {
			var machine MachineID
			var beginTime sql.NullTime
			err := rows.Scan(
				&machine.Hostname,
				&machine.MachineID,
				&beginTime,
			)
			if err != nil {
				logger.Error("Error iterating machine_id: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
				return
			}
			if beginTime.Valid {
				machine.BeginTime = &beginTime.Time
			}
			machines = append(machines, machine)
		}

		c.JSON(http.StatusOK, machines)
	}
}

func GetMachineID(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		machineID := c.Param("machine_id")
		if machineID == "" {
			c.HTML(http.StatusBadRequest, "500.html", gin.H{
				"error": "machine_id is required",
			})
			return
		}

		// Last 10 executions with transactions sent
		// All machines with this hostname

		row := database.QueryRow(`
        SELECT hostname FROM executions WHERE machine_id = $1
    `, machineID)

		var hostname string
		if err := row.Scan(&hostname); err != nil {
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}

		if hostname == "" {
			c.HTML(http.StatusNotFound, "404.html", gin.H{
				"error": "Asset ID not found",
			})
			return
		}

		rows, err := database.Query(`
      SELECT id, machine_id, hostname, executed_at, success,
        details, transactions_processed, transactions_sent,
        agent_version, os
      FROM executions
      WHERE machine_id = $1
      AND transactions_sent > 0
      ORDER BY executed_at DESC
      LIMIT 10`,
			machineID)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}
		defer rows.Close()

		var executions []models.Execution
		for rows.Next() {
			var executedAt sql.NullTime
			var agentVersion sql.NullString
			var os sql.NullString
			var exec models.Execution
			err := rows.Scan(
				&exec.ExecutionID,
				&exec.MachineID,
				&exec.Hostname,
				&executedAt,
				&exec.Success,
				&exec.Details,
				&exec.TransactionsProcessed,
				&exec.TransactionsSent,
				&agentVersion,
				&os,
			)
			if err != nil {
				c.HTML(http.StatusInternalServerError, "500.html", gin.H{
					"error": err.Error(),
				})
				return
			}
			if executedAt.Valid {
				exec.ExecutedAt = &executedAt.Time
			}
			if agentVersion.Valid {
				exec.AgentVersion = agentVersion.String
			}
			if os.Valid {
				exec.OS = os.String
			}
			executions = append(executions, exec)
		}

		rows, err = database.Query(`
      SELECT
        transaction_id,
        begin_time,
        actions,
        altered,
        "user",
        command_line
      FROM public.transactions
      WHERE machine_id = $1
      ORDER BY transaction_id DESC`,
			machineID)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}
		defer rows.Close()

		var transactions []models.Transaction
		for rows.Next() {
			var beginTime sql.NullTime
			var transaction models.Transaction
			err := rows.Scan(
				&transaction.TransactionID,
				&beginTime,
				&transaction.Actions,
				&transaction.Altered,
				&transaction.User,
				&transaction.CommandLine,
			)
			if err != nil {
				c.HTML(http.StatusInternalServerError, "500.html", gin.H{
					"error": err.Error(),
				})
				return
			}
			if beginTime.Valid {
				transaction.BeginTime = &beginTime.Time
			}
			transactions = append(transactions, transaction)
		}

		rows, err = database.Query(`
      SELECT e.machine_id, e.hostname, e.executed_at, e.agent_version, e.os
      FROM executions e
      INNER JOIN (
        SELECT machine_id, MAX(executed_at) as max_executed_at
        FROM executions
        WHERE hostname = $1
        AND machine_id != $2
        GROUP BY machine_id
      ) latest ON e.machine_id = latest.machine_id
        AND e.executed_at = latest.max_executed_at
      WHERE e.hostname = $1
      AND e.machine_id != $2
      AND e.success is true
      ORDER BY e.executed_at DESC;`,
			hostname, machineID)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}
		defer rows.Close()

		var otherAssets []models.Execution
		for rows.Next() {
			var executedAt sql.NullTime
			var agentVersion sql.NullString
			var os sql.NullString
			var exec models.Execution
			err := rows.Scan(
				&exec.MachineID,
				&exec.Hostname,
				&executedAt,
				&agentVersion,
				&os,
			)
			if err != nil {
				c.HTML(http.StatusInternalServerError, "500.html", gin.H{
					"error": err.Error(),
				})
				return
			}
			if executedAt.Valid {
				exec.ExecutedAt = &executedAt.Time
			}
			if agentVersion.Valid {
				exec.AgentVersion = agentVersion.String
			}
			if os.Valid {
				exec.OS = os.String
			}
			otherAssets = append(otherAssets, exec)
		}

		// query if this asset must be restarted
		var needsRestarting sql.NullBool
		var restartingReason sql.NullString
		err = database.QueryRow(`
      SELECT needs_restarting, restarting_reason
      FROM executions
      WHERE machine_id = $1
      ORDER BY executed_at DESC
      LIMIT 1
      `, machineID).Scan(&needsRestarting, &restartingReason)
		if err != nil && err != sql.ErrNoRows {
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}

		// Prepare the boolean value for the template
		var displayNeedsRestarting bool
		if needsRestarting.Valid {
			displayNeedsRestarting = needsRestarting.Bool
		} else {
			// If needsRestarting.Valid is false (NULL in database),
			// displayNeedsRestarting will be false.
			// This ensures that {{ if .needs_restarting }} will only be true
			// if needs_restarting is TRUE in the database.
			displayNeedsRestarting = false
		}

		c.HTML(http.StatusOK, "machine_id.html", gin.H{
			"Context":           c,
			"title":             "Assets",
			"hostname":          hostname,
			"machine_id":        machineID,
			"transactions":      transactions,
			"executions":        executions,
			"other_assets":      otherAssets,
			"needs_restarting":  displayNeedsRestarting,
			"restarting_reason": restartingReason.String,
		})
	}
}
