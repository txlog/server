-- Sortable key for RPM version/release strings.
-- Text ordering puts 1.9.4 above 1.18.0 because '9' > '1'. This function
-- tokenizes the string into numeric and alphabetic runs and zero-pads the
-- numeric ones, so plain text ordering of the key matches version ordering.
CREATE OR REPLACE FUNCTION public.version_sort_key(v text)
RETURNS text
LANGUAGE sql IMMUTABLE PARALLEL SAFE STRICT
AS $$
  SELECT string_agg(
           CASE WHEN m[1] ~ '^[0-9]' THEN lpad(m[1], 10, '0') ELSE m[1] END,
           '.' ORDER BY ord
         )
  FROM regexp_matches(v, '[0-9]+|[A-Za-z]+', 'g') WITH ORDINALITY AS t(m, ord);
$$;

COMMENT ON FUNCTION public.version_sort_key(text) IS
'Sortable key for RPM version/release strings: numeric runs are zero-padded so
1.18.0 sorts above 1.9.4. Does not implement rpmvercmp tilde (~) semantics.';

-- Recreate mv_package_listing so "latest version" uses the same ordering as the
-- package detail timeline. Body is unchanged except the LatestVersions ORDER BY.
DROP MATERIALIZED VIEW IF EXISTS mv_package_listing;

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
        version_sort_key(version) DESC,
        version_sort_key(release) DESC
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

COMMENT ON MATERIALIZED VIEW mv_package_listing IS
'Pre-computed package listing data for the /packages endpoint.
Refresh this view periodically using: REFRESH MATERIALIZED VIEW CONCURRENTLY mv_package_listing;';
