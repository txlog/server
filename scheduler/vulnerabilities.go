package scheduler

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
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
        SELECT DISTINCT ti.package, ti.version, COALESCE(ti.release, '') AS release, a.os
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
		Release   string
		Ecosystem string
	}
	var packages []pkg

	pkgMap := make(map[string]pkg)

	for rows.Next() {
		var pName, pVersion, pRelease, pOs sql.NullString
		if err := rows.Scan(&pName, &pVersion, &pRelease, &pOs); err != nil {
			logger.Error("Vulnerabilities scan error: " + err.Error())
			continue
		}

		if !pName.Valid || !pVersion.Valid {
			continue
		}

		ecos := util.ExtractOSVEcosystems(pOs.String)
		for _, eco := range ecos {
			key := fmt.Sprintf("%s|%s|%s|%s", pName.String, pVersion.String, pRelease.String, eco)
			pkgMap[key] = pkg{Name: pName.String, Version: pVersion.String, Release: pRelease.String, Ecosystem: eco}
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
			fullVersion := p.Version
			if p.Release != "" {
				fullVersion = fullVersion + "-" + p.Release
			}
			osvQueries = append(osvQueries, util.OSVQuery{
				Package: util.OSVPackage{Name: p.Name, Ecosystem: p.Ecosystem},
				Version: fullVersion,
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
							INSERT INTO package_vulnerabilities (package_name, version, release, vulnerability_id, ecosystem)
							VALUES ($1, $2, $3, $4, $5)
							ON CONFLICT DO NOTHING
						`, targetPkg.Name, targetPkg.Version, targetPkg.Release, vuln.ID, targetPkg.Ecosystem)
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
	logger.Info("Vulnerabilities: Fetching transactions for scoreboard calculation...")

	type txKey struct {
		TransactionID string
		MachineID     string
	}
	var keys []txKey

	rows, err := database.Db.Query("SELECT DISTINCT transaction_id, machine_id FROM transactions")
	if err != nil {
		logger.Error("Failed to fetch transactions list: " + err.Error())
		return
	}
	for rows.Next() {
		var k txKey
		if err := rows.Scan(&k.TransactionID, &k.MachineID); err == nil {
			keys = append(keys, k)
		}
	}
	rows.Close()

	total := len(keys)
	logger.Info(fmt.Sprintf("Vulnerabilities: Total of %d transactions to process.", total))

	chunkSize := 500
	for i := 0; i < total; i += chunkSize {
		end := i + chunkSize
		if end > total {
			end = total
		}

		pct := float64(i) / float64(total) * 100
		logger.Info(fmt.Sprintf("Vulnerabilities: Processing batch %d to %d of %d (%.1f%%)...", i+1, end, total, pct))

		chunk := keys[i:end]
		var txnIDs []string
		var mchnIDs []string
		for _, k := range chunk {
			txnIDs = append(txnIDs, k.TransactionID)
			mchnIDs = append(mchnIDs, k.MachineID)
		}

		stmt := `
WITH batch AS (
    SELECT unnest($1::text[])::integer AS transaction_id, unnest($2::text[]) AS machine_id
),
raw_removed AS (
    SELECT
        ti.transaction_id,
        ti.machine_id,
        pv.vulnerability_id,
        v.severity,
        v.cvss_score
    FROM transaction_items ti
    JOIN batch b ON ti.transaction_id = b.transaction_id AND ti.machine_id = b.machine_id
    JOIN transactions trans ON ti.transaction_id = trans.transaction_id AND ti.machine_id = trans.machine_id
    JOIN assets a ON trans.machine_id = a.machine_id AND trans.hostname = a.hostname
    JOIN package_vulnerabilities pv ON pv.package_name = ti.package AND pv.version = ti.version
         AND pv.release = COALESCE(ti.release, '')
         AND (
             (a.os ILIKE '%AlmaLinux%' AND pv.ecosystem = 'AlmaLinux:' || SUBSTRING(a.os FROM '[0-9]+')) OR
             (a.os ILIKE '%Rocky%' AND pv.ecosystem = 'Rocky Linux:' || SUBSTRING(a.os FROM '[0-9]+')) OR
             ((a.os ILIKE '%Red Hat%' OR a.os ILIKE '%RHEL%' OR a.os ILIKE '%CentOS%' OR a.os ILIKE '%Oracle%') AND pv.ecosystem = 'AlmaLinux:' || SUBSTRING(a.os FROM '[0-9]+'))
         )
    JOIN vulnerabilities v ON v.id = pv.vulnerability_id
    WHERE ti.action IN ('Removed', 'Obsoleted', 'Upgraded', 'Downgraded', 'removed')
),
raw_installed AS (
    SELECT
        ti.transaction_id,
        ti.machine_id,
        pv.vulnerability_id,
        v.severity,
        v.cvss_score
    FROM transaction_items ti
    JOIN batch b ON ti.transaction_id = b.transaction_id AND ti.machine_id = b.machine_id
    JOIN transactions trans ON ti.transaction_id = trans.transaction_id AND ti.machine_id = trans.machine_id
    JOIN assets a ON trans.machine_id = a.machine_id AND trans.hostname = a.hostname
    JOIN package_vulnerabilities pv ON pv.package_name = ti.package AND pv.version = ti.version
         AND pv.release = COALESCE(ti.release, '')
         AND (
             (a.os ILIKE '%AlmaLinux%' AND pv.ecosystem = 'AlmaLinux:' || SUBSTRING(a.os FROM '[0-9]+')) OR
             (a.os ILIKE '%Rocky%' AND pv.ecosystem = 'Rocky Linux:' || SUBSTRING(a.os FROM '[0-9]+')) OR
             ((a.os ILIKE '%Red Hat%' OR a.os ILIKE '%RHEL%' OR a.os ILIKE '%CentOS%' OR a.os ILIKE '%Oracle%') AND pv.ecosystem = 'AlmaLinux:' || SUBSTRING(a.os FROM '[0-9]+'))
         )
    JOIN vulnerabilities v ON v.id = pv.vulnerability_id
    WHERE ti.action IN ('Install', 'Upgrade', 'Downgrade', 'Reinstall', 'installed', 'upgrade')
),
removed_vulns AS (
    SELECT
        transaction_id,
        machine_id,
        COUNT(DISTINCT vulnerability_id) as total_removed,
        COUNT(DISTINCT CASE WHEN severity = 'CRITICAL' THEN vulnerability_id END) as critical_removed,
        COUNT(DISTINCT CASE WHEN severity = 'HIGH' THEN vulnerability_id END) as high_removed,
        COUNT(DISTINCT CASE WHEN severity = 'MEDIUM' THEN vulnerability_id END) as medium_removed,
        COUNT(DISTINCT CASE WHEN severity = 'LOW' THEN vulnerability_id END) as low_removed,
        COALESCE(SUM(cvss_score), 0) as removed_cvss
    FROM (
        SELECT DISTINCT transaction_id, machine_id, vulnerability_id, severity, cvss_score
        FROM raw_removed r
        WHERE NOT EXISTS (
            SELECT 1 FROM raw_installed i
            WHERE i.transaction_id = r.transaction_id AND i.machine_id = r.machine_id AND i.vulnerability_id = r.vulnerability_id
        )
    ) as truly_fixed
    GROUP BY transaction_id, machine_id
),
installed_vulns AS (
    SELECT
        transaction_id,
        machine_id,
        COUNT(DISTINCT vulnerability_id) as total_installed,
        COUNT(DISTINCT CASE WHEN severity = 'CRITICAL' THEN vulnerability_id END) as critical_installed,
        COUNT(DISTINCT CASE WHEN severity = 'HIGH' THEN vulnerability_id END) as high_installed,
        COUNT(DISTINCT CASE WHEN severity = 'MEDIUM' THEN vulnerability_id END) as medium_installed,
        COUNT(DISTINCT CASE WHEN severity = 'LOW' THEN vulnerability_id END) as low_installed,
        COALESCE(SUM(cvss_score), 0) as installed_cvss
    FROM (
        SELECT DISTINCT transaction_id, machine_id, vulnerability_id, severity, cvss_score
        FROM raw_installed i
        WHERE NOT EXISTS (
            SELECT 1 FROM raw_removed r
            WHERE r.transaction_id = i.transaction_id AND r.machine_id = i.machine_id AND r.vulnerability_id = i.vulnerability_id
        )
    ) as truly_introduced
    GROUP BY transaction_id, machine_id
)
UPDATE transactions t
SET
    vulns_fixed = COALESCE(r.total_removed, 0),
    vulns_introduced = COALESCE(i.total_installed, 0),
    critical_vulns_fixed = COALESCE(r.critical_removed, 0),
    critical_vulns_introduced = COALESCE(i.critical_installed, 0),
    high_vulns_fixed = COALESCE(r.high_removed, 0),
    high_vulns_introduced = COALESCE(i.high_installed, 0),
    medium_vulns_fixed = COALESCE(r.medium_removed, 0),
    medium_vulns_introduced = COALESCE(i.medium_installed, 0),
    low_vulns_fixed = COALESCE(r.low_removed, 0),
    low_vulns_introduced = COALESCE(i.low_installed, 0),
    risk_score_mitigated = COALESCE(r.removed_cvss, 0) - COALESCE(i.installed_cvss, 0),
    is_security_patch = (COALESCE(r.critical_removed, 0) > 0 OR COALESCE(r.high_removed, 0) > 0 OR COALESCE(r.total_removed, 0) > 0)
FROM removed_vulns r
FULL OUTER JOIN installed_vulns i ON r.transaction_id = i.transaction_id AND r.machine_id = i.machine_id
WHERE t.transaction_id = COALESCE(r.transaction_id, i.transaction_id)
  AND t.machine_id = COALESCE(r.machine_id, i.machine_id)
  AND (COALESCE(r.total_removed, 0) > 0 OR COALESCE(i.total_installed, 0) > 0);
		`

		_, err := database.Db.Exec(stmt, pq.Array(txnIDs), pq.Array(mchnIDs))
		if err != nil {
			logger.Error("Failed to update transaction scoreboards for batch: " + err.Error())
		}
	}
}
