# Txlog Server — Codebase Research Report

## 1. Project Overview

**Txlog Server** is a centralized Go web application that receives, stores, and
visualizes RPM package transaction data from distributed **Txlog Agent** instances
running on Linux servers. It acts as a single source of truth for tracking what
software packages were installed, upgraded, downgraded, or removed across an
entire fleet of machines.

| Attribute         | Detail                                              |
|-------------------|-----------------------------------------------------|
| Language          | Go 1.26.0                                           |
| Web Framework     | Gin (`github.com/gin-gonic/gin` v1.11.0)            |
| Database          | PostgreSQL (via `lib/pq`)                           |
| Current Version   | 1.21.0                                              |
| Container Image   | `ghcr.io/txlog/server`                              |
| License           | MIT                                                 |
| Base Docker Image | `scratch` (with `txlog` user from AlmaLinux 10)     |

---

## 2. Architecture

### 2.1 High-Level Flow

```text
┌──────────────┐     POST /v1/transactions     ┌────────────────┐
│  Txlog Agent │ ──────────────────────────────▶│                │
│  (on server) │     POST /v1/executions        │  Txlog Server  │
│              │ ──────────────────────────────▶│   (Gin HTTP)   │
└──────────────┘                                │                │
                                                │  ┌──────────┐  │
       ┌───────────────────────────────────────▶│  │ PostgreSQL│  │
       │  Browser (Dashboard, Admin, Analytics) │  └──────────┘  │
       │                                        └────────────────┘
       │
  ┌────┴─────┐
  │  User /  │
  │  Admin   │
  └──────────┘
```

Agents periodically run on managed servers, collect `dnf`/`yum` transaction
history, and POST the data to the server's `/v1/transactions` and
`/v1/executions` endpoints. The server stores everything in PostgreSQL and
exposes a web dashboard, analytics pages, and a Swagger-documented REST API.

### 2.2 Package Structure

```text
txlog/server/
├── main.go              # Entry point: routing, middleware, template setup
├── auth/                # OIDC and LDAP authentication services
│   ├── oidc.go          # OpenID Connect (go-oidc/v3, oauth2)
│   └── ldap.go          # LDAP authentication (go-ldap/v3)
├── controllers/         # Web (HTML) controllers
│   ├── api/v1/          # REST API v1 controllers (JSON)
│   ├── root_controller  # Dashboard (/)
│   ├── assets_controller# Asset listing (/assets, /assets/:id)
│   ├── packages_*       # Package listing and progression
│   ├── admin_controller # Admin panel (/admin)
│   ├── auth_controller  # Login/logout flows
│   └── analytics_*      # Analytics pages
├── database/            # Database connection and migration management
│   ├── main.go          # ConnectDatabase(), connection pool config
│   ├── migrations.go    # Migration status and execution helpers
│   └── migrations/      # 19 SQL migration pairs (.up.sql / .down.sql)
├── middleware/           # Gin middleware
│   ├── auth.go          # Session-based authentication + admin check
│   └── api_key.go       # API key validation for /v1 endpoints
├── models/              # Data structures (no ORM)
├── scheduler/           # Cron-based background jobs
├── statistics/          # 30-day statistical aggregations
├── util/                # Template functions, health checks, API key gen
├── logger/              # slog-based structured logging
├── version/             # SemVer (injected via ldflags at build time)
├── templates/           # 19 Go HTML templates (Tabler UI framework)
├── images/              # SVG brand logos + favicon
├── docs/                # Swagger spec (auto-generated) + documentation
└── tests/               # Integration test suite
```

---

## 3. Data Model

The server uses raw SQL queries (no ORM) with `database/sql` and `lib/pq`.

### 3.1 Core Tables

#### `transactions` (composite PK: `transaction_id`, `machine_id`)

Stores DNF/YUM transaction history. Each transaction represents one
`dnf install/upgrade/remove` operation on a machine.

| Column             | Type                     | Purpose                          |
|--------------------|--------------------------|----------------------------------|
| `transaction_id`   | INTEGER                  | Transaction sequence number      |
| `machine_id`       | TEXT                     | `/etc/machine-id` UUID           |
| `hostname`         | TEXT                     | Logical hostname                 |
| `begin_time`       | TIMESTAMP WITH TIME ZONE | Transaction start                |
| `end_time`         | TIMESTAMP WITH TIME ZONE | Transaction end                  |
| `actions`          | TEXT                     | Action codes (I,U,D,E,R,O,C)    |
| `altered`          | TEXT                     | Number of altered packages       |
| `user`             | TEXT                     | User who ran the command         |
| `return_code`      | TEXT                     | Exit code                        |
| `release_version`  | TEXT                     | OS release version               |
| `command_line`     | TEXT                     | Full command line                |
| `comment`          | TEXT                     | Transaction comment              |
| `scriptlet_output` | TEXT                     | Scriptlet output                 |

