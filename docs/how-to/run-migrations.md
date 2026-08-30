# How to Run Database Migrations

Database migrations are essential when upgrading the Txlog Server to a new version that includes schema changes.

## Method 1: Automatic on Startup (Default)

The server applies every pending migration when it starts. `ConnectDatabase()` runs them before serving traffic, so
deploying a new version is enough to bring the schema up to date; there is no manual step. If the database is marked
dirty from an earlier failure, startup forces it back to its current version first and then continues.

Failures are logged and do **not** stop the server, so check the startup log after deploying a version that changes the
schema.

## Method 2: Via Admin Panel

Useful when the schema changed but restarting the server is inconvenient, or to inspect what is pending.

1. Log in to the **Admin Panel** (`/admin`) as an administrator.
2. Locate the **Migration Status** section.
   - It will show the "Current Version" and a list of "Pending Migrations".
3. If there are pending migrations, a **Run Migrations** button will be visible.
4. Click **Run Migrations**.
5. The page will reload, and the status should show all migrations as "Applied".

## Writing a New Migration

Migrations are plain SQL files under `database/migrations/`, embedded into the binary with `go:embed` and applied by
[golang-migrate](https://github.com/golang-migrate/migrate). Add them in `.up.sql`/`.down.sql` pairs, and never edit a
migration that has already shipped — it has been applied elsewhere and its version is already recorded.

Name the pair with a full 14-digit `YYYYMMDDHHMMSS` prefix:

```text
20260901143000_add_widget_table.up.sql
20260901143000_add_widget_table.down.sql
```

The prefix is the migration's **version number**, read as an integer rather than as text. A shorter prefix yields a far
smaller number — `20260901` is 20,260,901, while `20260309201502` is 20,260,309,201,502 — so an 8-digit migration added
today would sort *before* every 14-digit migration already applied, and golang-migrate would treat it as run and skip it
in silence.

Some older files use 8- and 12-digit prefixes from before this rule. Their ranges do not overlap the 14-digit ones, so
the existing order is correct, but do not copy them.

## Troubleshooting

- **"Dirty Database"**: If a migration fails halfway, the database might be marked as "dirty". You may need to manually
  fix the schema or update the `schema_migrations` table in the database to resolve this.
