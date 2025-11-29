# How to Configure Data Retention

Txlog Server includes a built-in scheduler to automatically clean up old execution data. This helps manage database size
and performance.

## Configuration Variables

The retention policy is controlled by two environment variables:

### 1. `CRON_RETENTION_DAYS`

Defines how many days of data to keep.

- **Default**: `7`
- **Example**: Set to `30` to keep data for a month.

### 2. `CRON_RETENTION_EXPRESSION`

Defines **when** the cleanup job runs (Cron syntax).

- **Default**: `0 2 * * *` (Every day at 2:00 AM)
- **Format**: `Minute Hour Day Month DayOfWeek`

## Steps to Configure

1. Open your `.env` file (or deployment configuration).
2. Add or update the variables:

    ```bash
    # Keep data for 30 days
    CRON_RETENTION_DAYS=30

    # Run cleanup every Sunday at 3:00 AM
    CRON_RETENTION_EXPRESSION=0 3 * * 0
    ```

3. Restart the server for changes to take effect.

## Verifying the Job

Check the server logs for the "Housekeeping" task. You should see entries like:

```text
[INFO] Housekeeping: executing task...
[INFO] Housekeeping: executions older than 30 days are deleted.
```