#### `transaction_items` (PK: auto-increment `item_id`)

Line items within a transaction — one row per package affected.

| Column           | Type    | Purpose                                |
|------------------|---------|----------------------------------------|
| `transaction_id` | INTEGER | FK → transactions                      |
| `machine_id`     | TEXT    | FK → transactions (composite)          |
| `action`         | TEXT    | Install, Upgrade, Downgrade, etc.      |
| `package`        | TEXT    | RPM package name                       |
| `version`        | TEXT    | Package version                        |
| `release`        | TEXT    | Package release                        |
| `epoch`          | TEXT    | Package epoch                          |
| `arch`           | TEXT    | Architecture (x86_64, noarch, etc.)    |
| `repo`           | TEXT    | Target repository                      |
| `from_repo`      | TEXT    | Source repository (for upgrades)       |

#### `executions` (PK: auto-increment `id`)

Agent run logs — records each time the agent successfully (or unsuccessfully)
executed and sent data.

| Column                   | Type      | Purpose                           |
|--------------------------|-----------|-----------------------------------|
| `machine_id`             | TEXT      | Machine UUID                      |
| `hostname`               | TEXT      | Logical hostname                  |
| `executed_at`            | TIMESTAMP | Execution timestamp               |
| `success`                | BOOLEAN   | Whether the agent run succeeded   |
| `details`                | TEXT      | Error details (if failed)         |
| `transactions_processed` | INTEGER   | Transactions found on the machine |
| `transactions_sent`      | INTEGER   | Transactions sent to the server   |
| `agent_version`          | TEXT      | Agent version string              |
| `os`                     | TEXT      | OS release string                 |
| `needs_restarting`       | BOOLEAN   | Whether the server needs a reboot |
| `restarting_reason`      | TEXT      | Reason for required restart       |

#### `assets` (PK: auto-increment `asset_id`)

Central registry of managed machines. Tracks both logical identity (hostname)
and physical identity (machine_id) to handle OS reinstalls.

| Column              | Type      | Purpose                              |
|---------------------|-----------|--------------------------------------|
| `hostname`          | TEXT      | Logical server name                  |
| `machine_id`        | TEXT      | Physical installation UUID           |
| `first_seen`        | TIMESTAMP | First agent report                   |
| `last_seen`         | TIMESTAMP | Most recent agent report             |
| `is_active`         | BOOLEAN   | Current active asset for hostname    |
| `deactivated_at`    | TIMESTAMP | When this asset was replaced         |
| `needs_restarting`  | BOOLEAN   | Reboot required flag                 |
| `restarting_reason` | TEXT      | Reboot reason                        |
| `os`                | TEXT      | OS release string                    |
| `agent_version`     | TEXT      | Agent version string                 |

**Key design**: Only one asset per hostname is active at a time. When a new
`machine_id` is seen for the same hostname (e.g., after OS reinstall), the old
asset is deactivated and the new one becomes active. This is handled by the
`AssetManager.UpsertAsset()` method.

### 3.2 Supporting Tables

| Table            | Purpose                                               |
|------------------|-------------------------------------------------------|
| `users`          | OIDC/LDAP authenticated users with admin/active flags |
| `user_sessions`  | Session tokens with expiration (7-day TTL)            |
| `api_keys`       | SHA-256-hashed API keys with `txlog_` prefix          |
| `statistics`     | Cached 30-day metric aggregations                     |
| `cron_lock`      | Distributed lock table for scheduler jobs             |

### 3.3 Materialized Views

| View                        | Purpose                                  | Refresh Interval |
|-----------------------------|------------------------------------------|------------------|
| `mv_package_listing`        | Pre-computed package catalog             | Every 5 minutes  |
| `mv_dashboard_os_stats`     | OS distribution for dashboard            | Every 5 minutes  |
| `mv_dashboard_agent_stats`  | Agent version distribution for dashboard | Every 5 minutes  |
| `mv_dashboard_most_updated` | Most updated packages for dashboard      | Every 5 minutes  |

