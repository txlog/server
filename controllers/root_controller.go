package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/txlog/server/models"
)

func GetRootIndex(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rows *sql.Rows
		var err error

		limit := 10
		page := 1

		if pageStr := c.DefaultQuery("page", "1"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}
		offset := (page - 1) * limit

		// First, get total count
		var total int
		err = database.QueryRow("SELECT COUNT(*) FROM executions").Scan(&total)
		if err != nil {
			fmt.Println("Error counting executions:", err)
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}

		totalPages := (total + limit - 1) / limit

		rows, err = database.Query(`
      SELECT
        id,
        machine_id,
        hostname,
        executed_at,
        success,
        details,
        transactions_processed,
        transactions_sent
      FROM executions
      ORDER BY executed_at DESC
      LIMIT $1 OFFSET $2
    `, limit, offset)

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
			"Context":      c,
			"title":        "Assets",
			"executions":   executions,
			"page":         page,
			"totalPages":   totalPages,
			"totalRecords": total,
			"limit":        limit,
			"offset":       offset,
		})
	}
}
