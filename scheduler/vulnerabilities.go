package scheduler

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/txlog/server/database"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/util"
)

func UpdateVulnerabilitiesJob() {
	logger.Info("Vulnerabilities: executing update task...")

	lockName := "vulnerabilities"

	locked, err := acquireLock(lockName)
	if err != nil {
		logger.Error("Error acquiring lock for vulnerabilities: " + err.Error())
		return
	}
	if !locked {
		logger.Info("Another instance is running this vulnerabilities job.")
		return
	}
	defer releaseLock(lockName)

	// Extrai todos os pacotes das transações recentes ou do ecossistema geral.
	query := `
        SELECT DISTINCT ti.package, ti.version, a.os
        FROM transaction_items ti
        JOIN transactions t ON ti.transaction_id = t.transaction_id AND ti.machine_id = t.machine_id
        JOIN assets a ON t.machine_id = a.machine_id AND t.hostname = a.hostname
        WHERE ti.action IN ('Install', 'Upgrade', 'Downgrade', 'Reinstall', 'installed', 'upgrade',
                             'Removed', 'Upgraded', 'Downgraded', 'Obsoleted', 'removed')
    `
	rows, err := database.Db.Query(query)
	if err != nil {
		logger.Error("Vulnerabilities: " + err.Error())
		return
	}
	defer rows.Close()

	type pkg struct {
		Name      string
		Version   string
		Ecosystem string
	}
	var packages []pkg

	pkgMap := make(map[string]pkg)

	for rows.Next() {
		var pName, pVersion, pOs sql.NullString
		if err := rows.Scan(&pName, &pVersion, &pOs); err != nil {
			logger.Error("Vulnerabilities scan error: " + err.Error())
			continue
		}

		if !pName.Valid || !pVersion.Valid {
			continue
		}

		ecos := util.ExtractOSVEcosystems(pOs.String)
		for _, eco := range ecos {
			key := fmt.Sprintf("%s|%s|%s", pName.String, pVersion.String, eco)
			pkgMap[key] = pkg{Name: pName.String, Version: pVersion.String, Ecosystem: eco}
		}
	}

	for _, v := range pkgMap {
		packages = append(packages, v)
	}

	logger.Info(fmt.Sprintf("Vulnerabilities: found %d discrete package/ecosystem pairs to check.", len(packages)))

	// Caches the detailed vulnerability data
	fetchedVulns := make(map[string]*util.OSVVuln)

	chunkSize := 500
	for i := 0; i < len(packages); i += chunkSize {
		end := i + chunkSize
		if end > len(packages) {
			end = len(packages)
		}

		logger.Info(fmt.Sprintf("Vulnerabilities: fetching %d to %d of %d...", i+1, end, len(packages)))

		chunk := packages[i:end]
		var osvQueries []util.OSVQuery

		for _, p := range chunk {
			osvQueries = append(osvQueries, util.OSVQuery{
				Package: util.OSVPackage{Name: p.Name, Ecosystem: p.Ecosystem},
				Version: p.Version,
			})
		}

		resp, err := util.FetchOSVVulnerabilitiesBatch(osvQueries)
		if err != nil {
			logger.Error("Vulnerabilities fetch error: " + err.Error())
			continue
		}

		for j, result := range resp.Results {
			targetPkg := chunk[j]
			if len(result.Vulns) > 0 {
				for _, batchVuln := range result.Vulns {
					vuln, found := fetchedVulns[batchVuln.ID]
					if !found {
						fetched, err := util.FetchOSVVulnerabilityDetails(batchVuln.ID)
						if err == nil && fetched != nil {
							vuln = fetched
							fetchedVulns[batchVuln.ID] = vuln
						} else {
							vuln = &batchVuln // fallback
						}
					}

					var modifiedAt, publishedAt *time.Time
					if !vuln.ModifiedAt.IsZero() {
						modifiedAt = &vuln.ModifiedAt
					}
					if !vuln.Published.IsZero() {
						publishedAt = &vuln.Published
					}

					// Infer severity loosely if possible, OSV does not always explicitly provide top-level string.
					severity := "UNKNOWN"
					cvssScore := 0.0

					summaryUpper := strings.ToUpper(vuln.Summary)
					detailsUpper := strings.ToUpper(vuln.Details)

					if strings.Contains(summaryUpper, "CRITICAL") || strings.Contains(detailsUpper, "CRITICAL") {
						severity = "CRITICAL"
						cvssScore = 9.5
					} else if strings.Contains(summaryUpper, "IMPORTANT") || strings.Contains(summaryUpper, "HIGH") || strings.Contains(detailsUpper, "HIGH") {
						severity = "HIGH"
						cvssScore = 8.0
					} else if strings.Contains(summaryUpper, "MODERATE") || strings.Contains(summaryUpper, "MEDIUM") || strings.Contains(detailsUpper, "MEDIUM") {
						severity = "MEDIUM"
						cvssScore = 5.5
					} else if strings.Contains(summaryUpper, "LOW") || strings.Contains(detailsUpper, "LOW") {
						severity = "LOW"
						cvssScore = 3.0
					}

					_, err := database.Db.Exec(`
						INSERT INTO vulnerabilities (id, summary, details, severity, cvss_score, modified_at, published_at)
						VALUES ($1, $2, $3, $4, $5, $6, $7)
						ON CONFLICT (id) DO UPDATE SET
							summary = EXCLUDED.summary,
							details = EXCLUDED.details,
							severity = EXCLUDED.severity,
							cvss_score = EXCLUDED.cvss_score,
							modified_at = EXCLUDED.modified_at
					`, vuln.ID, vuln.Summary, vuln.Details, severity, cvssScore, modifiedAt, publishedAt)
					if err != nil {
						logger.Error("Error inserting vulnerability: " + err.Error())
						continue
					}

					if targetPkg.Ecosystem != "" {
						_, err = database.Db.Exec(`
							INSERT INTO package_vulnerabilities (package_name, version, vulnerability_id, ecosystem)
							VALUES ($1, $2, $3, $4)
							ON CONFLICT DO NOTHING
						`, targetPkg.Name, targetPkg.Version, vuln.ID, targetPkg.Ecosystem)
						if err != nil {
							logger.Error("Error linking pkg to vulnerability: " + err.Error())
						}
					}
				}
			}
		}
	}

	logger.Info("Vulnerabilities downloaded. Proceeding to calculate transaction scoreboards...")
	updateTransactionScoreboards()
	logger.Info("Vulnerabilities and transaction scoreboards updated successfully.")
}

