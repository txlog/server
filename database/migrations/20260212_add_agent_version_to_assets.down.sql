DROP INDEX IF EXISTS idx_assets_agent_version;
ALTER TABLE assets DROP COLUMN IF EXISTS agent_version;
