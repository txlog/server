-- Remove needs_restarting and restarting_reason columns from assets table

DROP INDEX IF EXISTS idx_assets_needs_restarting;

ALTER TABLE assets
DROP COLUMN IF EXISTS needs_restarting,
DROP COLUMN IF EXISTS restarting_reason;
