-- Recreate mv_package_listing with the original text-based version ordering,
-- then drop the sort-key function.
DROP MATERIALIZED VIEW IF EXISTS mv_package_listing;

CREATE MATERIALIZED VIEW mv_package_listing AS
WITH DistinctPackages AS (
    SELECT DISTINCT
        CASE
            WHEN package LIKE 'Change %' THEN SUBSTRING(package FROM 8)
            ELSE package
        END AS clean_package
    FROM public.transaction_items
),
LatestVersions AS (
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

CREATE UNIQUE INDEX idx_mv_package_listing_package ON mv_package_listing (package);

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

DROP FUNCTION IF EXISTS public.version_sort_key(text);