---

## 4. Authentication & Authorization

The server supports **three authentication modes** that can operate independently
or simultaneously:

### 4.1 No Authentication (Default)

When neither OIDC nor LDAP environment variables are set, the server runs
without any authentication. All web pages and API endpoints are fully accessible.

### 4.2 OIDC (OpenID Connect)

- Provider discovery via `OIDC_ISSUER_URL`
- Standard authorization code flow with PKCE state parameter
- Automatic user creation/update from ID token claims (`sub`, `email`, `name`,
  `picture`)
- First OIDC user is auto-promoted to administrator
- Supports TLS certificate verification skip for self-signed certs
- Uses `go-oidc/v3` and `golang.org/x/oauth2`

### 4.3 LDAP

- Bind + search authentication pattern
- Supports both service account bind and anonymous bind
- Group-based authorization: `LDAP_ADMIN_GROUP` and `LDAP_VIEWER_GROUP`
- Users are created/updated in local `users` table on each login
- UID extraction from Distinguished Name
- Uses `go-ldap/v3`

### 4.4 API Key Authentication

When authentication is enabled, `/v1/*` API endpoints require an API key:

- Keys have format `txlog_{base64url_random_string}`
- Stored as SHA-256 hashes (never in plaintext)
- Validated via `X-API-Key` header or `Authorization: Bearer` header
- Session-authenticated users can also access `/v1` endpoints directly
- Keys are managed through the admin panel (create/revoke/delete)
- `last_used_at` is updated asynchronously (non-blocking goroutine)

### 4.5 Middleware Stack

```
Request → EnvironmentVariablesMiddleware → AuthMiddleware → [AdminMiddleware] → Handler
                                                    │
                                            APIKeyMiddleware (for /v1 only)
```

- `AuthMiddleware`: Validates session cookies, redirects to `/login` if invalid.
  Skips `/v1/`, `/health`, `/auth/`, `/images/`, `/login`.
- `AdminMiddleware`: Checks `user.IsAdmin` flag for `/admin/*` routes.
- `APIKeyMiddleware`: Validates API keys or session cookies for `/v1/*` routes.

---

## 5. REST API (v1)

All API endpoints are under `/v1/` with Swagger documentation at
`/swagger/index.html`.

### 5.1 Data Ingestion Endpoints (Agent → Server)

| Method | Endpoint                 | Purpose                                |
|--------|--------------------------|----------------------------------------|
| POST   | `/v1/transactions`       | Submit transaction data (batch insert) |
| POST   | `/v1/executions`         | Submit agent execution report          |
| GET    | `/v1/transactions/ids`   | Get existing transaction IDs for host  |
| GET    | `/v1/version`            | Get server version                     |

**Transaction ingestion** uses a PostgreSQL database transaction with batch
`COPY`-style inserts via `pq.CopyIn()` for performance. Existing transactions
are skipped (idempotent). The `PostExecutions` endpoint also calls
`AssetManager.UpsertAsset()` to maintain the asset registry.

### 5.2 Query Endpoints

| Method | Endpoint                                          | Purpose                        |
|--------|---------------------------------------------------|--------------------------------|
| GET    | `/v1/machines/ids`                                | List machine IDs               |
| GET    | `/v1/machines`                                    | List machines with details     |
| GET    | `/v1/executions`                                  | List executions (paginated)    |
| GET    | `/v1/transactions`                                | List transactions (paginated)  |
| GET    | `/v1/items/ids`                                   | List transaction item IDs      |
| GET    | `/v1/items`                                       | List transaction items         |
| GET    | `/v1/assets/requiring-restart`                    | Assets needing reboot          |
| GET    | `/v1/packages/:name/:version/:release/assets`     | Assets using specific package  |

### 5.3 Reports & Analytics Endpoints

| Method | Endpoint                          | Purpose                           |
|--------|-----------------------------------|-----------------------------------|
| GET    | `/v1/reports/monthly`             | Monthly package update report     |
| GET    | `/v1/reports/compare-packages`    | Compare packages across assets    |
| GET    | `/v1/reports/package-freshness`   | Package version age analysis      |
| GET    | `/v1/reports/package-adoption`    | Package adoption across fleet     |
| GET    | `/v1/reports/anomalies`           | Transaction anomaly detection     |

**Anomaly detection** identifies three types of anomalies:

