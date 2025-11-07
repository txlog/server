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
	// Groups by package name and counts:
	// - Total number of updates (install + upgrade)
	// - Number of unique machines affected
	query := `
		SELECT
			ti.package,
			COUNT(*) AS total_updates,
			COUNT(DISTINCT ti.machine_id) AS machine_count
		FROM
			transaction_items AS ti
		JOIN
			transactions AS t ON ti.transaction_id = t.transaction_id AND ti.machine_id = t.machine_id
		WHERE
			ti.action IN ('Install', 'Upgraded')
			AND EXTRACT(MONTH FROM t.begin_time) = $1
			AND EXTRACT(YEAR FROM t.begin_time) = $2
		GROUP BY
			ti.package
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
	csvBuilder.WriteString("Package Name,Total Updates,Servers Affected\n")

	for rows.Next() {
		var packageName string
		var totalUpdates, machineCount int

		err := rows.Scan(&packageName, &totalUpdates, &machineCount)
		if err != nil {
			return "", err
		}

		// Escape package name if it contains commas or quotes
		escapedName := packageName
		if strings.Contains(packageName, ",") || strings.Contains(packageName, "\"") {
			escapedName = fmt.Sprintf("\"%s\"", strings.ReplaceAll(packageName, "\"", "\"\""))
		}

		csvBuilder.WriteString(fmt.Sprintf("%s,%d,%d\n", escapedName, totalUpdates, machineCount))
	}

	if err := rows.Err(); err != nil {
		return "", err
	}

	return csvBuilder.String(), nil
}
