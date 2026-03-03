package v1

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/models"
)

// GetAnomalies returns detected transaction anomalies
//
//	@summary		Get transaction anomalies
//	@description	Detects unusual transaction patterns like high volume, rapid changes, or downgrades
//	@tags			reports
//	@produce		json
//	@param			days		query		int		false	"Number of days to analyze (default 7)"
//	@param			severity	query		string	false	"Minimum severity filter (low, medium, high)"
//	@success		200			{object}	models.AnomalyReport
//	@failure		400			{object}	map[string]string	"Bad request - invalid parameters"
//	@failure		500			{object}	map[string]string	"Internal server error"
//	@router			/v1/reports/anomalies [get]
//	@security		ApiKeyAuth
func GetAnomalies(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		daysStr := c.DefaultQuery("days", "7")
		severityFilter := c.DefaultQuery("severity", "")

		days, err := strconv.Atoi(daysStr)
		if err != nil || days < 1 || days > 90 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid days parameter (1-90)"})
			return
		}

		if severityFilter != "" && severityFilter != "low" && severityFilter != "medium" && severityFilter != "high" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid severity parameter (low, medium, high)"})
			return
		}

		report, err := detectAnomalies(database, days, severityFilter)
		if err != nil {
			logger.Error("Error detecting anomalies: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error detecting anomalies: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, report)
	}
}

