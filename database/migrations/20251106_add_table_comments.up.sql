-- Add descriptive comments to all existing tables and columns that don't have them yet
-- This improves database documentation and makes schema self-documenting

-- ============================================================
-- TRANSACTIONS TABLE
-- ============================================================
COMMENT ON TABLE transactions IS 'Package management transactions recorded by DNF/YUM. Each transaction represents a single package manager operation (install, update, remove) on a specific machine.';

COMMENT ON COLUMN transactions.transaction_id IS 'Transaction ID from the package manager (from DNF/YUM history). Not globally unique - must be combined with machine_id.';
COMMENT ON COLUMN transactions.machine_id IS 'Physical/OS installation identifier (UUID from /etc/machine-id). Links to asset machine_id.';
COMMENT ON COLUMN transactions.hostname IS 'Hostname of the machine where the transaction occurred. Used for display and grouping.';
COMMENT ON COLUMN transactions.begin_time IS 'When the package manager transaction started';
COMMENT ON COLUMN transactions.end_time IS 'When the package manager transaction completed';
COMMENT ON COLUMN transactions.actions IS 'Summary of actions performed (e.g., "Install 3, Update 5, Remove 2")';
COMMENT ON COLUMN transactions.altered IS 'Number of packages affected by this transaction';
COMMENT ON COLUMN transactions."user" IS 'System user who executed the package manager command';
COMMENT ON COLUMN transactions.return_code IS 'Exit code from the package manager (0 = success)';
COMMENT ON COLUMN transactions.release_version IS 'OS release version (e.g., "AlmaLinux 9.4")';
COMMENT ON COLUMN transactions.command_line IS 'Full command line that triggered the transaction (e.g., "dnf update kernel")';
COMMENT ON COLUMN transactions.comment IS 'Optional comment or reason for the transaction';
COMMENT ON COLUMN transactions.scriptlet_output IS 'Output from package installation/removal scripts (scriptlets)';

-- ============================================================
-- TRANSACTION_ITEMS TABLE
-- ============================================================
COMMENT ON TABLE transaction_items IS 'Individual packages affected by each transaction. Links to transactions table to show what was installed/updated/removed.';

COMMENT ON COLUMN transaction_items.item_id IS 'Unique identifier for this transaction item (surrogate key)';
COMMENT ON COLUMN transaction_items.transaction_id IS 'References the parent transaction in transactions table';
COMMENT ON COLUMN transaction_items.machine_id IS 'Machine identifier from the parent transaction (denormalized for query performance)';
COMMENT ON COLUMN transaction_items.action IS 'Action performed on this package (Install, Update, Upgraded, Removed, Erase, Reinstall, Downgrade, Obsoleted)';
COMMENT ON COLUMN transaction_items.package IS 'Package name (e.g., "kernel", "httpd", "postgresql")';
COMMENT ON COLUMN transaction_items.version IS 'Package version number (e.g., "5.14.0")';
COMMENT ON COLUMN transaction_items.release IS 'Package release number (e.g., "362.18.1.el9_3")';
COMMENT ON COLUMN transaction_items.epoch IS 'Package epoch number for version comparison (often 0 or empty)';
COMMENT ON COLUMN transaction_items.arch IS 'CPU architecture (x86_64, aarch64, noarch, i686)';
COMMENT ON COLUMN transaction_items.repo IS 'Repository where the package came from (e.g., "baseos", "appstream", "epel")';
COMMENT ON COLUMN transaction_items.from_repo IS 'For updates/downgrades, the repository of the previous version';

-- ============================================================
-- EXECUTIONS TABLE
-- ============================================================
COMMENT ON TABLE executions IS 'Execution history of the txlog-agent on managed assets. Records each time the agent runs to collect and send transaction data to the server.';

COMMENT ON COLUMN executions.id IS 'Unique identifier for this execution record (surrogate key)';
COMMENT ON COLUMN executions.machine_id IS 'Physical/OS installation identifier (UUID from /etc/machine-id). Links to assets table.';
COMMENT ON COLUMN executions.hostname IS 'Hostname of the machine where the agent executed (denormalized for query performance)';
COMMENT ON COLUMN executions.executed_at IS 'Timestamp when the agent execution occurred';
COMMENT ON COLUMN executions.success IS 'Whether the agent execution completed successfully without errors';
COMMENT ON COLUMN executions.details IS 'Additional details about the execution (error messages, warnings, status information)';
COMMENT ON COLUMN executions.transactions_processed IS 'Number of transactions found and processed by the agent during this execution';
COMMENT ON COLUMN executions.transactions_sent IS 'Number of transactions successfully sent to the server during this execution';

