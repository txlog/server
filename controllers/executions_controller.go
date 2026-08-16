package controllers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/txlog/server/models"
)

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