func detectAnomalies(database *sql.DB, days int, severityFilter string) (*models.AnomalyReport, error) {
	report := &models.AnomalyReport{
		TimeWindow: fmt.Sprintf("%d days", days),
		Anomalies:  make([]models.TransactionAnomaly, 0),
		Summary: models.AnomalySummary{
			ByType:     make(map[string]int),
			BySeverity: make(map[string]int),
		},
	}

	affectedHosts := make(map[string]bool)

	// Detect high volume transactions (>50 packages in single transaction)
	highVolumeQuery := `
		SELECT
			t.transaction_id,
			t.machine_id,
			t.hostname,
			t.begin_time,
			COUNT(*) as package_count
		FROM transactions t
		JOIN transaction_items ti ON t.transaction_id = ti.transaction_id AND t.machine_id = ti.machine_id
		WHERE t.begin_time >= NOW() - make_interval(days => $1)
		GROUP BY t.transaction_id, t.machine_id, t.hostname, t.begin_time
		HAVING COUNT(*) > 50
		ORDER BY package_count DESC
	`

	rows, err := database.Query(highVolumeQuery, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var txID int
		var machineID, hostname string
		var beginTime time.Time
		var packageCount int

		if err := rows.Scan(&txID, &machineID, &hostname, &beginTime, &packageCount); err != nil {
			return nil, err
		}

		severity := models.SeverityMedium
		if packageCount > 100 {
			severity = models.SeverityHigh
		}

		if severityFilter != "" && string(severity) != severityFilter {
			continue
		}

		anomaly := models.TransactionAnomaly{
			Type:        models.AnomalyHighVolume,
			MachineID:   machineID,
			Hostname:    hostname,
			DetectedAt:  beginTime,
			Description: fmt.Sprintf("Transaction with %d packages (threshold: 50)", packageCount),
			Severity:    severity,
			Details: models.HighVolumeDetails{
				TransactionID:   txID,
				PackageCount:    packageCount,
				TransactionTime: beginTime.Format(time.RFC3339),
			},
		}

		report.Anomalies = append(report.Anomalies, anomaly)
		affectedHosts[machineID] = true
	}

	// Detect rapid changes (same package changed >3 times in 24h)
	rapidChangeQuery := `
		WITH PackageChanges AS (
			SELECT
				ti.machine_id,
				t.hostname,
				CASE
					WHEN ti.package LIKE 'Change %' THEN SUBSTRING(ti.package FROM 8)
					ELSE ti.package
				END AS package,
				ti.action,
				t.begin_time,
				DATE_TRUNC('day', t.begin_time) as change_day
			FROM transaction_items ti
			JOIN transactions t ON ti.transaction_id = t.transaction_id AND ti.machine_id = t.machine_id
			WHERE t.begin_time >= NOW() - make_interval(days => $1)
				AND ti.action IN ('Install', 'Upgrade', 'Downgrade', 'Erase', 'Reinstall')
		)
		SELECT
			machine_id,
			hostname,
			package,
			change_day,
			COUNT(*) as change_count,
			ARRAY_AGG(DISTINCT action) as actions
		FROM PackageChanges
		GROUP BY machine_id, hostname, package, change_day
		HAVING COUNT(*) > 3
		ORDER BY change_count DESC
	`

	rows2, err := database.Query(rapidChangeQuery, days)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()

	for rows2.Next() {
		var machineID, hostname, pkg string
		var changeDay time.Time
		var changeCount int
		var actions []string

		if err := rows2.Scan(&machineID, &hostname, &pkg, &changeDay, &changeCount, &actions); err != nil {
			// Handle array scanning issue
			continue
		}

		severity := models.SeverityMedium
		if changeCount > 5 {
			severity = models.SeverityHigh
		}

		if severityFilter != "" && string(severity) != severityFilter {
			continue
		}

		anomaly := models.TransactionAnomaly{
			Type:        models.AnomalyRapidChange,
			MachineID:   machineID,
			Hostname:    hostname,
			DetectedAt:  changeDay,
			Description: fmt.Sprintf("Package '%s' changed %d times in 24h", pkg, changeCount),
			Severity:    severity,
			Details: models.RapidChangeDetails{
				Package:     pkg,
				ChangeCount: changeCount,
				TimeWindow:  "24 hours",
				Actions:     actions,
			},
		}

		report.Anomalies = append(report.Anomalies, anomaly)
		affectedHosts[machineID] = true
	}

	// Detect downgrades
	downgradeQuery := `
		SELECT
			ti.machine_id,
			t.hostname,
			CASE
				WHEN ti.package LIKE 'Change %' THEN SUBSTRING(ti.package FROM 8)
				ELSE ti.package
			END AS package,
			ti.version as to_version,
			t.begin_time
		FROM transaction_items ti
		JOIN transactions t ON ti.transaction_id = t.transaction_id AND ti.machine_id = t.machine_id
		WHERE t.begin_time >= NOW() - make_interval(days => $1)
			AND ti.action = 'Downgrade'
		ORDER BY t.begin_time DESC
	`

	rows3, err := database.Query(downgradeQuery, days)
	if err != nil {
		return nil, err
	}
	defer rows3.Close()

	for rows3.Next() {
		var machineID, hostname, pkg, toVersion string
		var beginTime time.Time

		if err := rows3.Scan(&machineID, &hostname, &pkg, &toVersion, &beginTime); err != nil {
			return nil, err
		}

		severity := models.SeverityLow
		if severityFilter != "" && string(severity) != severityFilter {
			continue
		}

		anomaly := models.TransactionAnomaly{
			Type:        models.AnomalyDowngrade,
			MachineID:   machineID,
			Hostname:    hostname,
			DetectedAt:  beginTime,
			Description: fmt.Sprintf("Package '%s' was downgraded to version %s", pkg, toVersion),
			Severity:    severity,
			Details: models.DowngradeDetails{
				Package:   pkg,
				ToVersion: toVersion,
			},
		}

		report.Anomalies = append(report.Anomalies, anomaly)
		affectedHosts[machineID] = true
	}

	// Calculate summary
	for _, anomaly := range report.Anomalies {
		report.Summary.TotalCount++
		report.Summary.ByType[string(anomaly.Type)]++
		report.Summary.BySeverity[string(anomaly.Severity)]++
	}
	report.Summary.AffectedHosts = len(affectedHosts)

	return report, nil
}
