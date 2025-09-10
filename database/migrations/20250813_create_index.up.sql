CREATE INDEX idx_ti_pkg_ver_rel ON public.transaction_items (package, version DESC, release DESC);
CREATE INDEX idx_ti_pkg_machid ON public.transaction_items (package, machine_id);
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_ti_package_gin ON public.transaction_items USING GIN (package gin_trgm_ops);
