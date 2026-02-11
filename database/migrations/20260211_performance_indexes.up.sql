-- Performance indexes for most frequent query patterns
-- D1: Composite index for the most common JOIN (transaction_items ↔ transactions)
CREATE INDEX IF NOT EXISTS idx_ti_txid_machineid
    ON transaction_items (transaction_id, machine_id);

-- D2: Composite index for queries filtering by machine_id + hostname with temporal ordering
-- Used by: getAssetsByOS, getAssetsByAgentVersion, GetMachineID, GetMachines
CREATE INDEX IF NOT EXISTS idx_executions_machine_hostname_time
    ON executions (machine_id, hostname, executed_at DESC);

-- D3: Covering index for action-based filters (statistics, package progression)
-- INCLUDE avoids table lookups for covered columns
CREATE INDEX IF NOT EXISTS idx_ti_action_covering
    ON transaction_items (action)
    INCLUDE (transaction_id, machine_id, package);
