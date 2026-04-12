package scheduler

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/util"
)

// Package-level types used across multiple functions in the vulnerability pipeline.

type vulnRecord struct {
	ID          string
	Summary     string
	Details     string
	Severity    string
	CVSSScore   float64
	ModifiedAt  *time.Time
	PublishedAt *time.Time
}

type pvRecord struct {
	PackageName     string
	Version         string
	Release         string
	VulnerabilityID string
	Ecosystem       string
}

type vulnPkgKey struct {
	Name    string
	Version string
	Release string
}

type vulnTxKey struct {
	TransactionID string
	MachineID     string
}

func UpdateVulnerabilitiesJob(db *sql.DB) {
	logger.Info("Vulnerabilities: executing update task...")

	lockName := "vulnerabilities"

	locked, err := acquireLock(db, lockName)
	if err != nil {
		logger.Error("Error acquiring lock for vulnerabilities: " + err.Error())
		return
	}
	if !locked {
		logger.Info("Another instance is running this vulnerabilities job.")
		return
	}
	defer releaseLock(db, lockName)

	// Extract all distinct packages from transaction items, joined with asset OS.
	query := `
        SELECT DISTINCT ti.package, ti.version, COALESCE(ti.release, '') AS release, a.os, COALESCE(ti.repo, '') AS repo
        FROM transaction_items ti
        JOIN transactions t ON ti.transaction_id = t.transaction_id AND ti.machine_id = t.machine_id
        JOIN assets a ON t.machine_id = a.machine_id AND t.hostname = a.hostname
        WHERE ti.action IN ('Install', 'Upgrade', 'Downgrade', 'Reinstall', 'installed', 'upgrade',
                             'Removed', 'Upgraded', 'Downgraded', 'Obsoleted', 'removed')
    `
	rows, err := db.Query(query)
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
	pkgMap := make(map[string]pkg)

	for rows.Next() {
		var pName, pVersion, pRelease, pOs, pRepo sql.NullString
		if err := rows.Scan(&pName, &pVersion, &pRelease, &pOs, &pRepo); err != nil {
			logger.Error("Vulnerabilities scan error: " + err.Error())
			continue
		}

		if !pName.Valid || !pVersion.Valid {
			continue
		}

		ecos := util.ExtractOSVEcosystems(pOs.String, pRepo.String)
		for _, eco := range ecos {
			key := fmt.Sprintf("%s|%s|%s|%s", pName.String, pVersion.String, pRelease.String, eco)
			pkgMap[key] = pkg{Name: pName.String, Version: pVersion.String, Release: pRelease.String, Ecosystem: eco}
		}
	}

	packages := make([]pkg, 0, len(pkgMap))
	for _, v := range pkgMap {
		packages = append(packages, v)
	}

	logger.Info(fmt.Sprintf("Vulnerabilities: found %d discrete package/ecosystem pairs to check.", len(packages)))

	// Cache for detailed vulnerability data
	fetchedVulns := make(map[string]*util.OSVVuln)
	var fetchedMu sync.Mutex

	// Track which packages had vulnerability data changed for incremental scoreboard
	updatedPackages := make(map[vulnPkgKey]bool)

	chunkSize := 500
	for i := 0; i < len(packages); i += chunkSize {
		end := i + chunkSize
		if end > len(packages) {
			end = len(packages)
		}

		logger.Info(fmt.Sprintf("Vulnerabilities: fetching %d to %d of %d...", i+1, end, len(packages)))

		chunk := packages[i:end]
		osvQueries := make([]util.OSVQuery, 0, len(chunk))

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

		// Collect unique vulnerability IDs that need detail fetching
		var idsToFetch []string
		for _, result := range resp.Results {
			for _, batchVuln := range result.Vulns {
				fetchedMu.Lock()
				_, found := fetchedVulns[batchVuln.ID]
				fetchedMu.Unlock()
				if !found {
					idsToFetch = append(idsToFetch, batchVuln.ID)
				}
			}
		}

		// Deduplicate IDs
		idSet := make(map[string]bool)
		uniqueIDs := make([]string, 0, len(idsToFetch))
		for _, id := range idsToFetch {
			if !idSet[id] {
				idSet[id] = true
				uniqueIDs = append(uniqueIDs, id)
			}
		}

		// Fetch vulnerability details concurrently with a worker pool
		if len(uniqueIDs) > 0 {
			logger.Info(fmt.Sprintf("Vulnerabilities: fetching details for %d unique CVEs...", len(uniqueIDs)))
			const workers = 10
			idChan := make(chan string, len(uniqueIDs))
			var wg sync.WaitGroup

			for _, id := range uniqueIDs {
				idChan <- id
			}
			close(idChan)

			for w := 0; w < workers && w < len(uniqueIDs); w++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for id := range idChan {
						fetched, err := util.FetchOSVVulnerabilityDetails(id)
						fetchedMu.Lock()
						if err == nil && fetched != nil {
							fetchedVulns[id] = fetched
						}
						fetchedMu.Unlock()
					}
				}()
			}
			wg.Wait()
		}

		// Collect batch data for bulk inserts
		vulnBatch := make(map[string]vulnRecord)
		var pvBatch []pvRecord

		for j, result := range resp.Results {
			targetPkg := chunk[j]
			if len(result.Vulns) > 0 {
				for _, batchVuln := range result.Vulns {
					fetchedMu.Lock()
					vuln, found := fetchedVulns[batchVuln.ID]
					fetchedMu.Unlock()
					if !found {
						vuln = &batchVuln // fallback
					}

					// Use structured severity extraction
					severity, cvssScore := vuln.ExtractSeverityAndScore()

					var modifiedAt, publishedAt *time.Time
					if !vuln.ModifiedAt.IsZero() {
						modifiedAt = &vuln.ModifiedAt
					}
					if !vuln.Published.IsZero() {
						publishedAt = &vuln.Published
					}

					vulnBatch[vuln.ID] = vulnRecord{
						ID:          vuln.ID,
						Summary:     vuln.Summary,
						Details:     vuln.Details,
						Severity:    severity,
						CVSSScore:   cvssScore,
						ModifiedAt:  modifiedAt,
						PublishedAt: publishedAt,
					}

					if targetPkg.Ecosystem != "" {
						pvBatch = append(pvBatch, pvRecord{
							PackageName:     targetPkg.Name,
							Version:         targetPkg.Version,
							Release:         targetPkg.Release,
							VulnerabilityID: vuln.ID,
							Ecosystem:       targetPkg.Ecosystem,
						})

						updatedPackages[vulnPkgKey{
							Name:    targetPkg.Name,
							Version: targetPkg.Version,
							Release: targetPkg.Release,
						}] = true
					}
				}
			}
		}

		// Batch upsert vulnerabilities (multi-row INSERT ... ON CONFLICT)
		if len(vulnBatch) > 0 {
			batchUpsertVulnerabilities(db, vulnBatch)
		}

		// Batch upsert package_vulnerabilities
		if len(pvBatch) > 0 {
			batchUpsertPackageVulnerabilities(db, pvBatch)
		}
	}

	logger.Info("Vulnerabilities downloaded. Proceeding to calculate transaction scoreboards...")
	updateTransactionScoreboards(db, updatedPackages)
	logger.Info("Vulnerabilities and transaction scoreboards updated successfully.")
}

