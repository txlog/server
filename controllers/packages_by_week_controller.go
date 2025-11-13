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
)

func GetPackagesByWeekIndex(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		graphData, err := getGraphData(database)
		if err != nil {
			logger.Error("Error getting statistics:" + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}

		c.HTML(http.StatusOK, "packages_by_week.html", gin.H{
			"Context":   c,
			"title":     "Packages",
			"graphData": graphData,
		})
	}
}

func getGraphData(database *sql.DB) ([]models.PackageProgression, error) {
	rows, err := database.Query(`
    SELECT
      week,
      install,
      upgraded
    FROM (
      SELECT
        DATE_TRUNC('week', t.begin_time)::DATE AS week,
        COUNT(*) FILTER (WHERE ti.action = 'Install') AS install,
        COUNT(*) FILTER (WHERE ti.action = 'Upgraded') AS upgraded
      FROM
        transaction_items AS ti
      JOIN
        transactions AS t ON ti.transaction_id = t.transaction_id AND ti.machine_id = t.machine_id
      WHERE
        ti.action IN ('Install', 'Upgraded')
      GROUP BY
        week
      ORDER BY
        week DESC
      LIMIT 15
    ) AS recent_weeks
    ORDER BY
      week ASC;`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	progressions := []models.PackageProgression{}

	for rows.Next() {
		var progression = models.PackageProgression{}
		err := rows.Scan(
			&progression.Week,
			&progression.Install,
			&progression.Upgraded,
		)

		if err != nil {
			return nil, err
		}
		progressions = append(progressions, progression)
	}

	return progressions, nil
}

// MonthlyReportData contains the data needed for generating a management report
type MonthlyReportData struct {
	CSVContent string `json:"csv_content"`
	AssetCount int    `json:"asset_count"`
}

// GetPackagesByMonth returns package update data for a specific month along with asset count
func GetPackagesByMonth(database *sql.DB) gin.HandlerFunc {
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

		csvData, err := getMonthlyPackageData(database, month, year)
		if err != nil {
			logger.Error("Error getting monthly package data: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching package data: " + err.Error()})
			return
		}

		assetCount, err := getTotalActiveAssets(database)
		if err != nil {
			logger.Error("Error getting total active assets: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching asset count: " + err.Error()})
			return
		}

		response := MonthlyReportData{
			CSVContent: csvData,
			AssetCount: assetCount,
		}

		c.JSON(http.StatusOK, response)
	}
}

func getMonthlyPackageData(database *sql.DB, month, year int) (string, error) {
	// Query to get package updates for the specified month
	// Gets OS version from executions table (most recent execution per machine)
	// Groups by OS version, package name, version, release, and arch
	// Counts:
	// - Number of unique machines that received the update
	// - Total number of update transactions (some packages may be updated multiple times)
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
			ti.package,
			ti.version,
			ti.release,
			ti.arch,
			COUNT(DISTINCT ti.machine_id) AS machine_count,
			COUNT(*) AS total_updates
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
			machine_count DESC, total_updates DESC
		LIMIT 500
	`

	rows, err := database.Query(query, month, year)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// Build CSV content
	var csvBuilder strings.Builder
	csvBuilder.WriteString("OS Version,Package RPM,Servers Affected,Total Updates\n")

	for rows.Next() {
		var osVersion, packageName, version, release, arch string
		var machineCount, totalUpdates int

		err := rows.Scan(&osVersion, &packageName, &version, &release, &arch, &machineCount, &totalUpdates)
		if err != nil {
			return "", err
		}

		// Build full RPM package name: package-version-release.arch
		fullPackageName := fmt.Sprintf("%s-%s-%s.%s", packageName, version, release, arch)

		// Escape fields if they contain commas or quotes
		escapedOS := escapeCSVField(osVersion)
		escapedPackage := escapeCSVField(fullPackageName)

		csvBuilder.WriteString(fmt.Sprintf("%s,%s,%d,%d\n", escapedOS, escapedPackage, machineCount, totalUpdates))
	}

	if err := rows.Err(); err != nil {
		return "", err
	}

	return csvBuilder.String(), nil
}

// escapeCSVField escapes a CSV field if it contains commas, quotes, or newlines
func escapeCSVField(field string) string {
	if strings.ContainsAny(field, ",\"\n") {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(field, "\"", "\"\""))
	}
	return field
}
