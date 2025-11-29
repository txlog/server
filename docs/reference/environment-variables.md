# Environment Variables Reference

Configuration reference for Txlog Server.

## General

| Variable | Default | Description |
| :--- | :--- | :--- |
| `INSTANCE` | - | Name of the environment (e.g., "Production"). Displayed in UI. |
| `LOG_LEVEL` | `INFO` | Logging verbosity (`DEBUG`, `INFO`, `WARN`, `ERROR`). |
| `GIN_MODE` | `release` | Gin framework mode (`debug`, `release`). |
| `PORT` | `8080` | HTTP port to listen on. |

## Database

| Variable | Required | Description |
| :--- | :--- | :--- |
| `PGSQL_HOST` | Yes | Database hostname or IP. |
| `PGSQL_PORT` | Yes | Database port (default `5432`). |
| `PGSQL_USER` | Yes | Database username. |
| `PGSQL_PASSWORD` | Yes | Database password. |
| `PGSQL_DB` | Yes | Database name. |
| `PGSQL_SSLMODE` | Yes | SSL mode (`disable`, `require`, `verify-full`). |

## Authentication (OIDC)

| Variable | Required | Description |
| :--- | :--- | :--- |
| `OIDC_ISSUER_URL` | No | OIDC Provider URL (e.g., `https://accounts.google.com`). |
| `OIDC_CLIENT_ID` | No | Client ID from provider. |
| `OIDC_CLIENT_SECRET` | No | Client Secret from provider. |
| `OIDC_REDIRECT_URL` | No | Callback URL (must match provider config). |
| `OIDC_SKIP_TLS_VERIFY`| No | Set `true` to skip TLS checks (dev only). |

## Authentication (LDAP)

| Variable | Required | Description |
| :--- | :--- | :--- |
| `LDAP_HOST` | No | LDAP server hostname. |
| `LDAP_PORT` | No | LDAP port (`389` or `636`). |
| `LDAP_USE_TLS` | No | `true` for LDAPS. |
| `LDAP_BIND_DN` | No | Service account DN. |
| `LDAP_BIND_PASSWORD` | No | Service account password. |
| `LDAP_BASE_DN` | No | Base DN for user search. |
| `LDAP_USER_FILTER` | No | Filter for users (e.g., `(uid=%s)`). |
| `LDAP_ADMIN_GROUP` | No | DN of admin group. |
| `LDAP_VIEWER_GROUP` | No | DN of viewer group. |

## Scheduler & Retention

| Variable | Default | Description |
| :--- | :--- | :--- |
| `CRON_RETENTION_DAYS` | `7` | Days to keep execution history. |
| `CRON_RETENTION_EXPRESSION` | `0 2 * * *` | Cron schedule for cleanup job. |
| `CRON_STATS_EXPRESSION` | `0 * * * *` | Cron schedule for statistics calculation. |
