CREATE TABLE IF NOT EXISTS service_environments (
    service_name_id     INT NOT NULL REFERENCES service_names(id)     ON DELETE CASCADE,
    environment_name_id INT NOT NULL REFERENCES environment_names(id) ON DELETE CASCADE,
    PRIMARY KEY (service_name_id, environment_name_id)
);

CREATE INDEX IF NOT EXISTS idx_service_environments_env ON service_environments (environment_name_id);

COMMENT ON TABLE service_environments IS 'Many-to-many association between topology services and environments; a service is only offered on /topology for the environments it belongs to';
COMMENT ON COLUMN service_environments.service_name_id IS 'Service the association belongs to; cascades on service deletion';
COMMENT ON COLUMN service_environments.environment_name_id IS 'Environment the service is available in; cascades on environment deletion';

-- Backfill: existing services belong to every existing environment, preserving today's behavior.
INSERT INTO service_environments (service_name_id, environment_name_id)
SELECT s.id, e.id FROM service_names s CROSS JOIN environment_names e
ON CONFLICT DO NOTHING;
