package controllers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/models"
	"github.com/txlog/server/util"
)

// GetAssetsIndex returns a Gin handler function that serves the root index page.
// It takes a database connection as parameter and returns HTML content with:
//   - A paginated list of executions from the database
//   - Statistics data
//   - Pagination information
//
// The handler supports query parameter "page" for pagination.
// Each page shows up to 10 records.
//
// Parameters:
//   - database: *sql.DB - The database connection to query data from
//
// Returns:
//   - gin.HandlerFunc - A handler that renders the index.html template with execution and statistics data
//
// The handler will return HTTP 500 if there are any database errors during:
//   - Counting total executions
//   - Querying executions data
//   - Querying statistics data
func GetAssetsIndex(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rows *sql.Rows
		var err error

		search := c.Query("search")

		searchType := "hostname"
		if len(search) == 32 && !util.ContainsSpecialCharacters(search) {
			searchType = "machine_id"
		}

		limit := 100
		page := 1

		if pageStr := c.DefaultQuery("page", "1"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}
		offset := (page - 1) * limit

		// First, get total asset count
		var total int
		var query string

		if search != "" {
			query = `
        SELECT
          count(hostname)
        FROM (
          SELECT
            hostname,
            ROW_NUMBER() OVER(PARTITION BY hostname ORDER BY executed_at DESC) as rn
          FROM
            executions
          WHERE
            ` + searchType + ` ILIKE $1
        ) sub
        WHERE
          sub.rn = 1
      `
			err = database.QueryRow(query, util.FormatSearchTerm(search)).Scan(&total)
		} else {
			query = `
        SELECT
          count(hostname)
        FROM (
          SELECT
            hostname,
            ROW_NUMBER() OVER(PARTITION BY hostname ORDER BY executed_at DESC) as rn
          FROM
            executions
        ) sub
        WHERE
          sub.rn = 1
      `
			err = database.QueryRow(query).Scan(&total)
		}

		if err != nil {
			logger.Error("Error counting executions:" + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}

		totalPages := (total + limit - 1) / limit

		if search != "" {
			query = `
        SELECT
          execution_id,
          hostname,
          executed_at,
          machine_id,
          os,
          needs_restarting
        FROM (
          SELECT
            id as execution_id,
            hostname,
            executed_at,
            machine_id,
            os,
            needs_restarting,
            ROW_NUMBER() OVER(PARTITION BY hostname ORDER BY executed_at DESC) as rn
          FROM
            executions
          WHERE
            ` + searchType + ` ILIKE $3
        ) sub
        WHERE
            sub.rn = 1
        ORDER BY
            hostname
        LIMIT $1 OFFSET $2
      `
			rows, err = database.Query(query, limit, offset, util.FormatSearchTerm(search))
		} else {
			query = `
        SELECT
          execution_id,
          hostname,
          executed_at,
          machine_id,
          os,
          needs_restarting
        FROM (
          SELECT
            id as execution_id,
            hostname,
            executed_at,
            machine_id,
            os,
            needs_restarting,
            ROW_NUMBER() OVER(PARTITION BY hostname ORDER BY executed_at DESC) as rn
          FROM
            executions
        ) sub
        WHERE
          sub.rn = 1
        ORDER BY
          hostname
        LIMIT $1 OFFSET $2
      `
			rows, err = database.Query(query, limit, offset)
		}

		if err != nil {
			logger.Error("Error listing executions:" + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}
		defer rows.Close()

		assets := []models.Execution{}
		for rows.Next() {
			var asset models.Execution
			var executedAt sql.NullTime
			err := rows.Scan(
				&asset.ExecutionID,
				&asset.Hostname,
				&asset.ExecutedAt,
				&asset.MachineID,
				&asset.OS,
				&asset.NeedsRestarting,
			)
			if err != nil {
				logger.Error("Error iterating machine_id:" + err.Error())
				c.HTML(http.StatusInternalServerError, "500.html", gin.H{
					"error": err.Error(),
				})
				return
			}
			if executedAt.Valid {
				asset.ExecutedAt = &executedAt.Time
			}
			assets = append(assets, asset)
		}

		rows, err = database.Query(`
      SELECT
        name,
        value,
        percentage,
        updated_at
      FROM statistics;`)

		if err != nil {
			logger.Error("Error listing executions:" + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}
		defer rows.Close()

		statistics := []models.Statistic{}

		for rows.Next() {
			var statistic = models.Statistic{}
			var updatedAt sql.NullTime
			err := rows.Scan(
				&statistic.Name,
				&statistic.Value,
				&statistic.Percentage,
				&statistic.UpdatedAt,
			)
			if err != nil {
				logger.Error("Error iterating machine_id:" + err.Error())
				c.HTML(http.StatusInternalServerError, "500.html", gin.H{
					"error": err.Error(),
				})
				return
			}
			if updatedAt.Valid {
				statistic.UpdatedAt = &updatedAt.Time
			}
			statistics = append(statistics, statistic)
		}

		c.HTML(http.StatusOK, "assets.html", gin.H{
			"Context":      c,
			"title":        "Assets",
			"assets":       assets,
			"page":         page,
			"totalPages":   totalPages,
			"totalRecords": total,
			"limit":        limit,
			"offset":       offset,
			"statistics":   statistics,
			"search":       search,
		})
	}
}

// GetMachineID returns a Gin handler function that processes requests for machine-specific information.
// It retrieves and displays detailed information about a specific machine identified by its machine_id,
// including:
//   - The last 10 executions where transactions were sent
//   - All transactions associated with the machine
//   - Information about other machines with the same hostname
//   - The machine's current restart status
//
// Parameters:
//   - database: *sql.DB - A pointer to the SQL database connection
//
// Returns:
//   - gin.HandlerFunc - A Gin handler that processes the request and renders the machine_id.html template
//
// The handler expects a machine_id parameter in the request URL. It will return:
//   - 400 Bad Request if machine_id is empty
//   - 404 Not Found if the machine doesn't exist
//   - 500 Internal Server Error if there are database errors
//   - 200 OK with the rendered template on success
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
