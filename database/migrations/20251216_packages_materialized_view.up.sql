-- Create a materialized view to pre-compute package listing data
-- This dramatically improves the /packages endpoint performance by avoiding
-- multiple full table scans on transaction_items for each request

-- Drop the view if it exists (for re-runs)
DROP MATERIALIZED VIEW IF EXISTS mv_package_listing;

-- Create the materialized view
CREATE MATERIALIZED VIEW mv_package_listing AS
WITH DistinctPackages AS (
    -- Get distinct package names (removing 'Change ' prefix if present)
    SELECT DISTINCT
        CASE
            WHEN package LIKE 'Change %' THEN SUBSTRING(package FROM 8)
            ELSE package
        END AS clean_package
    FROM public.transaction_items
),
LatestVersions AS (
    -- Get the latest version/release for each package
    SELECT DISTINCT ON (
        CASE
            WHEN package LIKE 'Change %' THEN SUBSTRING(package FROM 8)
            ELSE package
        END
    )
        CASE
            WHEN package LIKE 'Change %' THEN SUBSTRING(package FROM 8)
            ELSE package
        END AS package,
        version,
        release,
        arch,
        repo
    FROM public.transaction_items
    ORDER BY
        CASE
            WHEN package LIKE 'Change %' THEN SUBSTRING(package FROM 8)
            ELSE package
        END,
        version DESC,
        release DESC
),
VersionCounts AS (
    -- Count unique version/release combinations for each package
    SELECT
        CASE
            WHEN package LIKE 'Change %' THEN SUBSTRING(package FROM 8)
            ELSE package
        END AS package,
        COUNT(DISTINCT (version, release)) as total_versions
    FROM public.transaction_items
    GROUP BY
        CASE
            WHEN package LIKE 'Change %' THEN SUBSTRING(package FROM 8)
            ELSE package
        END
),
MachineCounts AS (
    -- Count unique active machines for each package
    SELECT
        CASE
            WHEN ti.package LIKE 'Change %' THEN SUBSTRING(ti.package FROM 8)
            ELSE ti.package
        END AS package,
        COUNT(DISTINCT ti.machine_id) as machine_count
    FROM public.transaction_items ti
    INNER JOIN public.assets a ON ti.machine_id = a.machine_id
    WHERE a.is_active = TRUE
    GROUP BY
        CASE
            WHEN ti.package LIKE 'Change %' THEN SUBSTRING(ti.package FROM 8)
            ELSE ti.package
        END
)
SELECT
    lv.package,
    lv.version,
    lv.release,
    lv.arch,
    lv.repo,
    COALESCE(vc.total_versions, 1) - 1 as other_versions_count,
    COALESCE(mc.machine_count, 0) as machine_count
FROM LatestVersions lv
LEFT JOIN VersionCounts vc ON lv.package = vc.package
LEFT JOIN MachineCounts mc ON lv.package = mc.package
ORDER BY lv.package;

-- Create indexes on the materialized view for fast lookups
CREATE UNIQUE INDEX idx_mv_package_listing_package ON mv_package_listing (package);

-- Try to create GIN index for fast text search (requires pg_trgm extension)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm') THEN
        CREATE INDEX IF NOT EXISTS idx_mv_package_listing_package_text
            ON mv_package_listing USING GIN (package gin_trgm_ops);
    END IF;
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Could not create GIN index on mv_package_listing: %', SQLERRM;
END $$;

-- Add comment to document the view
COMMENT ON MATERIALIZED VIEW mv_package_listing IS
'Pre-computed package listing data for the /packages endpoint.
Refresh this view periodically using: REFRESH MATERIALIZED VIEW CONCURRENTLY mv_package_listing;';