// batchUpsertVulnerabilities inserts/updates vulnerabilities in batches of 200 rows.
func batchUpsertVulnerabilities(db *sql.DB, records map[string]vulnRecord) {
	var all []vulnRecord
	for _, r := range records {
		all = append(all, r)
	}

	batchSize := 200
	for i := 0; i < len(all); i += batchSize {
		end := i + batchSize
		if end > len(all) {
			end = len(all)
		}
		batch := all[i:end]

		var valueParts []string
		var args []interface{}
		idx := 1

		for _, r := range batch {
			valueParts = append(valueParts, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				idx, idx+1, idx+2, idx+3, idx+4, idx+5, idx+6))
			args = append(args, r.ID, r.Summary, r.Details, r.Severity, r.CVSSScore, r.ModifiedAt, r.PublishedAt)
			idx += 7
		}

		stmt := fmt.Sprintf(`
			INSERT INTO vulnerabilities (id, summary, details, severity, cvss_score, modified_at, published_at)
			VALUES %s
			ON CONFLICT (id) DO UPDATE SET
				summary = EXCLUDED.summary,
				details = EXCLUDED.details,
				severity = EXCLUDED.severity,
				cvss_score = EXCLUDED.cvss_score,
				modified_at = EXCLUDED.modified_at
		`, strings.Join(valueParts, ", "))

		_, err := db.Exec(stmt, args...)
		if err != nil {
			logger.Error("Batch upsert vulnerabilities error: " + err.Error())
		}
	}
}

