-- Drop indexes first
DROP INDEX IF EXISTS idx_assets_deactivated_at;
DROP INDEX IF EXISTS idx_assets_machine_id;
DROP INDEX IF EXISTS idx_assets_hostname_active;

-- Drop the assets table
-- Note: This does not affect executions or transactions tables,
-- which remain as the source of truth for historical data
DROP TABLE IF EXISTS assets;
