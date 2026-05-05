ALTER TABLE service_names ADD COLUMN has_pods BOOLEAN NOT NULL DEFAULT FALSE;
COMMENT ON COLUMN service_names.has_pods IS 'If TRUE, assets matching this service will be grouped into pods based on the :seq tag. If FALSE, they will be grouped together under "All Assets".';
