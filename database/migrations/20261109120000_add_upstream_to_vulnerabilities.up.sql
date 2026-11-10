-- Upstream identifiers (usually CVEs) of each distribution advisory, so that
-- advisories such as ALSA-2024:5529 can be found by CVE.
ALTER TABLE vulnerabilities ADD COLUMN upstream TEXT[];

-- The primary key does not start with vulnerability_id.
CREATE INDEX IF NOT EXISTS idx_pv_vulnerability_id
    ON package_vulnerabilities (vulnerability_id);
