-- Create topology configuration tables.
-- These tables allow administrators to define hostname-based patterns
-- that classify assets into Environments, Services, and Pods.
-- Users define templates (e.g. ":env-dc01-:svc-database:seq") instead of raw
-- regex; the server compiles templates to regex internally.

-- Hostname templates for topology resolution.
-- Each row represents a different hostname format in the infrastructure.
-- Templates use the tags :env, :svc, :seq which are compiled to regex groups.
-- Multiple templates are tried in display_order; the first match wins.
CREATE TABLE IF NOT EXISTS topology_patterns (
    id SERIAL PRIMARY KEY,
    template TEXT NOT NULL,
    compiled_pattern TEXT NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE topology_patterns IS 'Hostname templates for topology classification. Each template uses :env, :svc, :seq tags that the server compiles to regex capture groups.';
COMMENT ON COLUMN topology_patterns.template IS 'Human-readable template, e.g. ":env-dc01-:svc-database:seq". Uses :env, :svc, :seq tags.';
COMMENT ON COLUMN topology_patterns.compiled_pattern IS 'Auto-generated PostgreSQL regex compiled from the template.';
COMMENT ON COLUMN topology_patterns.display_order IS 'Order in which templates are tried during hostname resolution. Lower values are tried first.';

-- Maps captured :env values to friendly environment names.
-- Example: match_value="prd" -> name="Production"
CREATE TABLE IF NOT EXISTS environment_names (
    id SERIAL PRIMARY KEY,
    match_value TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE environment_names IS 'Maps values captured by the :env tag to friendly environment names.';
COMMENT ON COLUMN environment_names.match_value IS 'The raw value extracted from the hostname by the :env tag, e.g. "prd", "hlg".';
COMMENT ON COLUMN environment_names.name IS 'Friendly display name, e.g. "Production", "Homologation".';

-- Maps captured :svc values to friendly service names.
-- Example: match_value="acme-system" -> name="ACME System"
CREATE TABLE IF NOT EXISTS service_names (
    id SERIAL PRIMARY KEY,
    match_value TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE service_names IS 'Maps values captured by the :svc tag to friendly service names.';
COMMENT ON COLUMN service_names.match_value IS 'The raw value extracted from the hostname by the :svc tag, e.g. "acme-system", "billing".';
COMMENT ON COLUMN service_names.name IS 'Friendly display name, e.g. "ACME System", "Billing".';
