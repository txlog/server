package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
		restart := c.Query("restart")
		inactive := c.Query("inactive")

		copyFailFilter := false
		if strings.Contains(search, "copyfail:true") {
			copyFailFilter = true
			search = strings.ReplaceAll(search, "copyfail:true", "")
			search = strings.TrimSpace(search)
		}

		// Parse topology keywords: env:<name>, svc:<name>, pod:<id>
		envFilter := extractKeyword(&search, "env:")
		svcFilter := extractKeyword(&search, "svc:")
		podFilter := extractKeyword(&search, "pod:")
		search = strings.TrimSpace(search)

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

		var total int
		var queryArgs []interface{}

		activeFilter := "is_active = TRUE"
		// When searching by machine_id, include both active and inactive assets.
		// This allows users to view historical assets by their unique machine ID,
		// even if the asset is no longer active.
		if searchType == "machine_id" {
			activeFilter = "1=1"
		}

		// Build WHERE clause for filtering
		whereClause := ""
		paramNum := 1

		if search != "" {
			whereClause += " AND " + searchType + " ILIKE $" + strconv.Itoa(paramNum)
			queryArgs = append(queryArgs, util.FormatSearchTerm(search))
			paramNum++
		}

		if restart == "true" {
			whereClause += " AND needs_restarting IS TRUE"
		}

		if copyFailFilter {
			whereClause += " AND copy_fail IS TRUE"
		}

		// env:name → filter assets whose hostname matches the pattern of the named environment.
		if envFilter != "" {
			whereClause += ` AND EXISTS (
				SELECT 1
				FROM topology_patterns tp
				WHERE assets.hostname ~ tp.compiled_pattern
				  AND (regexp_match(assets.hostname, tp.compiled_pattern))[1] IN (
					SELECT match_value FROM environment_names
					WHERE name ILIKE $` + strconv.Itoa(paramNum) + `
					   OR match_value ILIKE $` + strconv.Itoa(paramNum) + `
				  )
				LIMIT 1
			)`
			queryArgs = append(queryArgs, envFilter)
			paramNum++
		}

		// svc:name → filter assets whose hostname matches a service pattern.
		if svcFilter != "" {
			whereClause += ` AND EXISTS (
				SELECT 1
				FROM topology_patterns tp
				WHERE assets.hostname ~ tp.compiled_pattern
				  AND (regexp_match(assets.hostname, tp.compiled_pattern))[2] IN (
					SELECT match_value FROM service_names
					WHERE name ILIKE $` + strconv.Itoa(paramNum) + `
					   OR match_value ILIKE $` + strconv.Itoa(paramNum) + `
				  )
				LIMIT 1
			)`
			queryArgs = append(queryArgs, svcFilter)
			paramNum++
		}

		// pod:id → filter assets whose hostname produces a matching :seq capture.
		if podFilter != "" {
			whereClause += ` AND EXISTS (
				SELECT 1
				FROM topology_patterns tp
				WHERE assets.hostname ~ tp.compiled_pattern
				  AND (regexp_match(assets.hostname, tp.compiled_pattern))[3] = $` + strconv.Itoa(paramNum) + `
				LIMIT 1
			)`
			queryArgs = append(queryArgs, podFilter)
			paramNum++
		}

		if inactive == "true" {
			whereClause += " AND last_seen < NOW() - INTERVAL '15 days'"
		}

		// Count query - direct on assets table (no LATERAL JOIN needed since os is now stored in assets)
		countQuery := `
			SELECT COUNT(DISTINCT hostname)
			FROM assets
			WHERE ` + activeFilter + whereClause

		err = database.QueryRowContext(c.Request.Context(), countQuery, queryArgs...).Scan(&total)
		if err != nil {
			logger.Error("Error counting assets:" + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}

		totalPages := (total + limit - 1) / limit

		// Main select query - direct on assets table
		// The os column is now stored directly in assets, updated on each execution
		selectQuery := `
			SELECT
				asset_id,
				hostname,
				last_seen,
				machine_id,
				os,
				needs_restarting,
				copy_fail
			FROM assets
			WHERE ` + activeFilter + whereClause + `
			ORDER BY hostname
			LIMIT $` + strconv.Itoa(paramNum) + ` OFFSET $` + strconv.Itoa(paramNum+1)

		queryArgs = append(queryArgs, limit, offset)
		rows, err = database.QueryContext(c.Request.Context(), selectQuery, queryArgs...)

		if err != nil {
			logger.Error("Error listing assets:" + err.Error())
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
			var os sql.NullString
			var copyFail sql.NullBool
			err := rows.Scan(
				&asset.ExecutionID,
				&asset.Hostname,
				&executedAt,
				&asset.MachineID,
				&os,
				&asset.NeedsRestarting,
				&copyFail,
			)
			if err != nil {
				logger.Error("Error iterating assets:" + err.Error())
				c.HTML(http.StatusInternalServerError, "500.html", gin.H{
					"error": err.Error(),
				})
				return
			}
			if executedAt.Valid {
				asset.ExecutedAt = &executedAt.Time
			}
			if os.Valid {
				asset.OS = os.String
			}
			if copyFail.Valid {
				asset.CopyFail = &copyFail.Bool
			}
			assets = append(assets, asset)
		}

		rows, err = database.QueryContext(c.Request.Context(), `
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
			"restart":      restart,
			"inactive":     inactive,
		})
	}
}

// DeleteMachineID returns a Gin handler function that deletes all records
// associated with a given machine ID from the database. It performs the
// following steps within a transaction:
//  1. Deletes transaction items related to transactions with the specified machine ID.
//  2. Deletes transactions with the specified machine ID.
//  3. Deletes executions with the specified machine ID.
//
// If any step fails, the transaction is rolled back and an error page is
// rendered. On success, the user is redirected to the assets page. The machine
// ID is expected as a URL parameter named "machine_id".
func DeleteMachineID(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		machineID := c.Param("machine_id")
		if machineID == "" {
			c.HTML(http.StatusBadRequest, "500.html", gin.H{
				"error": "machine_id is required",
			})
			return
		}

		tx, err := database.BeginTx(c.Request.Context(), nil)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": "Failed to start transaction: " + err.Error(),
			})
			return
		}
		defer func() {
			if p := recover(); p != nil {
				tx.Rollback()
				logger.Error(fmt.Sprintf("Critical panic caught deleting machine %s: %v", machineID, p))
				panic(p)
			}
		}()

		_, err = tx.Exec(`DELETE FROM transaction_items WHERE machine_id = $1`, machineID)
		if err != nil {
			tx.Rollback()
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": "Failed to delete transaction_items: " + err.Error(),
			})
			return
		}

		_, err = tx.Exec(`DELETE FROM transactions WHERE machine_id = $1`, machineID)
		if err != nil {
			tx.Rollback()
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": "Failed to delete transactions: " + err.Error(),
			})
			return
		}

		_, err = tx.Exec(`DELETE FROM executions WHERE machine_id = $1`, machineID)
		if err != nil {
			tx.Rollback()
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": "Failed to delete executions: " + err.Error(),
			})
			return
		}

		_, err = tx.Exec(`DELETE FROM assets WHERE machine_id = $1`, machineID)
		if err != nil {
			tx.Rollback()
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": "Failed to delete asset: " + err.Error(),
			})
			return
		}

		if err := tx.Commit(); err != nil {
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": "Failed to commit transaction: " + err.Error(),
			})
			return
		}

		c.Redirect(http.StatusSeeOther, "/assets")

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

		row := database.QueryRowContext(c.Request.Context(), `
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

		rows, err := database.QueryContext(c.Request.Context(), `
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

		rows, err = database.QueryContext(c.Request.Context(), `
      SELECT
        transaction_id,
        begin_time,
        actions,
        altered,
        "user",
        command_line,
        COALESCE(is_security_patch, false),
        COALESCE(vulns_fixed, 0),
        COALESCE(max_severity_fixed, '')
      FROM public.transactions
      WHERE machine_id = $1
      ORDER BY transaction_id DESC
      LIMIT 50`,
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
				&transaction.IsSecurityPatch,
				&transaction.VulnsFixed,
				&transaction.MaxSeverityFixed,
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

		rows, err = database.QueryContext(c.Request.Context(), `
      SELECT a.machine_id, a.hostname, a.last_seen, e.agent_version, e.os
      FROM assets a
      LEFT JOIN LATERAL (
        SELECT agent_version, os
        FROM executions
        WHERE machine_id = a.machine_id AND hostname = a.hostname
        ORDER BY executed_at DESC
        LIMIT 1
      ) e ON true
      WHERE a.hostname = $1
      AND a.machine_id != $2
      ORDER BY a.last_seen DESC;`,
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
		var copyFail sql.NullBool
		err = database.QueryRowContext(c.Request.Context(), `
      SELECT needs_restarting, restarting_reason, copy_fail
      FROM assets
      WHERE machine_id = $1
      LIMIT 1
      `, machineID).Scan(&needsRestarting, &restartingReason, &copyFail)
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

		var displayCopyFail bool
		if copyFail.Valid {
			displayCopyFail = copyFail.Bool
		} else {
			displayCopyFail = false
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
			"copy_fail":         displayCopyFail,
		})
	}
}

// extractKeyword finds and removes a "prefix<value>" token from the search string.
// The value is terminated by a space or end of string.
// Returns the extracted value (trimmed) and modifies *search in-place.
//
// Example:
//
//	search = "env:production webserver"
//	extractKeyword(&search, "env:") → "production", search = "webserver"
func extractKeyword(search *string, prefix string) string {
	s := *search
	idx := strings.Index(strings.ToLower(s), strings.ToLower(prefix))
	if idx == -1 {
		return ""
	}
	// Find end of value (space or EOL).
	rest := s[idx+len(prefix):]
	end := strings.IndexByte(rest, ' ')
	var value string
	if end == -1 {
		value = rest
		*search = strings.TrimSpace(s[:idx])
	} else {
		value = rest[:end]
		*search = strings.TrimSpace(s[:idx] + rest[end:])
	}
	return strings.TrimSpace(value)
}
