# How to Run Database Migrations

Database migrations are essential when upgrading the Txlog Server to a new version that includes schema changes.

## Method 1: Automatic on Startup (Not Default)

Currently, the server does **not** automatically run migrations on startup to prevent accidental schema changes in
production. You must trigger them manually.

## Method 2: Via Admin Panel (Recommended)

This is the safest and easiest way to apply migrations.

1. Log in to the **Admin Panel** (`/admin`) as an administrator.
2. Locate the **Migration Status** section.
    - It will show the "Current Version" and a list of "Pending Migrations".
3. If there are pending migrations, a **Run Migrations** button will be visible.
4. Click **Run Migrations**.
5. The page will reload, and the status should show all migrations as "Applied".

## Troubleshooting

- **"Dirty Database"**: If a migration fails halfway, the database might be marked as "dirty". You may need to manually
  fix the schema or update the `schema_migrations` table in the database to resolve this.
