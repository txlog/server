-- Add os column to assets table to avoid expensive LATERAL JOIN queries
-- The OS is updated whenever an execution is received, keeping it always current

-- Add the os column to the assets table
ALTER TABLE assets ADD COLUMN IF NOT EXISTS os TEXT;

-- Create an index for filtering by OS (commonly used in the UI)
CREATE INDEX IF NOT EXISTS idx_assets_os ON assets(os) WHERE os IS NOT NULL;

-- Add comment to document the column
COMMENT ON COLUMN assets.os IS 'Operating system of the asset, updated with each execution. Cached here to avoid expensive LATERAL JOIN queries when listing assets.';

-- Populate the os column for existing assets from their latest execution
-- This is a one-time migration to backfill existing data
UPDATE assets a
SET os = (
    SELECT e.os
    FROM executions e
    WHERE e.machine_id = a.machine_id AND e.hostname = a.hostname
    ORDER BY e.executed_at DESC
    LIMIT 1
)
WHERE a.os IS NULL;

-- Create an optimized index for the asset listing query pattern
-- This helps with ORDER BY hostname queries
CREATE INDEX IF NOT EXISTS idx_assets_hostname_active ON assets(hostname) WHERE is_active = TRUE;
