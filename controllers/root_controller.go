package controllers

import (
	"database/sql"
	"net/http"

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

		// Get OS statistics
		rows, err = database.Query(`
      SELECT
        os,
        COUNT(DISTINCT machine_id) AS num_machines
      FROM (
        SELECT
          os,
          machine_id,
          hostname,
          ROW_NUMBER() OVER(PARTITION BY hostname ORDER BY executed_at DESC) as rn
        FROM
          executions
      ) sub
      WHERE
        sub.rn = 1
      GROUP BY
        os
      ORDER BY
        num_machines DESC;`)

		if err != nil {
			logger.Error("Error getting OS statistics:" + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}
		defer rows.Close()

		type OSStats struct {
			OS          string
			NumMachines int
		}

		assetsByOS := []OSStats{}
		for rows.Next() {
			var stat OSStats
			if err := rows.Scan(&stat.OS, &stat.NumMachines); err != nil {
				logger.Error("Error scanning OS statistics:" + err.Error())
				c.HTML(http.StatusInternalServerError, "500.html", gin.H{
					"error": err.Error(),
				})
				return
			}
			assetsByOS = append(assetsByOS, stat)
		}

		// Get Agent statistics
		rows, err = database.Query(`
      SELECT
        agent_version,
        COUNT(DISTINCT machine_id) AS num_machines
      FROM (
        SELECT
          agent_version,
          machine_id,
          hostname,
          ROW_NUMBER() OVER(PARTITION BY hostname ORDER BY executed_at DESC) as rn
        FROM
          executions
      ) sub
      WHERE
        sub.rn = 1
      GROUP BY
        agent_version
      ORDER BY
        num_machines DESC;`)

		if err != nil {
			logger.Error("Error getting Agent statistics:" + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}
		defer rows.Close()

		type AgentStats struct {
			AgentVersion string
			NumMachines  int
		}

		assetsByAgentVersion := []AgentStats{}
		for rows.Next() {
			var stat AgentStats
			if err := rows.Scan(&stat.AgentVersion, &stat.NumMachines); err != nil {
				logger.Error("Error scanning Agent statistics:" + err.Error())
				c.HTML(http.StatusInternalServerError, "500.html", gin.H{
					"error": err.Error(),
				})
				return
			}
			assetsByAgentVersion = append(assetsByAgentVersion, stat)
		}
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Context":              c,
			"title":                "Transaction Overview",
			"statistics":           statistics,
			"assetsByOS":           assetsByOS,
			"assetsByAgentVersion": assetsByAgentVersion,
		})
	}
}
