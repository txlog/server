-- Remove copy_fail column from assets and executions tables

DROP INDEX IF EXISTS idx_assets_copy_fail;

ALTER TABLE assets
DROP COLUMN IF EXISTS copy_fail;

ALTER TABLE executions
DROP COLUMN IF EXISTS copy_fail;
