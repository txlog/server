DROP INDEX IF EXISTS idx_pv_vulnerability_id;
ALTER TABLE vulnerabilities DROP COLUMN IF EXISTS upstream;
