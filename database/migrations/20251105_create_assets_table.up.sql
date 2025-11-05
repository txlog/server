-- Create assets table to track physical/logical machines in the infrastructure.
-- This table serves as the single source of truth for asset identity, where:
-- - hostname is the logical identifier (e.g., "webserver01")
-- - machine_id tracks physical/OS installations (changes on reinstall/reimaging)
-- This design allows tracking asset replacements while maintaining logical continuity.
CREATE TABLE IF NOT EXISTS assets (
    asset_id SERIAL PRIMARY KEY,
    hostname TEXT NOT NULL,
    machine_id TEXT NOT NULL,
    first_seen TIMESTAMP WITH TIME ZONE NOT NULL,
    last_seen TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deactivated_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT assets_hostname_machine_id_unique UNIQUE(hostname, machine_id)
);

-- Index for querying active assets by hostname (most common query pattern)
CREATE INDEX idx_assets_hostname_active ON assets(hostname, is_active) WHERE is_active = TRUE;

-- Index for looking up assets by machine_id (used for linking with executions/transactions)
CREATE INDEX idx_assets_machine_id ON assets(machine_id);

-- Index for finding recently deactivated assets (useful for auditing replacements)
CREATE INDEX idx_assets_deactivated_at ON assets(deactivated_at) WHERE deactivated_at IS NOT NULL;

-- Table and column comments for documentation
COMMENT ON TABLE assets IS 'Central registry of all managed assets (machines/servers). Tracks both logical identity (hostname) and physical identity (machine_id) to handle asset replacements while maintaining historical continuity.';
COMMENT ON COLUMN assets.asset_id IS 'Unique identifier for this asset record (surrogate key)';
COMMENT ON COLUMN assets.hostname IS 'Logical name of the asset in the infrastructure (e.g., webserver01, database-primary). This is the stable identifier across reinstalls.';
COMMENT ON COLUMN assets.machine_id IS 'Physical/OS installation identifier (typically UUID from /etc/machine-id or similar). Changes when the OS is reinstalled or the hardware is replaced.';
COMMENT ON COLUMN assets.first_seen IS 'Timestamp when this hostname+machine_id combination was first reported to the server';
COMMENT ON COLUMN assets.last_seen IS 'Timestamp of the most recent activity from this asset (updated on each execution)';
COMMENT ON COLUMN assets.is_active IS 'Whether this is the currently active asset for this hostname. Only one asset per hostname should have is_active=TRUE at any time.';
COMMENT ON COLUMN assets.created_at IS 'Timestamp when this asset record was created in the database';
COMMENT ON COLUMN assets.deactivated_at IS 'Timestamp when this asset was deactivated (e.g., when a replacement asset with the same hostname was registered). NULL if still active.';

-- Migrate existing data from executions table
-- Strategy: For each unique (hostname, machine_id) combination, create an asset record
-- - first_seen = earliest execution timestamp
-- - last_seen = latest execution timestamp  
-- - is_active = TRUE if this is the most recent machine_id for this hostname
INSERT INTO assets (hostname, machine_id, first_seen, last_seen, is_active, created_at)
SELECT 
    hostname,
    machine_id,
    MIN(executed_at) as first_seen,
    MAX(executed_at) as last_seen,
    -- Asset is active if it's the most recent machine_id for this hostname
    CASE 
        WHEN machine_id = (
            SELECT e2.machine_id 
            FROM executions e2 
            WHERE e2.hostname = e1.hostname 
            ORDER BY e2.executed_at DESC 
            LIMIT 1
        ) THEN TRUE
        ELSE FALSE
    END as is_active,
    CURRENT_TIMESTAMP as created_at
FROM executions e1
GROUP BY hostname, machine_id
ON CONFLICT (hostname, machine_id) DO NOTHING;

-- Set deactivated_at for inactive assets based on when they were last seen
-- This helps identify when the asset was replaced
UPDATE assets
SET deactivated_at = last_seen
WHERE is_active = FALSE AND deactivated_at IS NULL;
