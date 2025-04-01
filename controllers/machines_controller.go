package controllers

import (
	"database/sql"
	"net/http"
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

		c.HTML(http.StatusOK, "machine_id.html", gin.H{
			"Context":    c,
			"title":      "Assets",
			"hostname":   hostname,
			"machine_id": machineID,
			"executions": executions,
		})
	}
}