-- ============================================================
-- CRON_LOCK TABLE
-- ============================================================
COMMENT ON TABLE cron_lock IS 'Distributed lock mechanism for scheduled jobs. Prevents multiple instances of the same cron job from running simultaneously.';

COMMENT ON COLUMN cron_lock.job_name IS 'Unique identifier for the scheduled job (e.g., "housekeeping", "statistics_update")';
COMMENT ON COLUMN cron_lock.locked_at IS 'Timestamp when the lock was acquired. Used to detect stale locks.';

-- ============================================================
-- STATISTICS TABLE
-- ============================================================
COMMENT ON TABLE statistics IS 'Cached statistics and metrics for dashboard display. Updated periodically by scheduled jobs to avoid expensive real-time queries.';

COMMENT ON COLUMN statistics.name IS 'Unique identifier for this statistic (e.g., "total_executions", "total_packages_installed")';
COMMENT ON COLUMN statistics.value IS 'Current numeric value of the statistic';
COMMENT ON COLUMN statistics.percentage IS 'Optional percentage value (e.g., growth rate, completion rate). Scale: 0.00 to 999.99';
COMMENT ON COLUMN statistics.updated_at IS 'Timestamp when this statistic was last calculated and updated';

-- ============================================================
-- USERS TABLE
-- ============================================================
COMMENT ON TABLE users IS 'User accounts for web interface authentication. Populated when using OIDC/LDAP authentication. Stores user profile and authorization information.';

COMMENT ON COLUMN users.id IS 'Unique identifier for this user (surrogate key)';
COMMENT ON COLUMN users.sub IS 'OIDC Subject identifier - unique user ID from the identity provider (e.g., "auth0|123456")';
COMMENT ON COLUMN users.email IS 'User email address from the identity provider. Must be unique.';
COMMENT ON COLUMN users.name IS 'User full name or display name from the identity provider';
COMMENT ON COLUMN users.picture IS 'URL to user profile picture/avatar from the identity provider';
COMMENT ON COLUMN users.is_active IS 'Whether the user account is active and can log in. Set to false to revoke access.';
COMMENT ON COLUMN users.is_admin IS 'Whether the user has administrative privileges (can manage other users, API keys, system settings)';
COMMENT ON COLUMN users.created_at IS 'Timestamp when the user account was created (first login)';
COMMENT ON COLUMN users.updated_at IS 'Timestamp when the user profile was last updated from the identity provider';
COMMENT ON COLUMN users.last_login_at IS 'Timestamp of the user most recent successful login';

-- ============================================================
-- USER_SESSIONS TABLE
-- ============================================================
COMMENT ON TABLE user_sessions IS 'Active user sessions for web interface. Manages session lifecycle and expiration for logged-in users.';

COMMENT ON COLUMN user_sessions.id IS 'Session ID stored in browser cookie (random secure token)';
COMMENT ON COLUMN user_sessions.user_id IS 'References the user account that owns this session';
COMMENT ON COLUMN user_sessions.created_at IS 'Timestamp when the session was created (at login)';
COMMENT ON COLUMN user_sessions.expires_at IS 'Timestamp when the session will expire and require re-authentication';
COMMENT ON COLUMN user_sessions.is_active IS 'Whether the session is currently valid. Set to false on logout or security events.';

-- ============================================================
-- API_KEYS TABLE (already has some comments, adding missing ones)
-- ============================================================
COMMENT ON COLUMN api_keys.id IS 'Unique identifier for this API key record (surrogate key)';
COMMENT ON COLUMN api_keys.name IS 'Descriptive name for the API key (e.g., "Production Agent", "CI/CD Pipeline"). Minimum 3 characters.';
COMMENT ON COLUMN api_keys.created_at IS 'Timestamp when the API key was created';
COMMENT ON COLUMN api_keys.is_active IS 'Whether the API key is currently valid and can be used for authentication. Set to false to revoke access.';
COMMENT ON COLUMN api_keys.created_by IS 'References the user who created this API key. NULL if created by system or user was deleted.';
