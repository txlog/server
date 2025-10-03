-- Create api_keys table for API authentication
CREATE TABLE IF NOT EXISTS api_keys (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(64) NOT NULL UNIQUE,
    key_prefix VARCHAR(16) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT api_keys_name_check CHECK (char_length(name) >= 3)
);

-- Create index for fast key lookup
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash) WHERE is_active = true;

-- Create index for admin panel queries
CREATE INDEX idx_api_keys_created_at ON api_keys(created_at DESC);

-- Add comment to table
COMMENT ON TABLE api_keys IS 'API keys for authenticating API requests to /v1 endpoints';
COMMENT ON COLUMN api_keys.key_hash IS 'SHA-256 hash of the actual API key';
COMMENT ON COLUMN api_keys.key_prefix IS 'First 8 characters of key for identification (e.g., txlog_ab)';
COMMENT ON COLUMN api_keys.last_used_at IS 'Timestamp of last successful authentication with this key';
