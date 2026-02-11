package controllers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/models"
	"golang.org/x/sync/errgroup"
)

type OSStats struct {
	OS          string
	NumMachines int
}

type AgentStats struct {
	AgentVersion string
	NumMachines  int
}

type DuplicatedAsset struct {
	Hostname    string
	NumMachines int
}

type UpdatedPackage struct {
	Package              string
	TotalUpdates         int
	DistinctHostsUpdated int
}

// GetRootIndex returns a Gin handler function that serves the root index page.
// It takes a database connection as parameter and returns HTML content with:
//   - Statistics data
//   - Asset counts by OS and agent version
//   - Duplicated assets and most updated packages
//
// All dashboard queries are executed in parallel to minimize response time.
//
// Parameters:
//   - database: *sql.DB - The database connection to query data from
//
// Returns:
//   - gin.HandlerFunc - A handler that renders the index.html template with execution and statistics data
//
// The handler will return HTTP 500 if there are any database errors.
func GetRootIndex(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			statistics           []models.Statistic
			totalActiveAssets    int
			assetsByOS           []OSStats
			assetsByAgentVersion []AgentStats
			duplicatedAssets     []DuplicatedAsset
			mostUpdatedPackages  []UpdatedPackage
		)

		g := new(errgroup.Group)

		g.Go(func() error {
			var err error
			statistics, err = getStatistics(database)
			return err
		})

		g.Go(func() error {
			var err error
			totalActiveAssets, err = getTotalActiveAssets(database)
			return err
		})

		g.Go(func() error {
			var err error
			assetsByOS, err = getAssetsByOS(database)
			return err
		})

		g.Go(func() error {
			var err error
			assetsByAgentVersion, err = getAssetsByAgentVersion(database)
			return err
		})

		g.Go(func() error {
			var err error
			duplicatedAssets, err = getDuplicatedAssets(database)
			return err
		})

		g.Go(func() error {
			var err error
			mostUpdatedPackages, err = getMostUpdatedPackages(database)
			return err
		})

		if err := g.Wait(); err != nil {
			logger.Error("Error loading dashboard data: " + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"Context":              c,
			"title":                "Transaction Overview",
			"statistics":           statistics,
			"totalActiveAssets":    totalActiveAssets,
			"assetsByOS":           assetsByOS,
			"assetsByAgentVersion": assetsByAgentVersion,
			"duplicatedAssets":     duplicatedAssets,
			"mostUpdatedPackages":  mostUpdatedPackages,
		})
	}
}

// getStatistics retrieves statistics records from the database.
// It takes a database connection pointer as input and returns a slice of Statistic models
// along with any error encountered during the query execution.
//
// The function queries the statistics table for name, value, percentage and updated_at fields.
// It handles NULL timestamps by using sql.NullTime and converts them to *time.Time in the returned models.
//
// Returns:
//   - []models.Statistic: Slice containing all statistics records
//   - error: Any error that occurred during database operations, nil if successful
func getStatistics(database *sql.DB) ([]models.Statistic, error) {
	rows, err := database.Query(`SELECT name, value, percentage, updated_at FROM statistics;`)

	if err != nil {
		return nil, err
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
			return nil, err
		}
		if updatedAt.Valid {
			statistic.UpdatedAt = &updatedAt.Time
		}
		statistics = append(statistics, statistic)
	}

	return statistics, nil
}

