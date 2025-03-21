package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
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
			fmt.Println("Invalid JSON input:", err)
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
			fmt.Println("Error beginning transaction:", err)
			c.AbortWithStatusJSON(500, gin.H{"error": "Database error"})
			return
		}

		_, err = tx.Exec(`
      INSERT INTO executions (
        machine_id, hostname, executed_at, success, details, transactions_processed, transactions_sent
      ) VALUES (
        $1, $2, $3, $4, $5, $6, $7
      )`,
			body.MachineID,
			body.Hostname,
			executedAt,
			body.Success,
			body.Details,
			body.TransactionsProcessed,
			body.TransactionsSent)

		if err != nil {
			tx.Rollback()
			fmt.Println("Error inserting execution:", err)
			c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
			return
		}

		// Commit the database transaction
		if err = tx.Commit(); err != nil {
			tx.Rollback()
			fmt.Println("Error committing execution:", err)
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
			fmt.Println("Error querying executions:", err)
			c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
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
				&executedAt,
				&execution.Success,
				&execution.Details,
				&execution.TransactionsProcessed,
				&execution.TransactionsSent,
			)
			if err != nil {
				fmt.Println("Error iterating executions:", err)
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

func GetExecutionsIndex(database *sql.DB) gin.HandlerFunc {
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
