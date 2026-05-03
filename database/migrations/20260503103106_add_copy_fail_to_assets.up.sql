-- Add copy_fail column to assets and executions tables

ALTER TABLE assets
ADD COLUMN copy_fail BOOLEAN;

ALTER TABLE executions
ADD COLUMN copy_fail BOOLEAN;

CREATE INDEX idx_assets_copy_fail ON assets(copy_fail) WHERE copy_fail = TRUE;

COMMENT ON COLUMN assets.copy_fail IS 'Whether this asset is vulnerable to CVE-2026-31431 (Copy Fail). NULL if unknown, TRUE if vulnerable, FALSE if safe.';

-- For each asset, get the most recent execution and copy copy_fail (if any past execution had it, though it will be empty initially)
UPDATE assets a
SET 
    copy_fail = e.copy_fail
FROM (
    SELECT 
        machine_id,
        hostname,
        copy_fail,
        ROW_NUMBER() OVER(PARTITION BY machine_id ORDER BY executed_at DESC) as rn
    FROM executions
    WHERE copy_fail IS NOT NULL
) e
WHERE a.machine_id = e.machine_id 
AND a.hostname = e.hostname 
AND e.rn = 1;
