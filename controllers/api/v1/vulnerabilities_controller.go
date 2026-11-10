package v1

import (
	"database/sql"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// ecosystemMatchesAsset restricts a package_vulnerabilities row (pv) to the OSV
// ecosystem of an asset (a).
const ecosystemMatchesAsset = `(
			      (a.os ILIKE '%AlmaLinux%' AND pv.ecosystem = 'AlmaLinux:' || SUBSTRING(a.os FROM '[0-9]+')) OR
			      (a.os ILIKE '%Rocky%' AND pv.ecosystem = 'Rocky Linux:' || SUBSTRING(a.os FROM '[0-9]+')) OR
			      ((a.os ILIKE '%Red Hat%' OR a.os ILIKE '%RHEL%' OR a.os ILIKE '%CentOS%' OR a.os ILIKE '%Oracle%') AND pv.ecosystem LIKE 'Red Hat:enterprise_linux:' || SUBSTRING(a.os FROM '[0-9]+') || '::%')
			  )`

var cvePattern = regexp.MustCompile(`^CVE-\d{4}-\d{4,}$`)

// CVEAsset is the exposure of one active asset to a CVE. Status is one of
// vulnerable, fixed, unknown or not_affected.
type CVEAsset struct {
	Hostname  string     `json:"hostname"`
	MachineID string     `json:"machine_id"`
	Status    string     `json:"status"`
	Packages  []string   `json:"packages,omitempty"` // affected builds still installed
	Since     *time.Time `json:"since,omitzero"`     // when the last affected build was removed
}

// CVEAssetsResponse lists the advisories covering a CVE and the fleet's exposure.
type CVEAssetsResponse struct {
	CVE        string     `json:"cve"`
	Advisories []string   `json:"advisories"`
	Assets     []CVEAsset `json:"assets"`
}

// GetCVEAssets Get the exposure of every active asset to a CVE
//
//	@summary		Get the exposure of every active asset to a CVE
//	@description	Lists active assets as vulnerable, fixed (with the date the last affected build was removed), not_affected (no affected build recorded) or unknown (operating system without OSV data)
//	@tags			vulnerabilities
//	@produce		json
//	@param			cve	path		string	true	"CVE identifier, e.g. CVE-2024-2398"
//	@success		200	{object}	CVEAssetsResponse
//	@failure		400	{object}	map[string]string	"Invalid CVE identifier"
//	@failure		404	{object}	map[string]string	"CVE not found in the OSV data"
//	@failure		500	{object}	map[string]string	"Internal server error"
//	@router			/v1/vulnerabilities/{cve}/assets [get]
//	@security		ApiKeyAuth
func GetCVEAssets(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		cve := strings.ToUpper(strings.TrimSpace(c.Param("cve")))
		if !cvePattern.MatchString(cve) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid CVE identifier. Expected a format like CVE-2024-2398."})
			return
		}
		ctx := c.Request.Context()

		var advisories []string
		err := database.QueryRowContext(ctx,
			`SELECT COALESCE(array_agg(id ORDER BY id), '{}') FROM vulnerabilities WHERE $1 = ANY(upstream)`,
			cve,
		).Scan(pq.Array(&advisories))
		if err != nil {
			slog.Error("Error querying CVE advisories: " + err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		if len(advisories) == 0 {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "No advisory covers " + cve + " in the OSV data. If this is a new installation or upgrade, run the OSV update in the admin panel."})
			return
		}

		// A build is looked at individually because installonly packages such as
		// the kernel keep several versions installed at once. transaction_id
		// follows the history of each machine, so its highest value is the
		// latest event of that build.
		rows, err := database.QueryContext(ctx, `
			WITH last_event AS (
			    SELECT DISTINCT ON (ti.machine_id, ti.package, ti.version, ti.release, ti.arch)
			           ti.machine_id,
			           ti.package || '-' || ti.version || '-' || COALESCE(ti.release, '') || '.' || COALESCE(ti.arch, '') AS build,
			           ti.action IN ('Install', 'Upgrade', 'Downgrade', 'Reinstall', 'installed', 'upgrade') AS installed,
			           COALESCE(t.end_time, t.begin_time) AS event_time
			    FROM transaction_items ti
			    JOIN transactions t ON t.transaction_id = ti.transaction_id AND t.machine_id = ti.machine_id
			    JOIN assets a ON a.machine_id = ti.machine_id AND a.is_active
			    JOIN package_vulnerabilities pv ON pv.package_name = ti.package AND pv.version = ti.version AND pv.release = COALESCE(ti.release, '')
			      AND `+ecosystemMatchesAsset+`
			    WHERE pv.vulnerability_id = ANY($1)
			      AND ti.action IN ('Install', 'Upgrade', 'Downgrade', 'Reinstall', 'installed', 'upgrade',
			                        'Removed', 'Upgraded', 'Downgraded', 'Obsoleted', 'removed')
			    ORDER BY ti.machine_id, ti.package, ti.version, ti.release, ti.arch, ti.transaction_id DESC, installed DESC
			),
			exposure AS (
			    SELECT machine_id,
			           bool_or(installed) AS vulnerable,
			           array_agg(build ORDER BY build) FILTER (WHERE installed) AS packages,
			           MAX(event_time) FILTER (WHERE NOT installed) AS since
			    FROM last_event
			    GROUP BY machine_id
			)
			SELECT hostname, machine_id, status, packages, since
			FROM (
			    SELECT a.hostname, a.machine_id, e.packages,
			           CASE WHEN e.vulnerable THEN NULL ELSE e.since END AS since,
			           CASE
			               WHEN e.vulnerable THEN 'vulnerable'
			               WHEN e.machine_id IS NOT NULL THEN 'fixed'
			               WHEN a.os ~* '(almalinux|rocky|red hat|rhel|centos|oracle).*[0-9]' THEN 'not_affected'
			               ELSE 'unknown'
			           END AS status
			    FROM assets a
			    LEFT JOIN exposure e ON e.machine_id = a.machine_id
			    WHERE a.is_active
			) r
			ORDER BY array_position(ARRAY['vulnerable', 'fixed', 'unknown', 'not_affected'], r.status), r.hostname`,
			pq.Array(advisories),
		)
		if err != nil {
			slog.Error("Error querying CVE assets: " + err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		defer rows.Close()

		assets := []CVEAsset{}
		for rows.Next() {
			var a CVEAsset
			if err := rows.Scan(&a.Hostname, &a.MachineID, &a.Status, pq.Array(&a.Packages), &a.Since); err != nil {
				slog.Error("Error scanning CVE asset: " + err.Error())
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			assets = append(assets, a)
		}
		if err := rows.Err(); err != nil {
			slog.Error("Error reading CVE assets: " + err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(http.StatusOK, CVEAssetsResponse{CVE: cve, Advisories: advisories, Assets: assets})
	}
}

type TransactionVulnerability struct {
	ID        string  `json:"id"`
	Summary   string  `json:"summary"`
	Severity  string  `json:"severity"`
	CvssScore float64 `json:"cvss_score"`
	Package   string  `json:"package"`
	Version   string  `json:"version"`
	Type      string  `json:"type"`
}

func GetTransactionVulnerabilities(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		machineID := c.Query("machine_id")
		transactionID := c.Query("transaction_id")

		if machineID == "" || transactionID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, "machine_id and transaction_id are required")
			return
		}

		rows, err := database.QueryContext(c.Request.Context(), `
			SELECT DISTINCT v.id, COALESCE(v.summary, ''), v.severity, v.cvss_score,
			       ti.package, ti.version,
			       CASE
			           WHEN ti.action IN ('Removed', 'Upgraded', 'Downgraded', 'Obsoleted', 'removed') THEN 'fixed'
			           ELSE 'introduced'
			       END as type
			FROM transaction_items ti
			JOIN assets a ON ti.machine_id = a.machine_id
			JOIN package_vulnerabilities pv ON pv.package_name = ti.package AND pv.version = ti.version AND pv.release = COALESCE(ti.release, '')
			  AND `+ecosystemMatchesAsset+`
			JOIN vulnerabilities v ON v.id = pv.vulnerability_id
			WHERE ti.machine_id = $1
			  AND ti.transaction_id = $2
			  -- Only package changes count: an action such as "Reason Change"
			  -- names an installed build without installing or removing it.
			  AND ti.action IN ('Install', 'Upgrade', 'Downgrade', 'Reinstall', 'installed', 'upgrade',
			                    'Removed', 'Upgraded', 'Downgraded', 'Obsoleted', 'removed')
			ORDER BY v.cvss_score DESC, v.id ASC`,
			machineID,
			transactionID,
		)

		if err != nil {
			slog.Error("Error querying transaction vulnerabilities: " + err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		defer rows.Close()

		var vulns []TransactionVulnerability
		for rows.Next() {
			var v TransactionVulnerability
			if err := rows.Scan(&v.ID, &v.Summary, &v.Severity, &v.CvssScore, &v.Package, &v.Version, &v.Type); err != nil {
				slog.Error("Error scanning vulnerability: " + err.Error())
				continue
			}
			vulns = append(vulns, v)
		}

		if vulns == nil {
			vulns = []TransactionVulnerability{}
		}

		c.JSON(http.StatusOK, vulns)
	}
}
