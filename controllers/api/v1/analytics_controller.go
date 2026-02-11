package v1

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/models"
)

// ComparePackages compares package sets between multiple assets
//
//	@summary		Compare packages between assets
//	@description	Compares the installed packages across 2-20 assets, highlighting differences
//	@tags			reports
//	@produce		json
//	@param			machine_ids	query		string	true	"Comma-separated list of machine IDs (2-20)"
//	@success		200			{object}	models.PackageComparisonResult
//	@failure		400			{object}	map[string]string	"Bad request - invalid parameters"
//	@failure		500			{object}	map[string]string	"Internal server error"
//	@router			/v1/reports/compare-packages [get]
//	@security		ApiKeyAuth
func ComparePackages(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		machineIDsStr := c.Query("machine_ids")
		if machineIDsStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "machine_ids parameter is required"})
			return
		}

		machineIDs := strings.Split(machineIDsStr, ",")
		for i := range machineIDs {
			machineIDs[i] = strings.TrimSpace(machineIDs[i])
		}

		if len(machineIDs) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least 2 machine IDs are required for comparison"})
			return
		}
		if len(machineIDs) > 20 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "maximum 20 machine IDs allowed for comparison"})
			return
		}

		result, err := comparePackageSets(database, machineIDs)
		if err != nil {
			logger.Error("Error comparing packages: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error comparing packages: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func comparePackageSets(database *sql.DB, machineIDs []string) (*models.PackageComparisonResult, error) {
	result := &models.PackageComparisonResult{
		Assets:    make([]models.AssetPackageSet, 0),
		OnlyIn:    make(map[string][]string),
		Different: make(map[string][]models.PackageVersionDiff),
		Common:    make([]string, 0),
	}

	// Get the latest package version for each machine
	query := `
		WITH LatestPackages AS (
			SELECT DISTINCT ON (ti.machine_id, ti.package)
				ti.machine_id,
				t.hostname,
				CASE
					WHEN ti.package LIKE 'Change %%' THEN SUBSTRING(ti.package FROM 8)
					ELSE ti.package
				END AS package,
				ti.version,
				ti.release,
				ti.action
			FROM transaction_items ti
			JOIN transactions t ON ti.transaction_id = t.transaction_id AND ti.machine_id = t.machine_id
			WHERE ti.machine_id = ANY($1)
			ORDER BY ti.machine_id, ti.package, t.begin_time DESC
		)
		SELECT machine_id, hostname, package, version, release
		FROM LatestPackages
		WHERE action IN ('Install', 'Upgrade', 'Downgrade')
		ORDER BY machine_id, package
	`

	rows, err := database.Query(query, pq.Array(machineIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Build package sets for each asset
	assetPackages := make(map[string]*models.AssetPackageSet)
	allPackages := make(map[string]bool)

	for rows.Next() {
		var machineID, hostname, pkg, version, release string
		if err := rows.Scan(&machineID, &hostname, &pkg, &version, &release); err != nil {
			return nil, err
		}

		if _, exists := assetPackages[machineID]; !exists {
			assetPackages[machineID] = &models.AssetPackageSet{
				MachineID: machineID,
				Hostname:  hostname,
				Packages:  make(map[string]string),
			}
		}
		assetPackages[machineID].Packages[pkg] = version + "-" + release
		allPackages[pkg] = true
	}

	// Convert map to slice
	for _, asset := range assetPackages {
		result.Assets = append(result.Assets, *asset)
	}

	// Analyze differences
	for pkg := range allPackages {
		presentIn := make([]string, 0)
		versions := make(map[string]models.PackageVersionDiff)
		versionStrings := make(map[string]string) // Store original version string for comparison

		for machineID, asset := range assetPackages {
			if ver, exists := asset.Packages[pkg]; exists {
				presentIn = append(presentIn, machineID)
				versionStrings[machineID] = ver // Store original for comparison
				parts := strings.SplitN(ver, "-", 2)
				version := parts[0]
				release := ""
				if len(parts) > 1 {
					release = parts[1]
				}
				versions[machineID] = models.PackageVersionDiff{
					MachineID: machineID,
					Hostname:  asset.Hostname,
					Version:   version,
					Release:   release,
				}
			}
		}

		// Check if package is only in some assets
		if len(presentIn) < len(machineIDs) {
			for _, machineID := range presentIn {
				result.OnlyIn[machineID] = append(result.OnlyIn[machineID], pkg)
			}
			continue
		}

		// Check if versions differ using original stored strings
		firstVersion := ""
		allSame := true
		for machineID := range versions {
			if firstVersion == "" {
				firstVersion = versionStrings[machineID]
			} else if versionStrings[machineID] != firstVersion {
				allSame = false
				break
			}
		}

		if allSame {
			result.Common = append(result.Common, pkg)
		} else {
			diffs := make([]models.PackageVersionDiff, 0)
			for _, v := range versions {
				diffs = append(diffs, v)
			}
			result.Different[pkg] = diffs
		}
	}

	return result, nil
}

// GetPackageFreshness returns package freshness analysis
//
//	@summary		Get package freshness analysis
//	@description	Returns analysis of how current the deployed package versions are
//	@tags			reports
//	@produce		json
//	@param			machine_id	query		string	false	"Optional machine ID for specific asset analysis"
//	@param			limit		query		int		false	"Number of packages to return (default 50)"
//	@success		200			{object}	models.PackageFreshnessReport
//	@failure		400			{object}	map[string]string	"Bad request - invalid parameters"
//	@failure		500			{object}	map[string]string	"Internal server error"
//	@router			/v1/reports/package-freshness [get]
//	@security		ApiKeyAuth
func GetPackageFreshness(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		machineID := c.Query("machine_id")
		limitStr := c.DefaultQuery("limit", "50")

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter (1-500)"})
			return
		}

		report, err := getPackageFreshnessReport(database, machineID, limit)
		if err != nil {
			logger.Error("Error getting package freshness: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting package freshness: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, report)
	}
}

func getPackageFreshnessReport(database *sql.DB, machineID string, limit int) (*models.PackageFreshnessReport, error) {
	report := &models.PackageFreshnessReport{
		MachineID: machineID,
		Packages:  make([]models.PackageFreshnessInfo, 0),
	}

	// Get hostname if machine_id is provided
	if machineID != "" {
		var hostname string
		err := database.QueryRow(`
			SELECT hostname FROM assets WHERE machine_id = $1 AND is_active = TRUE LIMIT 1
		`, machineID).Scan(&hostname)
		if err == nil {
			report.Hostname = hostname
		}
	}

	// Query for package freshness - using first seen time of each package version
	var query string
	var rows *sql.Rows
	var err error

	if machineID != "" {
		query = `
			WITH CurrentPackages AS (
				SELECT DISTINCT ON (ti.package)
					CASE
						WHEN ti.package LIKE 'Change %%' THEN SUBSTRING(ti.package FROM 8)
						ELSE ti.package
					END AS package,
					ti.version,
					ti.release
				FROM transaction_items ti
				JOIN transactions t ON ti.transaction_id = t.transaction_id AND ti.machine_id = t.machine_id
				WHERE ti.machine_id = $1
					AND ti.action IN ('Install', 'Upgrade', 'Downgrade')
				ORDER BY ti.package, t.begin_time DESC
			),
			PackageFirstSeen AS (
				SELECT
					cp.package,
					cp.version,
					cp.release,
					MIN(t.begin_time) as first_seen
				FROM CurrentPackages cp
				JOIN transaction_items ti ON (
					CASE
						WHEN ti.package LIKE 'Change %%' THEN SUBSTRING(ti.package FROM 8)
						ELSE ti.package
					END = cp.package
					AND ti.version = cp.version
					AND ti.release = cp.release
				)
				JOIN transactions t ON ti.transaction_id = t.transaction_id AND ti.machine_id = t.machine_id
				GROUP BY cp.package, cp.version, cp.release
			)
			SELECT
				package,
				version,
				release,
				first_seen,
				EXTRACT(DAY FROM NOW() - first_seen)::int as age_days,
				1 as asset_count
			FROM PackageFirstSeen
			ORDER BY age_days DESC
			LIMIT $2
		`
		rows, err = database.Query(query, machineID, limit)
	} else {
		query = `
			WITH LatestVersions AS (
				SELECT DISTINCT ON (
					CASE
						WHEN package LIKE 'Change %%' THEN SUBSTRING(package FROM 8)
						ELSE package
					END
				)
					CASE
						WHEN package LIKE 'Change %%' THEN SUBSTRING(package FROM 8)
						ELSE package
					END AS package,
					version,
					release
				FROM transaction_items
				ORDER BY
					CASE
						WHEN package LIKE 'Change %%' THEN SUBSTRING(package FROM 8)
						ELSE package
					END,
					version DESC, release DESC
			),
			PackageStats AS (
				SELECT
					lv.package,
					lv.version,
					lv.release,
					MIN(t.begin_time) as first_seen,
					COUNT(DISTINCT a.machine_id) as asset_count
				FROM LatestVersions lv
				JOIN transaction_items ti ON (
					CASE
						WHEN ti.package LIKE 'Change %%' THEN SUBSTRING(ti.package FROM 8)
						ELSE ti.package
					END = lv.package
					AND ti.version = lv.version
					AND ti.release = lv.release
				)
				JOIN transactions t ON ti.transaction_id = t.transaction_id AND ti.machine_id = t.machine_id
				LEFT JOIN assets a ON ti.machine_id = a.machine_id AND a.is_active = TRUE
				GROUP BY lv.package, lv.version, lv.release
			)
			SELECT
				package,
				version,
				release,
				first_seen,
				EXTRACT(DAY FROM NOW() - first_seen)::int as age_days,
				asset_count
			FROM PackageStats
			ORDER BY age_days DESC
			LIMIT $1
		`
		rows, err = database.Query(query, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totalAge int64 = 0
	var count int = 0
	var oldest, newest *models.PackageFreshnessInfo

	for rows.Next() {
		var pkg models.PackageFreshnessInfo
		if err := rows.Scan(&pkg.Package, &pkg.Version, &pkg.Release, &pkg.FirstSeenAt, &pkg.AgeInDays, &pkg.AssetCount); err != nil {
			return nil, err
		}

		report.Packages = append(report.Packages, pkg)
		totalAge += int64(pkg.AgeInDays)
		count++

		if oldest == nil || pkg.AgeInDays > oldest.AgeInDays {
			oldestCopy := pkg
			oldest = &oldestCopy
		}
		if newest == nil || pkg.AgeInDays < newest.AgeInDays {
			newestCopy := pkg
			newest = &newestCopy
		}
	}

	if count > 0 {
		report.AverageAge = float64(totalAge) / float64(count)
	}
	report.OldestPackage = oldest
	report.NewestPackage = newest

	return report, nil
}

// GetPackageAdoption returns package adoption report
//
//	@summary		Get package adoption report
//	@description	Returns how widely packages have been adopted across active assets
//	@tags			reports
//	@produce		json
//	@param			limit		query		int	false	"Number of packages to return (default 50)"
//	@param			min_assets	query		int	false	"Minimum number of assets (default 1)"
//	@success		200			{object}	models.PackageAdoptionReport
//	@failure		400			{object}	map[string]string	"Bad request - invalid parameters"
//	@failure		500			{object}	map[string]string	"Internal server error"
//	@router			/v1/reports/package-adoption [get]
//	@security		ApiKeyAuth
func GetPackageAdoption(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "50")
		minAssetsStr := c.DefaultQuery("min_assets", "1")

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter (1-500)"})
			return
		}

		minAssets, err := strconv.Atoi(minAssetsStr)
		if err != nil || minAssets < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid min_assets parameter"})
			return
		}

		report, err := getPackageAdoptionReport(database, limit, minAssets)
		if err != nil {
			logger.Error("Error getting package adoption: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting package adoption: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, report)
	}
}

func getPackageAdoptionReport(database *sql.DB, limit, minAssets int) (*models.PackageAdoptionReport, error) {
	report := &models.PackageAdoptionReport{
		Packages: make([]models.PackageAdoptionInfo, 0),
	}

	// Get total active assets
	err := database.QueryRow(`SELECT COUNT(*) FROM assets WHERE is_active = TRUE`).Scan(&report.TotalActiveAssets)
	if err != nil {
		return nil, err
	}

	// Get package adoption
	query := `
		WITH LatestPackagePerAsset AS (
			SELECT DISTINCT ON (ti.machine_id,
				CASE
					WHEN ti.package LIKE 'Change %%' THEN SUBSTRING(ti.package FROM 8)
					ELSE ti.package
				END
			)
				ti.machine_id,
				CASE
					WHEN ti.package LIKE 'Change %%' THEN SUBSTRING(ti.package FROM 8)
					ELSE ti.package
				END AS package,
				ti.action
			FROM transaction_items ti
			JOIN transactions t ON ti.transaction_id = t.transaction_id AND ti.machine_id = t.machine_id
			ORDER BY ti.machine_id,
				CASE
					WHEN ti.package LIKE 'Change %%' THEN SUBSTRING(ti.package FROM 8)
					ELSE ti.package
				END,
				t.begin_time DESC
		),
		ActivePackages AS (
			SELECT lp.machine_id, lp.package
			FROM LatestPackagePerAsset lp
			JOIN assets a ON lp.machine_id = a.machine_id AND a.is_active = TRUE
			WHERE lp.action IN ('Install', 'Upgrade', 'Downgrade')
		),
		RecentUpdates AS (
			SELECT
				CASE
					WHEN ti.package LIKE 'Change %%' THEN SUBSTRING(ti.package FROM 8)
					ELSE ti.package
				END AS package,
				COUNT(*) as update_count
			FROM transaction_items ti
			JOIN transactions t ON ti.transaction_id = t.transaction_id AND ti.machine_id = t.machine_id
			WHERE t.begin_time >= NOW() - INTERVAL '30 days'
				AND ti.action IN ('Upgrade', 'Install')
			GROUP BY
				CASE
					WHEN ti.package LIKE 'Change %%' THEN SUBSTRING(ti.package FROM 8)
					ELSE ti.package
				END
		)
		SELECT
			ap.package,
			COUNT(DISTINCT ap.machine_id) as active_assets,
			COALESCE(ru.update_count, 0) as update_count_30d
		FROM ActivePackages ap
		LEFT JOIN RecentUpdates ru ON ap.package = ru.package
		GROUP BY ap.package, ru.update_count
		HAVING COUNT(DISTINCT ap.machine_id) >= $1
		ORDER BY active_assets DESC
		LIMIT $2
	`

	rows, err := database.Query(query, minAssets, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pkg models.PackageAdoptionInfo
		if err := rows.Scan(&pkg.Package, &pkg.ActiveAssets, &pkg.UpdateCount30d); err != nil {
			return nil, err
		}
		pkg.TotalAssets = report.TotalActiveAssets
		if report.TotalActiveAssets > 0 {
			pkg.Percentage = float64(pkg.ActiveAssets) / float64(report.TotalActiveAssets) * 100
		}
		report.Packages = append(report.Packages, pkg)
	}

	return report, nil
}

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