// batchUpsertPackageVulnerabilities inserts package↔vulnerability links in batches.
func batchUpsertPackageVulnerabilities(db *sql.DB, records []pvRecord) {
	batchSize := 200
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}
		batch := records[i:end]

		var valueParts []string
		var args []interface{}
		idx := 1

		for _, r := range batch {
			valueParts = append(valueParts, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)",
				idx, idx+1, idx+2, idx+3, idx+4))
			args = append(args, r.PackageName, r.Version, r.Release, r.VulnerabilityID, r.Ecosystem)
			idx += 5
		}

		stmt := fmt.Sprintf(`
			INSERT INTO package_vulnerabilities (package_name, version, release, vulnerability_id, ecosystem)
			VALUES %s
			ON CONFLICT DO NOTHING
		`, strings.Join(valueParts, ", "))

		_, err := db.Exec(stmt, args...)
		if err != nil {
			logger.Error("Batch upsert package_vulnerabilities error: " + err.Error())
		}
	}
}

func updateTransactionScoreboards(db *sql.DB, updatedPackages map[vulnPkgKey]bool) {
	if len(updatedPackages) == 0 {
		logger.Info("Vulnerabilities: No packages were updated, skipping scoreboard recalculation.")
		return
	}

	logger.Info(fmt.Sprintf("Vulnerabilities: %d packages had vulnerability updates. Fetching affected transactions...", len(updatedPackages)))

	// Build arrays of package names/versions/releases that were updated
	var pkgNames, pkgVersions, pkgReleases []string
	for k := range updatedPackages {
		pkgNames = append(pkgNames, k.Name)
		pkgVersions = append(pkgVersions, k.Version)
		pkgReleases = append(pkgReleases, k.Release)
	}

	// Find only transactions that contain items matching the updated packages
	rows, err := db.Query(`
		SELECT DISTINCT ti.transaction_id, ti.machine_id
		FROM transaction_items ti
		WHERE EXISTS (
			SELECT 1 FROM unnest($1::text[], $2::text[], $3::text[]) AS u(pkg, ver, rel)
			WHERE ti.package = u.pkg AND ti.version = u.ver AND COALESCE(ti.release, '') = u.rel
		)
	`, pq.Array(pkgNames), pq.Array(pkgVersions), pq.Array(pkgReleases))
	if err != nil {
		logger.Error("Failed to fetch affected transactions: " + err.Error())
		// Fallback to processing all transactions
		updateAllTransactionScoreboards(db)
		return
	}

	var keys []vulnTxKey
	for rows.Next() {
		var k vulnTxKey
		if err := rows.Scan(&k.TransactionID, &k.MachineID); err == nil {
			keys = append(keys, k)
		}
	}
	rows.Close()

	total := len(keys)
	logger.Info(fmt.Sprintf("Vulnerabilities: %d affected transactions to process (incremental).", total))

	if total == 0 {
		return
	}

	processScoreboardBatch(db, keys)
}

