CREATE TABLE vulnerabilities (
    id VARCHAR(100) PRIMARY KEY,
    summary TEXT,
    details TEXT,
    severity VARCHAR(20),
    cvss_score DECIMAL(10,2) DEFAULT 0.0,
    modified_at TIMESTAMP WITH TIME ZONE,
    published_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE package_vulnerabilities (
    package_name VARCHAR(255) NOT NULL,
    version VARCHAR(255) NOT NULL,
    vulnerability_id VARCHAR(100) REFERENCES vulnerabilities(id) ON DELETE CASCADE,
    PRIMARY KEY (package_name, version, vulnerability_id)
);

ALTER TABLE transactions
ADD COLUMN vulns_fixed INTEGER DEFAULT 0,
ADD COLUMN vulns_introduced INTEGER DEFAULT 0,
ADD COLUMN critical_vulns_fixed INTEGER DEFAULT 0,
ADD COLUMN critical_vulns_introduced INTEGER DEFAULT 0,
ADD COLUMN high_vulns_fixed INTEGER DEFAULT 0,
ADD COLUMN high_vulns_introduced INTEGER DEFAULT 0,
ADD COLUMN medium_vulns_fixed INTEGER DEFAULT 0,
ADD COLUMN medium_vulns_introduced INTEGER DEFAULT 0,
ADD COLUMN low_vulns_fixed INTEGER DEFAULT 0,
ADD COLUMN low_vulns_introduced INTEGER DEFAULT 0,
ADD COLUMN is_security_patch BOOLEAN DEFAULT false,
ADD COLUMN risk_score_mitigated DECIMAL(10,2) DEFAULT 0.0,
ADD COLUMN max_severity_fixed VARCHAR(20),
ADD COLUMN vulnerable_packages_updated INTEGER DEFAULT 0;
