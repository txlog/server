-- Add agent_version column to assets table
ALTER TABLE assets ADD COLUMN IF NOT EXISTS agent_version TEXT;

-- Index for filtering by agent_version
CREATE INDEX IF NOT EXISTS idx_assets_agent_version ON assets(agent_version)
    WHERE agent_version IS NOT NULL;

-- Backfill agent_version from the latest execution for existing assets
UPDATE assets a SET agent_version = (
    SELECT e.agent_version FROM executions e
    WHERE e.machine_id = a.machine_id AND e.hostname = a.hostname
    ORDER BY e.executed_at DESC LIMIT 1
) WHERE a.agent_version IS NULL;