// updateAllTransactionScoreboards is the fallback that processes all transactions.
func updateAllTransactionScoreboards(db *sql.DB) {
	logger.Info("Vulnerabilities: Fallback - fetching ALL transactions for scoreboard calculation...")

	var keys []vulnTxKey

	rows, err := db.Query("SELECT DISTINCT transaction_id, machine_id FROM transactions")
	if err != nil {
		logger.Error("Failed to fetch transactions list: " + err.Error())
		return
	}
	for rows.Next() {
		var k vulnTxKey
		if err := rows.Scan(&k.TransactionID, &k.MachineID); err == nil {
			keys = append(keys, k)
		}
	}
	rows.Close()

	total := len(keys)
	logger.Info(fmt.Sprintf("Vulnerabilities: Total of %d transactions to process.", total))

	processScoreboardBatch(db, keys)
}

func processScoreboardBatch(db *sql.DB, keys []vulnTxKey) {
	total := len(keys)
	chunkSize := 500
	for i := 0; i < total; i += chunkSize {
		end := i + chunkSize
		if end > total {
			end = total
		}

		pct := float64(i) / float64(total) * 100
		logger.Info(fmt.Sprintf("Vulnerabilities: Processing batch %d to %d of %d (%.1f%%)...", i+1, end, total, pct))

		chunk := keys[i:end]
		txnIDs := make([]string, 0, len(chunk))
		mchnIDs := make([]string, 0, len(chunk))
		for _, k := range chunk {
			txnIDs = append(txnIDs, k.TransactionID)
			mchnIDs = append(mchnIDs, k.MachineID)
		}

		stmt := `
WITH batch AS (
    SELECT unnest($1::text[])::integer AS transaction_id, unnest($2::text[]) AS machine_id
),
mapped_assets AS (
    SELECT
        b.transaction_id,
        b.machine_id,
        CASE
            WHEN a.os ILIKE '%AlmaLinux%' THEN 'AlmaLinux:' || SUBSTRING(a.os FROM '[0-9]+')
            WHEN a.os ILIKE '%Rocky%' THEN 'Rocky Linux:' || SUBSTRING(a.os FROM '[0-9]+')
            WHEN a.os ILIKE '%Red Hat%' OR a.os ILIKE '%RHEL%' OR a.os ILIKE '%CentOS%' OR a.os ILIKE '%Oracle%' THEN 'Red Hat:enterprise_linux:' || SUBSTRING(a.os FROM '[0-9]+')
            ELSE ''
        END AS ecosystem_prefix,
        CASE
            WHEN a.os ILIKE '%Red Hat%' OR a.os ILIKE '%RHEL%' OR a.os ILIKE '%CentOS%' OR a.os ILIKE '%Oracle%' THEN TRUE
            ELSE FALSE
        END AS is_rh_family
    FROM batch b
    JOIN transactions trans ON b.transaction_id = trans.transaction_id AND b.machine_id = trans.machine_id
    JOIN assets a ON trans.machine_id = a.machine_id AND trans.hostname = a.hostname
),
vuln_actions AS (
    SELECT
        ti.transaction_id,
        ti.machine_id,
        pv.vulnerability_id,
        v.severity,
        v.cvss_score,
        CASE WHEN ti.action IN ('Install', 'Upgrade', 'Downgrade', 'Reinstall', 'installed', 'upgrade') THEN 1 ELSE 0 END AS is_installed,
        CASE WHEN ti.action IN ('Removed', 'Obsoleted', 'Upgraded', 'Downgraded', 'removed') THEN 1 ELSE 0 END AS is_removed
    FROM transaction_items ti
    JOIN mapped_assets ma ON ti.transaction_id = ma.transaction_id AND ti.machine_id = ma.machine_id
    JOIN package_vulnerabilities pv ON pv.package_name = ti.package AND pv.version = ti.version
         AND pv.release = COALESCE(ti.release, '')
         AND (
             (NOT ma.is_rh_family AND pv.ecosystem = ma.ecosystem_prefix) OR
             (ma.is_rh_family AND pv.ecosystem LIKE ma.ecosystem_prefix || '::%')
         )
    JOIN vulnerabilities v ON v.id = pv.vulnerability_id
    WHERE ti.action IN ('Install', 'Upgrade', 'Downgrade', 'Reinstall', 'installed', 'upgrade', 'Removed', 'Obsoleted', 'Upgraded', 'Downgraded', 'removed')
),
vuln_status AS (
    SELECT
        transaction_id,
        machine_id,
        vulnerability_id,
        severity,
        cvss_score,
        MAX(is_installed) AS has_installed,
        MAX(is_removed) AS has_removed
    FROM vuln_actions
    GROUP BY transaction_id, machine_id, vulnerability_id, severity, cvss_score
),
transaction_summary AS (
    SELECT
        transaction_id,
        machine_id,
        COUNT(DISTINCT CASE WHEN has_removed = 1 AND has_installed = 0 THEN vulnerability_id END) AS total_fixed,
        COUNT(DISTINCT CASE WHEN has_removed = 1 AND has_installed = 0 AND severity = 'CRITICAL' THEN vulnerability_id END) AS critical_fixed,
        COUNT(DISTINCT CASE WHEN has_removed = 1 AND has_installed = 0 AND severity = 'HIGH' THEN vulnerability_id END) AS high_fixed,
        COUNT(DISTINCT CASE WHEN has_removed = 1 AND has_installed = 0 AND severity = 'MEDIUM' THEN vulnerability_id END) AS medium_fixed,
        COUNT(DISTINCT CASE WHEN has_removed = 1 AND has_installed = 0 AND severity = 'LOW' THEN vulnerability_id END) AS low_fixed,
        COALESCE(SUM(CASE WHEN has_removed = 1 AND has_installed = 0 THEN cvss_score ELSE 0 END), 0) AS fixed_cvss,
        
        COUNT(DISTINCT CASE WHEN has_installed = 1 AND has_removed = 0 THEN vulnerability_id END) AS total_introduced,
        COUNT(DISTINCT CASE WHEN has_installed = 1 AND has_removed = 0 AND severity = 'CRITICAL' THEN vulnerability_id END) AS critical_introduced,
        COUNT(DISTINCT CASE WHEN has_installed = 1 AND has_removed = 0 AND severity = 'HIGH' THEN vulnerability_id END) AS high_introduced,
        COUNT(DISTINCT CASE WHEN has_installed = 1 AND has_removed = 0 AND severity = 'MEDIUM' THEN vulnerability_id END) AS medium_introduced,
        COUNT(DISTINCT CASE WHEN has_installed = 1 AND has_removed = 0 AND severity = 'LOW' THEN vulnerability_id END) AS low_introduced,
        COALESCE(SUM(CASE WHEN has_installed = 1 AND has_removed = 0 THEN cvss_score ELSE 0 END), 0) AS introduced_cvss
    FROM vuln_status
    GROUP BY transaction_id, machine_id
)
UPDATE transactions t
SET
    vulns_fixed = COALESCE(s.total_fixed, 0),
    vulns_introduced = COALESCE(s.total_introduced, 0),
    critical_vulns_fixed = COALESCE(s.critical_fixed, 0),
    critical_vulns_introduced = COALESCE(s.critical_introduced, 0),
    high_vulns_fixed = COALESCE(s.high_fixed, 0),
    high_vulns_introduced = COALESCE(s.high_introduced, 0),
    medium_vulns_fixed = COALESCE(s.medium_fixed, 0),
    medium_vulns_introduced = COALESCE(s.medium_introduced, 0),
    low_vulns_fixed = COALESCE(s.low_fixed, 0),
    low_vulns_introduced = COALESCE(s.low_introduced, 0),
    risk_score_mitigated = COALESCE(s.fixed_cvss, 0) - COALESCE(s.introduced_cvss, 0),
    is_security_patch = (COALESCE(s.critical_fixed, 0) > 0 OR COALESCE(s.high_fixed, 0) > 0 OR COALESCE(s.total_fixed, 0) > 0)
FROM transaction_summary s
WHERE t.transaction_id = s.transaction_id
  AND t.machine_id = s.machine_id;
		`

		_, err := db.Exec(stmt, pq.Array(txnIDs), pq.Array(mchnIDs))
		if err != nil {
			logger.Error("Failed to update transaction scoreboards for batch: " + err.Error())
		}
	}
}