func updateTransactionScoreboards() {
	// Puxa transações pendentes de serem "rated" com o Scoreboard.
	// Como a gente não adicionou uma flag processada, apenas rodamos um batch
	// atualizando quem não teve scoreboard injetado. Para evitar sobrecarga processamos em CTE ou batch.

	// A strategy: execute an UPDATE statement that uses the delta
	// of package_vulnerabilities from "removed" to "installed/upgrade" items of the same transaction.

	stmt := `
WITH removed_vulns AS (
    SELECT
        ti.transaction_id,
        COUNT(DISTINCT pv.vulnerability_id) as total_removed,
        COUNT(DISTINCT CASE WHEN v.severity = 'CRITICAL' THEN pv.vulnerability_id END) as critical_removed,
        COUNT(DISTINCT CASE WHEN v.severity = 'HIGH' THEN pv.vulnerability_id END) as high_removed,
        COUNT(DISTINCT CASE WHEN v.severity = 'MEDIUM' THEN pv.vulnerability_id END) as medium_removed,
        COUNT(DISTINCT CASE WHEN v.severity = 'LOW' THEN pv.vulnerability_id END) as low_removed,
        SUM(DISTINCT v.cvss_score) as removed_cvss
    FROM transaction_items ti
    JOIN transactions trans ON ti.transaction_id = trans.transaction_id AND ti.machine_id = trans.machine_id
    JOIN assets a ON trans.machine_id = a.machine_id AND trans.hostname = a.hostname
    JOIN package_vulnerabilities pv ON pv.package_name = ti.package AND pv.version = ti.version
         AND (
             (a.os ILIKE '%AlmaLinux%' AND pv.ecosystem = 'AlmaLinux:' || SUBSTRING(a.os FROM '[0-9]+')) OR
             (a.os ILIKE '%Rocky%' AND pv.ecosystem = 'Rocky Linux:' || SUBSTRING(a.os FROM '[0-9]+')) OR
             ((a.os ILIKE '%Red Hat%' OR a.os ILIKE '%RHEL%' OR a.os ILIKE '%CentOS%' OR a.os ILIKE '%Oracle%') AND pv.ecosystem = 'AlmaLinux:' || SUBSTRING(a.os FROM '[0-9]+'))
         )
    JOIN vulnerabilities v ON v.id = pv.vulnerability_id
    WHERE ti.action IN ('Removed', 'Obsoleted', 'Upgraded', 'Downgraded', 'removed')
    GROUP BY ti.transaction_id
),
installed_vulns AS (
    SELECT
        ti.transaction_id,
        COUNT(DISTINCT pv.vulnerability_id) as total_installed,
        COUNT(DISTINCT CASE WHEN v.severity = 'CRITICAL' THEN pv.vulnerability_id END) as critical_installed,
        COUNT(DISTINCT CASE WHEN v.severity = 'HIGH' THEN pv.vulnerability_id END) as high_installed,
        COUNT(DISTINCT CASE WHEN v.severity = 'MEDIUM' THEN pv.vulnerability_id END) as medium_installed,
        COUNT(DISTINCT CASE WHEN v.severity = 'LOW' THEN pv.vulnerability_id END) as low_installed,
        SUM(DISTINCT v.cvss_score) as installed_cvss
    FROM transaction_items ti
    JOIN transactions trans ON ti.transaction_id = trans.transaction_id AND ti.machine_id = trans.machine_id
    JOIN assets a ON trans.machine_id = a.machine_id AND trans.hostname = a.hostname
    JOIN package_vulnerabilities pv ON pv.package_name = ti.package AND pv.version = ti.version
         AND (
             (a.os ILIKE '%AlmaLinux%' AND pv.ecosystem = 'AlmaLinux:' || SUBSTRING(a.os FROM '[0-9]+')) OR
             (a.os ILIKE '%Rocky%' AND pv.ecosystem = 'Rocky Linux:' || SUBSTRING(a.os FROM '[0-9]+')) OR
             ((a.os ILIKE '%Red Hat%' OR a.os ILIKE '%RHEL%' OR a.os ILIKE '%CentOS%' OR a.os ILIKE '%Oracle%') AND pv.ecosystem = 'AlmaLinux:' || SUBSTRING(a.os FROM '[0-9]+'))
         )
    JOIN vulnerabilities v ON v.id = pv.vulnerability_id
    WHERE ti.action IN ('Install', 'Upgrade', 'Downgrade', 'Reinstall', 'installed', 'upgrade')
    GROUP BY ti.transaction_id
)
UPDATE transactions t
SET
    vulns_fixed = GREATEST(COALESCE(r.total_removed, 0) - COALESCE(i.total_installed, 0), 0),
    vulns_introduced = GREATEST(COALESCE(i.total_installed, 0) - COALESCE(r.total_removed, 0), 0),
    critical_vulns_fixed = GREATEST(COALESCE(r.critical_removed, 0) - COALESCE(i.critical_installed, 0), 0),
    critical_vulns_introduced = GREATEST(COALESCE(i.critical_installed, 0) - COALESCE(r.critical_removed, 0), 0),
    high_vulns_fixed = GREATEST(COALESCE(r.high_removed, 0) - COALESCE(i.high_installed, 0), 0),
    high_vulns_introduced = GREATEST(COALESCE(i.high_installed, 0) - COALESCE(r.high_removed, 0), 0),
    medium_vulns_fixed = GREATEST(COALESCE(r.medium_removed, 0) - COALESCE(i.medium_installed, 0), 0),
    medium_vulns_introduced = GREATEST(COALESCE(i.medium_installed, 0) - COALESCE(r.medium_removed, 0), 0),
    low_vulns_fixed = GREATEST(COALESCE(r.low_removed, 0) - COALESCE(i.low_installed, 0), 0),
    low_vulns_introduced = GREATEST(COALESCE(i.low_installed, 0) - COALESCE(r.low_removed, 0), 0),
    risk_score_mitigated = GREATEST(COALESCE(r.removed_cvss, 0) - COALESCE(i.installed_cvss, 0), 0),
    is_security_patch = (GREATEST(COALESCE(r.critical_removed, 0) - COALESCE(i.critical_installed, 0), 0) > 0 OR
                         GREATEST(COALESCE(r.high_removed, 0) - COALESCE(i.high_installed, 0), 0) > 0 OR
                         GREATEST(COALESCE(r.total_removed, 0) - COALESCE(i.total_installed, 0), 0) > 0)
FROM removed_vulns r
FULL OUTER JOIN installed_vulns i ON r.transaction_id = i.transaction_id
WHERE t.transaction_id = COALESCE(r.transaction_id, i.transaction_id)
  AND (COALESCE(r.total_removed, 0) > 0 OR COALESCE(i.total_installed, 0) > 0);
	`

	_, err := database.Db.Exec(stmt)
	if err != nil {
		logger.Error("Failed to update transaction scoreboards: " + err.Error())
	}
}
