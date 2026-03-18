package v1

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	logger "github.com/txlog/server/logger"
)

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

		rows, err := database.Query(`
			SELECT DISTINCT v.id, COALESCE(v.summary, ''), v.severity, v.cvss_score,
			       ti.package, ti.version,
			       CASE
			           WHEN ti.action IN ('Removed', 'Upgraded', 'Downgraded', 'Obsoleted', 'removed') THEN 'fixed'
			           ELSE 'introduced'
			       END as type
			FROM transaction_items ti
			JOIN assets a ON ti.machine_id = a.machine_id
			JOIN package_vulnerabilities pv ON pv.package_name = ti.package AND pv.version = ti.version AND pv.release = COALESCE(ti.release, '')
			  AND (
			      (a.os ILIKE '%AlmaLinux%' AND pv.ecosystem = 'AlmaLinux:' || SUBSTRING(a.os FROM '[0-9]+')) OR
			      (a.os ILIKE '%Rocky%' AND pv.ecosystem = 'Rocky Linux:' || SUBSTRING(a.os FROM '[0-9]+')) OR
			      ((a.os ILIKE '%Red Hat%' OR a.os ILIKE '%RHEL%' OR a.os ILIKE '%CentOS%' OR a.os ILIKE '%Oracle%') AND pv.ecosystem = 'AlmaLinux:' || SUBSTRING(a.os FROM '[0-9]+'))
			  )
			JOIN vulnerabilities v ON v.id = pv.vulnerability_id
			WHERE ti.machine_id = $1
			  AND ti.transaction_id = $2
			ORDER BY v.cvss_score DESC, v.id ASC`,
			machineID,
			transactionID,
		)

		if err != nil {
			logger.Error("Error querying transaction vulnerabilities: " + err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var vulns []TransactionVulnerability
		for rows.Next() {
			var v TransactionVulnerability
			if err := rows.Scan(&v.ID, &v.Summary, &v.Severity, &v.CvssScore, &v.Package, &v.Version, &v.Type); err != nil {
				logger.Error("Error scanning vulnerability: " + err.Error())
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
