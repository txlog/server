-- Scheduled jobs coordinate through PostgreSQL advisory locks rather than this
-- table. An advisory lock belongs to the session that took it, so it is released
-- automatically when a crashed instance's backend goes away -- which is what the
-- table could not do, and why it needed a reaper for locks older than 12 hours.
DROP TABLE IF EXISTS cron_lock;
