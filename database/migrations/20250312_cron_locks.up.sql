CREATE TABLE cron_lock
(
    job_name VARCHAR(255) PRIMARY KEY,
    locked_at TIMESTAMP WITH TIME ZONE
);
