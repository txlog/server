-- Remove all comments added by the 20251106_add_table_comments migration
-- This allows rolling back the documentation changes if needed

-- ============================================================
-- TRANSACTIONS TABLE
-- ============================================================
COMMENT ON TABLE transactions IS NULL;
COMMENT ON COLUMN transactions.transaction_id IS NULL;
COMMENT ON COLUMN transactions.machine_id IS NULL;
COMMENT ON COLUMN transactions.hostname IS NULL;
COMMENT ON COLUMN transactions.begin_time IS NULL;
COMMENT ON COLUMN transactions.end_time IS NULL;
COMMENT ON COLUMN transactions.actions IS NULL;
COMMENT ON COLUMN transactions.altered IS NULL;
COMMENT ON COLUMN transactions."user" IS NULL;
COMMENT ON COLUMN transactions.return_code IS NULL;
COMMENT ON COLUMN transactions.release_version IS NULL;
COMMENT ON COLUMN transactions.command_line IS NULL;
COMMENT ON COLUMN transactions.comment IS NULL;
COMMENT ON COLUMN transactions.scriptlet_output IS NULL;

-- ============================================================
-- TRANSACTION_ITEMS TABLE
-- ============================================================
COMMENT ON TABLE transaction_items IS NULL;
COMMENT ON COLUMN transaction_items.item_id IS NULL;
COMMENT ON COLUMN transaction_items.transaction_id IS NULL;
COMMENT ON COLUMN transaction_items.machine_id IS NULL;
COMMENT ON COLUMN transaction_items.action IS NULL;
COMMENT ON COLUMN transaction_items.package IS NULL;
COMMENT ON COLUMN transaction_items.version IS NULL;
COMMENT ON COLUMN transaction_items.release IS NULL;
COMMENT ON COLUMN transaction_items.epoch IS NULL;
COMMENT ON COLUMN transaction_items.arch IS NULL;
COMMENT ON COLUMN transaction_items.repo IS NULL;
COMMENT ON COLUMN transaction_items.from_repo IS NULL;

-- ============================================================
-- EXECUTIONS TABLE
-- ============================================================
COMMENT ON TABLE executions IS NULL;
COMMENT ON COLUMN executions.id IS NULL;
COMMENT ON COLUMN executions.machine_id IS NULL;
COMMENT ON COLUMN executions.hostname IS NULL;
COMMENT ON COLUMN executions.executed_at IS NULL;
COMMENT ON COLUMN executions.success IS NULL;
COMMENT ON COLUMN executions.details IS NULL;
COMMENT ON COLUMN executions.transactions_processed IS NULL;
COMMENT ON COLUMN executions.transactions_sent IS NULL;

-- ============================================================
-- CRON_LOCK TABLE
-- ============================================================
COMMENT ON TABLE cron_lock IS NULL;
COMMENT ON COLUMN cron_lock.job_name IS NULL;
COMMENT ON COLUMN cron_lock.locked_at IS NULL;

-- ============================================================
-- STATISTICS TABLE
-- ============================================================
COMMENT ON TABLE statistics IS NULL;
COMMENT ON COLUMN statistics.name IS NULL;
COMMENT ON COLUMN statistics.value IS NULL;
COMMENT ON COLUMN statistics.percentage IS NULL;
COMMENT ON COLUMN statistics.updated_at IS NULL;

-- ============================================================
-- USERS TABLE
-- ============================================================
COMMENT ON TABLE users IS NULL;
COMMENT ON COLUMN users.id IS NULL;
COMMENT ON COLUMN users.sub IS NULL;
COMMENT ON COLUMN users.email IS NULL;
COMMENT ON COLUMN users.name IS NULL;
COMMENT ON COLUMN users.picture IS NULL;
COMMENT ON COLUMN users.is_active IS NULL;
COMMENT ON COLUMN users.is_admin IS NULL;
COMMENT ON COLUMN users.created_at IS NULL;
COMMENT ON COLUMN users.updated_at IS NULL;
COMMENT ON COLUMN users.last_login_at IS NULL;

-- ============================================================
-- USER_SESSIONS TABLE
-- ============================================================
COMMENT ON TABLE user_sessions IS NULL;
COMMENT ON COLUMN user_sessions.id IS NULL;
COMMENT ON COLUMN user_sessions.user_id IS NULL;
COMMENT ON COLUMN user_sessions.created_at IS NULL;
COMMENT ON COLUMN user_sessions.expires_at IS NULL;
COMMENT ON COLUMN user_sessions.is_active IS NULL;

-- ============================================================
-- API_KEYS TABLE
-- ============================================================
COMMENT ON COLUMN api_keys.id IS NULL;
COMMENT ON COLUMN api_keys.name IS NULL;
COMMENT ON COLUMN api_keys.created_at IS NULL;
COMMENT ON COLUMN api_keys.is_active IS NULL;
COMMENT ON COLUMN api_keys.created_by IS NULL;
