package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/models"
)

// PostExecutions Create a new execution
//
//	@Summary		Create a new execution
//	@Description	Create a new execution
//	@Tags			executions
//	@Accept			json
//	@Produce		json
//	@Param			Execution	body		models.Execution	true	"Execution data"
//	@Success		200			{string}	string				"Execution created"
//	@Failure		400			{string}	string				"Invalid execution data"
//	@Failure		400			{string}	string				"Invalid JSON input"
//	@Failure		500			{string}	string				"Database error"
//	@Router			/v1/executions [post]
func PostExecutions(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		body := models.Execution{}
		data, err := c.GetRawData()
		if err != nil {
			c.AbortWithStatusJSON(400, "Invalid execution data")
			return
		}
		err = json.Unmarshal(data, &body)
		if err != nil {
			c.AbortWithStatusJSON(400, "Invalid JSON input")
			logger.Error("Invalid JSON input:" + err.Error())
			return
		}

		// Convert *time.Time to sql.NullTime
		var executedAt sql.NullTime
		if body.ExecutedAt != nil {
			executedAt.Time = *body.ExecutedAt
			executedAt.Valid = true
		}

		// Start database transaction
		tx, err := database.Begin()
		if err != nil {
			logger.Error("Error beginning transaction:" + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"error": "Database error"})
			return
		}

		_, err = tx.Exec(`
      INSERT INTO executions (
        machine_id, hostname, executed_at, success, details, transactions_processed, transactions_sent, agent_version, os
      ) VALUES (
        $1, $2, $3, $4, $5, $6, $7, $8, $9
      )`,
			body.MachineID,
			body.Hostname,
			executedAt,
			body.Success,
			body.Details,
			body.TransactionsProcessed,
			body.TransactionsSent,
			body.AgentVersion,
			body.OS)

		if err != nil {
			tx.Rollback()
			logger.Error("Error inserting execution:" + err.Error())
			c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
			return
		}

		// Commit the database transaction
		if err = tx.Commit(); err != nil {
			tx.Rollback()
			logger.Error("Error committing execution:" + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Execution created"})
	}
}

// GetExecutions List executions
//
//	@Summary		List executions
//	@Description	List executions
//	@Tags			executions
//	@Accept			json
//	@Produce		json
//	@Param			machine_id	query		string	false	"Machine ID"
//	@Param			success		query		boolean	false	"Success"
//	@Success		200			{object}	interface{}
//	@Failure		400			{string}	string	"Invalid execution data"
//	@Failure		400			{string}	string	"Invalid JSON input"
//	@Failure		500			{string}	string	"Database error"
//	@Router			/v1/executions [get]
func GetExecutions(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		machineID := c.Query("machine_id")
		success := c.Query("success")

		if machineID == "" {
			c.AbortWithStatusJSON(400, "machine_id is required")
			return
		}

		var rows *sql.Rows
		var err error
		if success != "" {
			rows, err = database.Query(
				`SELECT * FROM executions WHERE machine_id = $1 AND success = $2 ORDER BY executed_at DESC;`,
				machineID, success,
			)
		} else {
			rows, err = database.Query(
				`SELECT * FROM executions WHERE machine_id = $1 ORDER BY executed_at DESC;`,
				machineID,
			)
		}

		if err != nil {
			logger.Error("Error querying executions:" + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		executions := []models.Execution{}
		for rows.Next() {
			var execution models.Execution
			var executedAt sql.NullTime
			var agentVersion sql.NullString
			var os sql.NullString
			err := rows.Scan(
				&execution.ExecutionID,
				&execution.MachineID,
				&execution.Hostname,
				&executedAt,
				&execution.Success,
				&execution.Details,
				&execution.TransactionsProcessed,
				&execution.TransactionsSent,
				&agentVersion,
				&os,
			)
			execution.AgentVersion = agentVersion.String
			execution.OS = os.String
			if err != nil {
				logger.Error("Error iterating executions:" + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
				return
			}
			if executedAt.Valid {
				execution.ExecutedAt = &executedAt.Time
			}
			executions = append(executions, execution)
		}

		c.JSON(http.StatusOK, executions)
	}
}

func GetExecutionID(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var execution models.Execution
		if err := c.ShouldBindUri(&execution); err != nil {
			c.JSON(400, gin.H{"msg": err.Error()})
			return
		}

		row := database.QueryRow(`
      SELECT id, machine_id, hostname, executed_at, success,
        details, transactions_processed, transactions_sent,
        agent_version, os
      FROM executions
      WHERE id = $1`,
			execution.ExecutionID)

		var executedAt sql.NullTime
		var agentVersion sql.NullString
		var os sql.NullString
		execution = models.Execution{}
		err := row.Scan(
			&execution.ExecutionID,
			&execution.MachineID,
			&execution.Hostname,
			&executedAt,
			&execution.Success,
			&execution.Details,
			&execution.TransactionsProcessed,
			&execution.TransactionsSent,
			&agentVersion,
			&os,
		)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		if executedAt.Valid {
			execution.ExecutedAt = &executedAt.Time
		}
		if agentVersion.Valid {
			execution.AgentVersion = agentVersion.String
		}
		if os.Valid {
			execution.OS = os.String
		}

		c.HTML(http.StatusOK, "execution_id.html", gin.H{
			"Context":   c,
			"title":     "Execution",
			"execution": execution,
		})
	}
}
