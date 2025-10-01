-- Create basic indexes first (these always work)
CREATE INDEX IF NOT EXISTS idx_ti_pkg_ver_rel ON public.transaction_items (package, version DESC, release DESC);
CREATE INDEX IF NOT EXISTS idx_ti_pkg_machid ON public.transaction_items (package, machine_id);

-- Try to create pg_trgm extension and GIN index
-- Note: This requires postgresql-contrib package to be installed
-- If it fails, the migration continues with basic indexes only
DO $$
BEGIN
    -- Try to create the extension
    CREATE EXTENSION IF NOT EXISTS pg_trgm;
    
    -- If extension exists, create the GIN index
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm') THEN
        CREATE INDEX IF NOT EXISTS idx_ti_package_gin ON public.transaction_items USING GIN (package gin_trgm_ops);
    END IF;
EXCEPTION
    WHEN OTHERS THEN
        -- Log the error but don't fail the migration
        RAISE NOTICE 'Could not create pg_trgm extension or GIN index: %', SQLERRM;
END $$;
