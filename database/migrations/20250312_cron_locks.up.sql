CREATE TABLE cron_locks (
    task_name VARCHAR(255) PRIMARY KEY,
    locked BOOLEAN NOT NULL DEFAULT FALSE,
    locked_at TIMESTAMP WITH TIME ZONE
);
