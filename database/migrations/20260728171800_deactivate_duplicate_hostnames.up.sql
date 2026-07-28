-- Deactivate all duplicate active assets with the same hostname, keeping only the most recently seen one.
WITH RankedAssets AS (
    SELECT
        asset_id,
        ROW_NUMBER() OVER (PARTITION BY hostname ORDER BY last_seen DESC) as rn
    FROM assets
    WHERE is_active = TRUE
)
UPDATE assets
SET is_active = FALSE, deactivated_at = CURRENT_TIMESTAMP
WHERE asset_id IN (
    SELECT asset_id FROM RankedAssets WHERE rn > 1
);
