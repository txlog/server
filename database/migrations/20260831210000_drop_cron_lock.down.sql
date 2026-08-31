CREATE TABLE IF NOT EXISTS cron_lock
(
    job_name VARCHAR(255) PRIMARY KEY,
    locked_at TIMESTAMP WITH TIME ZONE
);

COMMENT ON TABLE cron_lock IS 'Distributed lock mechanism for scheduled jobs. Prevents multiple instances of the same cron job from running simultaneously.';
COMMENT ON COLUMN cron_lock.job_name IS 'Unique identifier for the scheduled job (e.g., "housekeeping", "statistics_update")';
COMMENT ON COLUMN cron_lock.locked_at IS 'Timestamp when the lock was acquired. Used to detect stale locks.';