- **High volume**: Transactions with unusually high package counts
- **Rapid change**: Packages changed multiple times in a short window
- **Downgrade**: Package version downgrades

---

## 6. Web Dashboard

The frontend uses the **Tabler** UI framework (Bootstrap-based) with
server-side Go HTML templates. No JavaScript framework; interactivity is
provided by Tabler's built-in components and **ApexCharts** for graphs.

### 6.1 Pages

| Route                      | Template                   | Purpose                                  |
|----------------------------|----------------------------|------------------------------------------|
| `/`                        | `index.html`               | Dashboard: stats, OS/agent distribution  |
| `/assets`                  | `assets.html`              | Paginated asset listing with search      |
| `/assets/:machine_id`      | `machine_id.html`          | Single asset detail with transactions    |
| `/packages`                | `packages.html`            | Package catalog with search              |
| `/packages/:name`          | `package_name.html`        | Package version history                  |
| `/package-progression`     | `packages_by_week.html`    | Weekly install/upgrade graph             |
| `/analytics/compare`       | `analytics_compare.html`   | Package comparison across assets         |
| `/analytics/freshness`     | `analytics_freshness.html` | Package freshness analysis               |
| `/analytics/adoption`      | `analytics_adoption.html`  | Package adoption rates                   |
| `/analytics/anomalies`     | `analytics_anomalies.html` | Anomaly detection results                |
| `/admin`                   | `admin.html`               | Admin panel (users, API keys, migrations)|
| `/login`                   | `login.html`               | Authentication page                      |
| `/insights`                | (static)                   | Insights page                            |
| `/license`                 | `license.html`             | License information                      |
| `/sponsor`                 | `sponsor.html`             | Sponsorship page                         |
| `/executions/:id`          | `execution_id.html`        | Single execution detail                  |

### 6.2 Template Functions

The templates use custom functions registered in `main.go`:

| Function           | Purpose                                                 |
|--------------------|---------------------------------------------------------|
| `brand`            | Returns SVG filename for Linux distro (AlmaLinux, etc.) |
| `formatInteger`    | Thousand separators with dots (Brazilian format)        |
| `formatPercentage` | Decimal with comma, thousands with dots                 |
| `formatDateTime`   | DD/MM/YYYY HH:MM:SS TZD format                         |
| `dnfUser`          | Extracts username from DNF user string                  |
| `hasAction`        | Checks if action exists in comma-separated codes        |
| `timeStatusClass`  | Returns CSS class based on last-seen recency            |
| `text2html`        | Converts plain text to HTML with `<br>` and `&nbsp;`    |
| `version`          | Returns current SemVer                                  |
| `versionsEqual`    | Normalizes and compares version strings                 |

### 6.3 Supported Linux Distributions

Brand logos are included for: **AlmaLinux**, **CentOS**, **Fedora**, **Oracle
Linux**, **Red Hat**, **Rocky Linux**, with a generic Linux fallback.

---

## 7. Background Scheduler

The scheduler uses `github.com/mileusna/crontab` and runs 4 jobs:

| Job                     | Schedule                          | Purpose                                      |
|-------------------------|-----------------------------------|----------------------------------------------|
| `housekeepingJob`       | `CRON_RETENTION_EXPRESSION`       | Delete old executions, clean orphan data     |
| `statsJob`              | `CRON_STATS_EXPRESSION`           | Compute 30-day statistics with % change      |
| `latestVersionJob`      | Every hour (`0 * * * *`)          | Fetch latest version from txlog.rda.run      |
| `refreshMaterializedViewsJob` | Every 5 min (`*/5 * * * *`) | Refresh all materialized views              |

### 7.1 Distributed Locking

All scheduler jobs use a **PostgreSQL-based distributed lock** via the
`cron_lock` table to ensure only one instance runs each job at a time (essential
for Kubernetes multi-replica deployments):

```sql
INSERT INTO cron_lock (job_name, locked_at)
VALUES ($1, NOW())
ON CONFLICT (job_name) DO NOTHING
```

### 7.2 Housekeeping Details

- Deletes executions older than `CRON_RETENTION_DAYS` (default: 7)
- Cleans orphan `transaction_items` and `transactions` from assets deactivated
  for more than 90 days

### 7.3 Statistics

Three metrics are computed with 30-day current period vs 30-day previous
period comparison:

- `executions-30-days`: Total agent executions
- `installed-packages-30-days`: Total package installations
- `upgraded-packages-30-days`: Total package upgrades

