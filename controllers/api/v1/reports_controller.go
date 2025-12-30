package v1

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	logger "github.com/txlog/server/logger"
)

// MonthlyReportPackage represents a package update entry in the monthly report
type MonthlyReportPackage struct {
	OSVersion      string `json:"os_version"`
	PackageRPM     string `json:"package_rpm"`
	AssetsAffected int    `json:"assets_affected"`
}

// MonthlyReportResponse represents the response for the monthly report endpoint
type MonthlyReportResponse struct {
	Month      int                    `json:"month"`
	Year       int                    `json:"year"`
	AssetCount int                    `json:"asset_count"`
	Packages   []MonthlyReportPackage `json:"packages"`
}

// GetMonthlyReport Get monthly package update report for management
// @summary		Get monthly package update report
// @description	Returns a list of package updates for a specific month/year, including the number of assets in the period, OS version, package RPM, servers affected, and total updates
// @tags			reports
// @produce		json
// @param			month	query		int	true	"Month (1-12)"
// @param			year	query		int	true	"Year (e.g., 2024)"
// @success		200		{object}	MonthlyReportResponse
// @failure		400		{object}	map[string]string	"Bad request - invalid parameters"
// @failure		500		{object}	map[string]string	"Internal server error"
// @router			/v1/reports/monthly [get]
// @security		ApiKeyAuth
func GetMonthlyReport(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		monthStr := c.Query("month")
		yearStr := c.Query("year")

		if monthStr == "" || yearStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "month and year parameters are required"})
			return
		}

		month, err := strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month parameter"})
			return
		}

		year, err := strconv.Atoi(yearStr)
		if err != nil || year < 2000 || year > 2100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year parameter"})
			return
		}

		packages, err := getMonthlyPackageReport(database, month, year)
		if err != nil {
			logger.Error("Error getting monthly package report: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching package data: " + err.Error()})
			return
		}

		assetCount, err := getTotalActiveAssetsForReport(database)
		if err != nil {
			logger.Error("Error getting total active assets: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching asset count: " + err.Error()})
			return
		}

		response := MonthlyReportResponse{
			Month:      month,
			Year:       year,
			AssetCount: assetCount,
			Packages:   packages,
		}

		c.JSON(http.StatusOK, response)
	}
}

// getMonthlyPackageReport retrieves package update data for a specific month
func getMonthlyPackageReport(database *sql.DB, month, year int) ([]MonthlyReportPackage, error) {
	query := `
		WITH latest_executions AS (
			SELECT DISTINCT ON (machine_id)
				machine_id,
				os
			FROM executions
			WHERE os IS NOT NULL AND os != ''
			ORDER BY machine_id, executed_at DESC
		)
		SELECT
			COALESCE(le.os, 'Unknown OS') AS os_version,
			CONCAT(ti.package, '-', ti.version, '-', ti.release, '.', ti.arch) AS package_rpm,
			COUNT(DISTINCT ti.machine_id) AS assets_affected
		FROM
			transaction_items AS ti
		JOIN
			transactions AS t ON ti.transaction_id = t.transaction_id AND ti.machine_id = t.machine_id
		LEFT JOIN
			latest_executions AS le ON ti.machine_id = le.machine_id
		WHERE
			ti.action IN ('Install', 'Upgraded')
			AND EXTRACT(MONTH FROM t.begin_time) = $1
			AND EXTRACT(YEAR FROM t.begin_time) = $2
		GROUP BY
			le.os, ti.package, ti.version, ti.release, ti.arch
		ORDER BY
			assets_affected DESC
		LIMIT 500
	`

	rows, err := database.Query(query, month, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	packages := []MonthlyReportPackage{}

	for rows.Next() {
		var pkg MonthlyReportPackage
		err := rows.Scan(&pkg.OSVersion, &pkg.PackageRPM, &pkg.AssetsAffected)
		if err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return packages, nil
}

// getTotalActiveAssetsForReport retrieves the count of active assets from the database
func getTotalActiveAssetsForReport(database *sql.DB) (int, error) {
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
