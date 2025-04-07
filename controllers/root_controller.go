package controllers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/models"
)

// GetRootIndex returns a Gin handler function that serves the root index page.
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
func GetRootIndex(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rows *sql.Rows
		var err error

		search := c.Query("search")

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
		err = database.QueryRow(`
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
    `).Scan(&total)
		if err != nil {
			logger.Error("Error counting executions:" + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}

		totalPages := (total + limit - 1) / limit

		rows, err = database.Query(`
    SELECT
        hostname,
        executed_at,
        machine_id
    FROM (
        SELECT
            hostname,
            executed_at,
            machine_id,
            ROW_NUMBER() OVER(PARTITION BY hostname ORDER BY executed_at DESC) as rn
        FROM
            executions
    ) sub
    WHERE
        sub.rn = 1
    ORDER BY
        hostname
    LIMIT $1 OFFSET $2
  `, limit, offset)

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
				&asset.Hostname,
				&asset.ExecutedAt,
				&asset.MachineID,
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

		c.HTML(http.StatusOK, "index.html", gin.H{
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
