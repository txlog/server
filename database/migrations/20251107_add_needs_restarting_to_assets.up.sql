-- Add needs_restarting and restarting_reason columns to assets table
-- These columns track whether an asset requires a restart and the reason why

ALTER TABLE assets
ADD COLUMN needs_restarting BOOLEAN,
ADD COLUMN restarting_reason TEXT;

-- Create index for quickly finding assets that need restarting
CREATE INDEX idx_assets_needs_restarting ON assets(needs_restarting) WHERE needs_restarting = TRUE;

-- Add comments to document the new columns
COMMENT ON COLUMN assets.needs_restarting IS 'Whether this asset requires a restart (e.g., due to kernel updates or system library changes). NULL if unknown, TRUE if restart needed, FALSE if no restart needed.';
COMMENT ON COLUMN assets.restarting_reason IS 'Explanation of why the asset needs restarting (e.g., list of services or kernel version). NULL if needs_restarting is FALSE or unknown.';

-- Migrate data from executions table to assets table
-- For each asset, get the most recent execution and copy needs_restarting and restarting_reason
UPDATE assets a
SET
    needs_restarting = e.needs_restarting,
    restarting_reason = e.restarting_reason
FROM (
    SELECT DISTINCT ON (machine_id)
        machine_id,
        needs_restarting,
        restarting_reason
    FROM executions
    WHERE needs_restarting IS NOT NULL
    ORDER BY machine_id, executed_at DESC
) e
WHERE a.machine_id = e.machine_id;