---

## 8. Database Migrations

The project uses `golang-migrate/migrate/v4` with **embedded SQL files** (via
Go's `embed` package). Migrations are applied automatically at startup.

### 8.1 Migration History

| Version    | Description                        | Key Changes                               |
|------------|------------------------------------|-------------------------------------------|
| 20250208   | Initial schema                     | transactions, transaction_items, executions|
| 20250312   | Cron locks                         | cron_lock table for distributed locking    |
| 20250320   | Statistics                         | statistics table                           |
| 20250325   | Agent version                      | agent_version column on executions         |
| 20250514   | Needs restarting                   | needs_restarting on executions             |
| 20250604   | Create index                       | Performance index                          |
| 20250813   | Create indexes                     | Performance indexes                        |
| 20250929   | Users tables                       | users, user_sessions tables                |
| 20251003   | API keys table                     | api_keys table                             |
| 20251016   | Performance indexes                | Additional composite indexes               |
| 20251105   | Assets table                       | assets table + data backfill from executions|
| 20251106   | Table comments                     | PostgreSQL COMMENT ON documentation        |
| 20251107   | Assets needs_restarting            | needs_restarting/restarting_reason on assets|
| 202512160002 | Assets OS column                 | os column on assets + backfill             |
| 20251216   | Packages materialized view         | mv_package_listing + GIN trigram index     |
| 20260211   | Performance indexes                | Composite indexes for query optimization   |
| 202602120002 | Dashboard materialized views     | mv_dashboard_os/agent/most_updated views   |
| 20260212   | Agent version on assets            | agent_version on assets + backfill         |
| 20260213   | Autovacuum tuning                  | Custom autovacuum params for large tables  |

### 8.2 Dirty State Handling

The server includes automatic dirty state recovery: if a migration was
partially applied (dirty flag), the server forces the version clean and
retries. This can also be triggered manually from the admin panel.

---

## 9. Performance Optimizations

### 9.1 Database

- **Connection pool**: `MaxOpenConns=25`, `MaxIdleConns=10`,
  `ConnMaxLifetime=5min`, `ConnMaxIdleTime=1min`
- **Batch inserts**: `pq.CopyIn()` for transaction items (COPY protocol)
- **Materialized views**: Pre-computed data for packages and dashboard, refreshed
  every 5 minutes with `CONCURRENTLY` support
- **Composite indexes**: Targeted indexes on frequently-queried columns
- **Autovacuum tuning**: Custom autovacuum settings for large tables
  (`transactions`, `transaction_items`, `executions`)

### 9.2 Application

- **Parallel dashboard queries**: Uses `golang.org/x/sync/errgroup` for
  concurrent database queries on the root page
- **Environment variable caching**: Static env vars are snapshotted once at
  startup with per-request copy to avoid concurrent map writes
- **Async API key tracking**: `last_used_at` is updated in a background
  goroutine to not block API requests
- **Savepoint-based backward compatibility**: `AssetManager.UpsertAsset()` uses
  PostgreSQL SAVEPOINTs to gracefully handle missing columns during migration
  windows

---

## 10. Deployment Model

### 10.1 Docker

- Multi-stage build: AlmaLinux 10 → `scratch`
- Non-root `txlog` user (UID 10001)
- CA certificates copied from AlmaLinux
- Single static binary (`CGO_ENABLED=0`)
- Image signed with Cosign (Sigstore)

### 10.2 Kubernetes

- Supports multi-replica deployments (distributed locks prevent job conflicts)
- Health/readiness probes via `/health` endpoint
- Secrets for database password via `secretKeyRef`

### 10.3 Environment Variables

| Variable                     | Required | Default       | Purpose                     |
|------------------------------|----------|---------------|-----------------------------|
| `INSTANCE`                   | No       |               | Instance name for display   |
| `LOG_LEVEL`                  | No       | INFO          | DEBUG, INFO, WARN, ERROR    |
| `GIN_MODE`                   | No       | release       | Gin framework mode          |
| `PGSQL_HOST`                 | Yes      |               | PostgreSQL host             |
| `PGSQL_PORT`                 | Yes      |               | PostgreSQL port             |
| `PGSQL_USER`                 | Yes      |               | PostgreSQL user             |
| `PGSQL_DB`                   | Yes      |               | PostgreSQL database name    |
| `PGSQL_PASSWORD`             | Yes      |               | PostgreSQL password         |
| `PGSQL_SSLMODE`              | Yes      |               | PostgreSQL SSL mode         |
| `CRON_RETENTION_DAYS`        | No       | 7             | Execution retention days    |
| `CRON_RETENTION_EXPRESSION`  | Yes      |               | Housekeeping cron schedule  |
| `CRON_STATS_EXPRESSION`      | Yes      |               | Statistics cron schedule    |
| `OIDC_ISSUER_URL`            | No       |               | OIDC provider URL           |
| `OIDC_CLIENT_ID`             | No       |               | OIDC client ID              |
| `OIDC_CLIENT_SECRET`         | No       |               | OIDC client secret          |
| `OIDC_REDIRECT_URL`          | No       |               | OIDC callback URL           |
| `OIDC_SKIP_TLS_VERIFY`       | No       | false         | Skip OIDC TLS verification  |
| `LDAP_HOST`                  | No       |               | LDAP server hostname        |
| `LDAP_PORT`                  | No       | 389           | LDAP port                   |
| `LDAP_USE_TLS`               | No       | false         | Enable LDAP TLS             |
| `LDAP_SKIP_TLS_VERIFY`       | No       | false         | Skip LDAP TLS verification  |
| `LDAP_BIND_DN`               | No       |               | LDAP service account DN     |
| `LDAP_BIND_PASSWORD`         | No       |               | LDAP service account pass   |
| `LDAP_BASE_DN`               | No       |               | LDAP user search base       |
| `LDAP_USER_FILTER`           | No       | (uid=%s)      | LDAP user search filter     |
| `LDAP_ADMIN_GROUP`           | No       |               | LDAP admin group DN         |
| `LDAP_VIEWER_GROUP`          | No       |               | LDAP viewer group DN        |
| `LDAP_GROUP_FILTER`          | No       | (member=%s)   | LDAP group membership filter|

---

## 11. Testing

The project includes:

- **Unit tests** for API v1 endpoints: `transactions`, `executions`, `items`,
  `machines`, `versions`, `reports`, `analytics` (all with `_test.go` files)
- **Unit tests** for web controllers: `root_controller`, `assets_controller`,
  `packages_by_week_controller`
- **Unit tests** for models: `asset_manager_test.go`
- **Unit tests** for packages: `scheduler/main_test.go`,
  `statistics/main_test.go`, `util/template_test.go`, `auth/ldap_test.go`
- **Integration tests**: `tests/integration_test.go` and
  `tests/asset_scenarios_test.go`

---

## 12. Development Tooling

| Tool   | Purpose                                    |
|--------|--------------------------------------------|
| Air    | Live reload for development (`make run`)   |
| Swaggo | Auto-generate Swagger documentation        |
| Make   | Build commands (`clean`, `fmt`, `vet`, `build`, `run`, `doc`) |

The version is stored in `.version` (currently `1.21.0`) and injected at
build time via `-ldflags -X 'github.com/txlog/server/version.SemVer=...'`.

---

## 13. Notable Design Decisions

1. **No ORM**: All database access uses raw SQL queries, giving full control
   over query optimization and making complex analytical queries straightforward.

2. **Embedded assets**: Templates, images, and migration files are embedded into
   the binary via Go's `//go:embed` directive, enabling the `scratch` Docker
   image with zero filesystem dependencies.

3. **Backward-compatible migrations**: The `AssetManager.UpsertAsset()` method
   uses PostgreSQL `SAVEPOINT`s to gracefully degrade when new columns (like
   `agent_version`) don't exist yet, allowing rolling deployments.

4. **Materialized views with graceful fallback**: The packages controller first
   tries the materialized view, falling back to the direct (slower) query if the
   view doesn't exist yet.

5. **Distributed locking**: All scheduled jobs use a PostgreSQL-based lock table
   instead of in-process mutexes, enabling safe multi-replica deployment.

6. **Authentication is optional**: The entire auth system is opt-in. Without
   OIDC/LDAP configuration, the server runs fully open — suitable for internal
   networks or development.

7. **Asset lifecycle tracking**: The assets table tracks machine identity
   changes (e.g., OS reinstall on same hostname) by deactivating old entries and
   activating new ones, maintaining full history.

8. **Brazilian locale conventions**: Number formatting uses dots for thousands
   and commas for decimals (e.g., `1.234,56`), and dates use DD/MM/YYYY format.