// getTotalActiveAssets retrieves the count of active assets from the database.
// It queries the assets table and returns the count of all assets where is_active = TRUE.
//
// Parameters:
//   - database: *sql.DB - The database connection to query data from
//
// Returns:
//   - int: The total number of active assets
//   - error: An error if the database query fails, nil otherwise
func getTotalActiveAssets(database *sql.DB) (int, error) {
	var count int
	err := database.QueryRow(`
		SELECT COUNT(*)
		FROM assets
		WHERE is_active = TRUE
	`).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

// getAssetsByOS retrieves statistics about the number of active machines per
// operating system from the assets table. Returns a slice of OSStats
// containing the OS name and the corresponding machine count, or an error if
// the query fails.
func getAssetsByOS(database *sql.DB) ([]OSStats, error) {
	rows, err := database.Query(`
  SELECT
    os,
    COUNT(*) AS num_machines
  FROM assets
  WHERE is_active = TRUE AND os IS NOT NULL
  GROUP BY os
  ORDER BY num_machines DESC;`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assetsByOS := []OSStats{}
	for rows.Next() {
		var stat OSStats
		if err := rows.Scan(&stat.OS, &stat.NumMachines); err != nil {
			return nil, err
		}
		assetsByOS = append(assetsByOS, stat)
	}

	return assetsByOS, nil
}

// getAssetsByAgentVersion retrieves statistics about agent versions and their
// distribution across active machines from the assets table. Returns a slice
// of AgentStats containing agent version and number of machines, or an error
// if the query fails.
// Falls back to querying executions table if agent_version column doesn't exist yet.
func getAssetsByAgentVersion(database *sql.DB) ([]AgentStats, error) {
	rows, err := database.Query(`
  SELECT
    agent_version,
    COUNT(*) AS num_machines
  FROM assets
  WHERE is_active = TRUE AND agent_version IS NOT NULL
  GROUP BY agent_version
  ORDER BY num_machines DESC;`)

	if err != nil {
		// Fallback: agent_version column may not exist yet (migration not applied)
		rows, err = database.Query(`
  SELECT
    e.agent_version,
    COUNT(DISTINCT a.hostname) AS num_machines
  FROM assets a
  INNER JOIN executions e ON e.machine_id = a.machine_id AND e.hostname = a.hostname
  WHERE a.is_active = TRUE
  AND e.executed_at = (
    SELECT MAX(e2.executed_at)
    FROM executions e2
    WHERE e2.machine_id = a.machine_id AND e2.hostname = a.hostname
  )
  GROUP BY e.agent_version
  ORDER BY num_machines DESC;`)
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()

	assetsByAgentVersion := []AgentStats{}
	for rows.Next() {
		var stat AgentStats
		if err := rows.Scan(&stat.AgentVersion, &stat.NumMachines); err != nil {
			return nil, err
		}
		assetsByAgentVersion = append(assetsByAgentVersion, stat)
	}

	return assetsByAgentVersion, nil
}

// getDuplicatedAssets queries the database to find assets (hosts) that have been
// reported from multiple distinct machines within the last 30 days.
//
// It returns a slice of DuplicatedAsset structs containing the hostname and the
// number of distinct machines that reported this hostname. The results are ordered
// by the number of machines in ascending order.
//
// The function considers only successful executions (where success = true) and
// looks for cases where the same hostname was reported from different machine_ids.
//
// Parameters:
//   - database: A pointer to sql.DB representing the database connection
//
// Returns:
//   - []DuplicatedAsset: A slice of DuplicatedAsset containing the duplicated hostnames
//   - error: An error if the database query fails, nil otherwise
//
// The function will return records only if:
//   - The hostname has been reported from more than one machine_id
//   - The second most recent execution occurred within the last 30 days
func getDuplicatedAssets(database *sql.DB) ([]DuplicatedAsset, error) {
	rows, err := database.Query(`
  SELECT
    hostname,
    COUNT(*) as num_machines
  FROM assets
  WHERE hostname IN (
    SELECT hostname
    FROM assets
    WHERE is_active = FALSE
    AND deactivated_at >= CURRENT_DATE - INTERVAL '30 day'
  )
  GROUP BY hostname
  ORDER BY num_machines DESC;`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	duplicatedAssets := []DuplicatedAsset{}
	for rows.Next() {
		var asset DuplicatedAsset
		if err := rows.Scan(&asset.Hostname, &asset.NumMachines); err != nil {
			return nil, err
		}
		duplicatedAssets = append(duplicatedAssets, asset)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return duplicatedAssets, nil
}

// getMostUpdatedPackages retrieves the top 10 most updated packages in the last 30 days from the database.
//
// The function performs a SQL query joining transaction_items and transactions tables to get:
// - Package names
// - Total number of updates for each package
// - Number of distinct hosts where each package was updated
//
// Parameters:
//   - database *sql.DB: A pointer to the SQL database connection
//
// Returns:
//   - []UpdatedPackage: A slice of UpdatedPackage structs containing the package statistics
//   - error: An error if the database query or scan operations fail, nil otherwise
//
// The results are ordered by total number of updates in descending order and limited to 10 entries.
func getMostUpdatedPackages(database *sql.DB) ([]UpdatedPackage, error) {
	rows, err := database.Query(`
  SELECT
    ti.package,
    COUNT(*) AS total_updates,
    COUNT(DISTINCT t.hostname) AS distinct_hosts_updated
  FROM
      public.transaction_items AS ti
  JOIN
      public.transactions AS t ON ti.transaction_id = t.transaction_id AND ti.machine_id = t.machine_id
  JOIN
      public.assets AS a ON t.machine_id = a.machine_id AND t.hostname = a.hostname
  WHERE
      ti.action = 'Upgrade'
      AND t.end_time >= NOW() - INTERVAL '30 days'
      AND a.is_active = TRUE
  GROUP BY
      ti.package
  ORDER BY
      total_updates DESC
  LIMIT 10;`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	updatedPackages := []UpdatedPackage{}
	for rows.Next() {
		var updatedPackage UpdatedPackage
		if err := rows.Scan(&updatedPackage.Package, &updatedPackage.TotalUpdates, &updatedPackage.DistinctHostsUpdated); err != nil {
			return nil, err
		}
		updatedPackages = append(updatedPackages, updatedPackage)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return updatedPackages, nil
}
