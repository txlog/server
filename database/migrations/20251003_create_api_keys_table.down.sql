-- Drop indexes
DROP INDEX IF EXISTS idx_api_keys_created_at;
DROP INDEX IF EXISTS idx_api_keys_key_hash;

-- Drop api_keys table
DROP TABLE IF EXISTS api_keys;
