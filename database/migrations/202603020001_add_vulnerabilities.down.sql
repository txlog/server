ALTER TABLE transactions
DROP COLUMN vulns_fixed,
DROP COLUMN vulns_introduced,
DROP COLUMN critical_vulns_fixed,
DROP COLUMN critical_vulns_introduced,
DROP COLUMN high_vulns_fixed,
DROP COLUMN high_vulns_introduced,
DROP COLUMN medium_vulns_fixed,
DROP COLUMN medium_vulns_introduced,
DROP COLUMN low_vulns_fixed,
DROP COLUMN low_vulns_introduced,
DROP COLUMN is_security_patch,
DROP COLUMN risk_score_mitigated,
DROP COLUMN max_severity_fixed,
DROP COLUMN vulnerable_packages_updated;

DROP TABLE IF EXISTS package_vulnerabilities;
DROP TABLE IF EXISTS vulnerabilities;
