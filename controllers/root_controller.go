package controllers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/txlog/server/models"
)

func GetRootIndex(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rows *sql.Rows
		var err error

		rows, err = database.Query(`
      SELECT
        id, machine_id, hostname, executed_at, success, details,
        transactions_processed, transactions_sent
      FROM executions
      ORDER BY executed_at DESC
    `)

		if err != nil {
			fmt.Println("Error listing executions:", err)
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}
		defer rows.Close()

		executions := []models.Execution{}
		for rows.Next() {
			var execution models.Execution
			var executedAt sql.NullTime
			err := rows.Scan(
				&execution.ExecutionID,
				&execution.MachineID,
				&execution.Hostname,
				&execution.ExecutedAt,
				&execution.Success,
				&execution.Details,
				&execution.TransactionsProcessed,
				&execution.TransactionsSent,
			)
			if err != nil {
				fmt.Println("Error iterating machine_id:", err)
				c.HTML(http.StatusInternalServerError, "500.html", gin.H{
					"error": err.Error(),
				})
				return
			}
			if executedAt.Valid {
				execution.ExecutedAt = &executedAt.Time
			}
			executions = append(executions, execution)
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"Context":    c,
			"title":      "Assets",
			"executions": executions,
		})
	}
}
