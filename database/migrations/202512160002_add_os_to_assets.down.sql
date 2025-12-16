-- Rollback: Remove the os column from assets table
ALTER TABLE assets DROP COLUMN IF EXISTS os;

-- Remove the index
DROP INDEX IF EXISTS idx_assets_os;
DROP INDEX IF EXISTS idx_assets_hostname_active;
