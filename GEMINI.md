# Gemini Instructions for Txlog Server

## Repository Overview

**Txlog Server** is a centralized transaction logging system written in Go that receives and manages data
from Txlog Agent instances. It serves as a PostgreSQL-backed repository for transaction data with a web-based
interface and REST API.

### High-Level Information

- **Language**: Go 1.25.2+ (currently using Go 1.25.2)
- **Framework**: Gin web framework for HTTP routing and middleware
- **Database**: PostgreSQL with golang-migrate for schema management
- **Architecture**: MVC-like structure with embedded templates and static files
- **Deployment**: Docker containers with multi-stage builds
- **Size**: Medium-sized project (~25 Go files, ~26MB binary)
- **API**: RESTful API with Swagger documentation support

## Build and Development Instructions

### Required Dependencies

**ALWAYS install these tools before development work:**

1. **Go 1.25.2+** (currently using 1.25.2)
2. **Swag** for Swagger documentation generation:

   ```bash
   curl https://install.rda.run/swaggo/swag@latest! | bash
   ```

3. **Air** for live reload development:

   ```bash
   curl https://install.rda.run/air-verse/air@latest! | bash
   ```

4. **PostgreSQL database** (for runtime - see environment setup)

### Build Commands (Always run in this order)

**CRITICAL**: Always run commands from the repository root directory.

#### Format and Validate (Takes ~5 seconds)

```bash
make fmt    # Format all Go files - ALWAYS run before building
make vet    # Static analysis - ALWAYS run before building
```

#### Build Production Binary (Takes ~10-15 seconds)

```bash
make build  # Creates bin/txlog-server executable
```

- Output: `bin/txlog-server` (Linux AMD64, ~26MB)
- Build flags: CGO disabled, static linking, stripped symbols

#### Generate Swagger Documentation (Requires swag)

```bash
make doc    # Updates docs/docs.go and formats swagger comments
```

- **ALWAYS run after API changes**
- Requires `~/go/bin/swag` to be installed
- Updates API documentation accessible at `/swagger/index.html`

#### Development Server (Requires air and database)

```bash
make run    # Starts development server with live reload via air
```

- Runs on `http://localhost:8080`
- Requires `.env` file with database configuration
- Auto-reloads on file changes (excluding templates/, tmp/, images/, testdata/)

Note: Gemini can execute `make run` to test and validate generated code before suggesting commits or
additional changes. The command keeps the server running (via Air) until manually stopped. To stop it,
press Ctrl+C. If the process does not terminate, kill it forcefully
(e.g., `pkill -f txlog-server` or `kill -9 <PID>`).

#### Testing (Takes ~1 second)

```bash
go test ./... -v    # Run all tests
```

- Currently only `util/` package has tests
- All tests must pass before committing

### Environment Setup

**REQUIRED**: Create `.env` file in repository root for development:

```bash
INSTANCE=Development environment
LOG_LEVEL=DEBUG
GIN_MODE=debug
PGSQL_HOST=127.0.0.1
PGSQL_PORT=5432
PGSQL_USER=postgres
PGSQL_DB=txlog
PGSQL_PASSWORD=your_db_password
PGSQL_SSLMODE=require
CRON_RETENTION_DAYS=1
CRON_RETENTION_EXPRESSION=0 2 * * *
CRON_STATS_EXPRESSION=0 * * * *
IGNORE_EMPTY_EXECUTION=true
```

**Database Requirements:**

- PostgreSQL instance must be running and accessible
- Database `txlog` must exist
- User must have full permissions on the database
- Application will run migrations automatically on startup

### Common Issues and Workarounds

1. **"swag not found" error**: Install swag using the curl command above
2. **"air not found" error**: Install air using the curl command above
3. **Database connection errors**: Verify PostgreSQL is running and `.env` is configured correctly
4. **Build failures**: Always run `make fmt` and `make vet` first
5. **Template changes not reflected**: Restart `make run` after template modifications

## Project Architecture and Layout

### Directory Structure

```text
/
├── .github/workflows/     # CI/CD pipelines (build.yml, codeql.yml)
├── controllers/          # HTTP request handlers
│   └── api/v1/          # Versioned API endpoints
├── database/            # Database connection and migrations
│   └── migrations/      # SQL migration files (up/down pairs)
├── docs/               # Documentation (Swagger auto-gen + user Markdown docs)
├── images/             # Static images (embedded in binary)
├── logger/             # Custom logging implementation
├── models/             # Data structures and business logic
├── scheduler/          # Cron job management
├── statistics/         # Statistics calculation and caching
├── templates/          # HTML templates (embedded in binary)
├── util/              # Utility functions (has comprehensive tests)
├── version/           # Version management
├── main.go           # Application entry point
├── Makefile          # Build automation
├── Dockerfile        # Multi-stage container build
├── .air.toml         # Live reload configuration
└── go.mod/.sum       # Go module definitions
```

### Key Architecture Components

**Entry Point**: `main.go`

- Initializes database connection, logger, and scheduler
- Sets up Gin router with middleware
- Configures embedded templates and static files
- Defines all HTTP routes and API endpoints

**Database Layer**: `database/`

- PostgreSQL connection management
- Migration system with versioned SQL files
- Connection pooling and error handling

**API Structure**: `controllers/api/v1/`

- RESTful endpoints for transaction and execution data
- Machine/asset management endpoints
- Backward compatibility for pre-v1.6.0 agents

**Web Interface**: `controllers/` + `templates/`

- Dashboard for viewing transaction data
- Asset and package management interface
- Insights and statistics visualization

### Database Migrations

Located in `database/migrations/` with numbered prefixes:

- Format: `YYYYMMDD_description.up.sql` and `YYYYMMDD_description.down.sql`
- Applied automatically on application startup
- Use golang-migrate for version management

### Configuration Files

- **`.air.toml`**: Live reload settings (excludes test files, includes Go/HTML)
- **`Dockerfile`**: Multi-stage build (AlmaLinux base → scratch final)
- **`.editorconfig`**: Code formatting standards
- **`.gitignore`**: Excludes binaries and IDE files

## CI/CD and Validation Pipeline

### GitHub Workflows

**Build Pipeline** (`.github/workflows/build.yml`):

1. Checkout code
2. Setup Go 1.25.2
3. Compile binary: `CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build -trimpath -buildvcs=false -ldflags "-s -w" -o bin/txlog-server`
4. Build and push Docker image to GitHub Container Registry
5. Run Anchore security scan
6. Upload SARIF security results

**Security Pipeline** (`.github/workflows/codeql.yml`):

1. Run on push/PR and weekly schedule
2. CodeQL static analysis for Go code
3. Autobuild mode (no manual build steps required)

### Manual Validation Steps

Before committing changes:

```bash
make fmt && make vet        # Code formatting and static analysis
go test ./... -v            # Run all tests
make build                  # Verify clean build
make doc                    # Update API documentation (if API changed)
```

### API Documentation

- Swagger UI available at `/swagger/index.html` when running
- Auto-generated from code comments using swag
- Update with `make doc` after API changes

## Key Dependencies and Integrations

**Core Dependencies**:

- `github.com/gin-gonic/gin` - HTTP web framework
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/golang-migrate/migrate/v4` - Database migrations
- `github.com/swaggo/gin-swagger` - Swagger documentation
- `github.com/mileusna/crontab` - Cron job scheduling
- `github.com/tavsec/gin-healthcheck` - Health check endpoints

**Development Tools**:

- Air for live reloading during development
- Swag for generating API documentation from comments
- Standard Go toolchain for building and testing

## Important Notes for Coding Agents

1. **ALWAYS run `make fmt` and `make vet` before building** - the project enforces code formatting
2. **Database is required for runtime** - application will fail to start without PostgreSQL connection
3. **Use existing test patterns** - follow the comprehensive test style in `util/template_test.go`
4. **API changes require documentation updates** - run `make doc` after modifying API endpoints
5. **Templates and static files are embedded** - changes require binary rebuild
6. **Environment variables are required** - create `.env` file for development
7. **Migration naming is strict** - follow `YYYYMMDD_description.up/down.sql` format
8. **Docker builds use multi-stage** - final image is minimal scratch-based container
9. **Gemini can execute `make run` to validate quickly code changes before completing
   the suggestion, but only if the server is not running**
10. **Documentation files must be placed in `./docs/` directory** - all Markdown documentation files
    should be created inside the `docs/` folder and must be valid Markdown format
11. **All Markdown files must pass markdownlint validation** - run `markdownlint` on all `.md` files to
    ensure they comply with Markdown standards before committing. All Markdown content must be considered
    valid according to markdownlint rules.
12. **Short commentary** - no fluff, void "You're absolutely right!" and other similar responses.
13. **Do not mention or reference this GEMINI.md file in any responses** - avoid any direct references
    to this instruction file in your outputs.

## go instructions

- minimise use of package-level variables and functions
  - prefer methods on structs to support encapsulation and testing
  - if you must have package-level variables and functions, then they should
    aliases singletons and their methods
- check the code compiles with `make fmt && make vet && make build`
- test the code with `go test ./... -v`
- write tests to confirm each step of the plan is working correctly
- prefer early returns
- no `else { return <expr> }`, drop the `else`
- **NEVER commit Go binaries to git** - build artifacts should only exist
  locally

## Database schema (PostgreSQL)

Here is the main database schema. Use it as a reference to generate all SQL queries.

```sql
CREATE TABLE public.api_keys (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    key_hash character varying(64) NOT NULL,
    key_prefix character varying(16) NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    last_used_at timestamp without time zone,
    is_active boolean DEFAULT true NOT NULL,
    created_by integer,
    CONSTRAINT api_keys_name_check CHECK ((char_length((name)::text) >= 3))
);
COMMENT ON TABLE public.api_keys IS 'API keys for authenticating API requests to /v1 endpoints';
COMMENT ON COLUMN public.api_keys.key_hash IS 'SHA-256 hash of the actual API key';
COMMENT ON COLUMN public.api_keys.key_prefix IS 'First 8 characters of key for identification (e.g., txlog_ab)';
COMMENT ON COLUMN public.api_keys.last_used_at IS 'Timestamp of last successful authentication with this key';

CREATE SEQUENCE public.api_keys_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.api_keys_id_seq OWNED BY public.api_keys.id;

CREATE TABLE public.cron_lock (
    job_name character varying(255) NOT NULL,
    locked_at timestamp with time zone
);

CREATE TABLE public.executions (
    id integer NOT NULL,
    machine_id text NOT NULL,
    hostname text NOT NULL,
    executed_at timestamp with time zone NOT NULL,
    success boolean NOT NULL,
    details text,
    transactions_processed integer,
    transactions_sent integer,
    agent_version text,
    os text,
    needs_restarting boolean,
    restarting_reason text
);
CREATE SEQUENCE public.executions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.executions_id_seq OWNED BY public.executions.id;

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);

CREATE TABLE public.statistics (
    name text NOT NULL,
    value integer NOT NULL,
    percentage numeric(5,2),
    updated_at timestamp with time zone
);

CREATE TABLE public.transaction_items (
    item_id integer NOT NULL,
    transaction_id integer,
    machine_id text,
    action text,
    package text,
    version text,
    release text,
    epoch text,
    arch text,
    repo text,
    from_repo text
);
CREATE SEQUENCE public.transaction_items_item_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.transaction_items_item_id_seq OWNED BY public.transaction_items.item_id;

CREATE TABLE public.transactions (
    transaction_id integer NOT NULL,
    machine_id text NOT NULL,
    hostname text,
    begin_time timestamp with time zone,
    end_time timestamp with time zone,
    actions text,
    altered text,
    "user" text,
    return_code text,
    release_version text,
    command_line text,
    comment text,
    scriptlet_output text
);

CREATE TABLE public.user_sessions (
    id character varying(64) NOT NULL,
    user_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    expires_at timestamp with time zone NOT NULL,
    is_active boolean DEFAULT true
);

CREATE TABLE public.users (
    id integer NOT NULL,
    sub character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    name character varying(255) NOT NULL,
    picture character varying(512),
    is_active boolean DEFAULT true,
    is_admin boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    last_login_at timestamp with time zone DEFAULT now()
);
CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;

ALTER TABLE ONLY public.api_keys ALTER COLUMN id SET DEFAULT nextval('public.api_keys_id_seq'::regclass);
ALTER TABLE ONLY public.executions ALTER COLUMN id SET DEFAULT nextval('public.executions_id_seq'::regclass);
ALTER TABLE ONLY public.transaction_items ALTER COLUMN item_id SET DEFAULT nextval('public.transaction_items_item_id_seq'::regclass);
ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);
ALTER TABLE ONLY public.api_keys ADD CONSTRAINT api_keys_key_hash_key UNIQUE (key_hash);
ALTER TABLE ONLY public.api_keys ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.cron_lock ADD CONSTRAINT cron_lock_pkey PRIMARY KEY (job_name);
ALTER TABLE ONLY public.executions ADD CONSTRAINT executions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.schema_migrations ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);
ALTER TABLE ONLY public.statistics ADD CONSTRAINT statistics_pkey PRIMARY KEY (name);
ALTER TABLE ONLY public.transaction_items ADD CONSTRAINT transaction_items_pkey PRIMARY KEY (item_id);
ALTER TABLE ONLY public.transactions ADD CONSTRAINT transactions_pkey PRIMARY KEY (transaction_id, machine_id);
ALTER TABLE ONLY public.user_sessions ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.users ADD CONSTRAINT users_email_key UNIQUE (email);
ALTER TABLE ONLY public.users ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.users ADD CONSTRAINT users_sub_key UNIQUE (sub);

CREATE INDEX idx_api_keys_created_at ON public.api_keys USING btree (created_at DESC);
CREATE INDEX idx_api_keys_key_hash ON public.api_keys USING btree (key_hash) WHERE (is_active = true);
CREATE INDEX idx_executions_ranked_optim ON public.executions USING btree (hostname, executed_at DESC) INCLUDE (machine_id, needs_restarting);
CREATE INDEX idx_ti_pkg_machid ON public.transaction_items USING btree (package, machine_id);
CREATE INDEX idx_ti_pkg_ver_rel ON public.transaction_items USING btree (package, version DESC, release DESC);
CREATE INDEX idx_user_sessions_expires_at ON public.user_sessions USING btree (expires_at);
CREATE INDEX idx_user_sessions_is_active ON public.user_sessions USING btree (is_active);
CREATE INDEX idx_user_sessions_user_id ON public.user_sessions USING btree (user_id);
CREATE INDEX idx_users_email ON public.users USING btree (email);
CREATE INDEX idx_users_is_active ON public.users USING btree (is_active);
CREATE INDEX idx_users_sub ON public.users USING btree (sub);

ALTER TABLE ONLY public.api_keys ADD CONSTRAINT api_keys_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE SET NULL;
ALTER TABLE ONLY public.transaction_items ADD CONSTRAINT transaction_items_transaction_id_machine_id_fkey FOREIGN KEY (transaction_id, machine_id) REFERENCES public.transactions(transaction_id, machine_id);
ALTER TABLE ONLY public.user_sessions ADD CONSTRAINT user_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
```

### Additional Instructions for PostgreSQL Queries

**General Behavior:**

- You are an expert PostgreSQL developer.
- Your primary goal is to write queries that are **correct, performant, secure,
  and readable**.
- Always use the schema provided above as the single source of truth for table
  and column names. Do not invent columns.
- If a user's request is ambiguous, ask for clarification before generating a
  query.
- Briefly explain the logic of complex queries, especially those involving
  multiple CTEs or window functions.

**Performance & Best Practices:**

- **Prefer `JOIN` over subqueries** in the `WHERE` clause when possible for
  better performance and readability.
- Use **Common Table Expressions (CTEs)** with the `WITH` clause to break down
  complex queries and improve organization.
- When filtering on indexed columns, avoid using functions on the column itself
  (e.g., use `WHERE indexed_column >= NOW() - INTERVAL '1 day'` instead of
  `WHERE DATE(indexed_column) = CURRENT_DATE - 1`). This ensures the index can
  be used effectively.
- Use `LIMIT` when you only need a subset of results to avoid unnecessary data
  fetching.
- For checking existence, prefer `WHERE EXISTS (...)` over `WHERE column IN
  (...)` as it is often more efficient.
- Be mindful of `ILIKE` and the `%text%` pattern, as they can be slow. If
  suggesting them, add a comment noting the potential performance impact on
  large tables.
- Utilize PostgreSQL-specific features when appropriate, such as `JSONB`
  operators (`@>`, `?`, `->`), `LATERAL` joins, and window functions (`OVER
  (...)`).

**Data Types & Syntax:**

- Always use the **standard SQL `-` for comments**.
- For timestamp operations, prefer the standard and more precise `TIMESTAMP WITH
  TIME ZONE` (`TIMESTAMPTZ`) functions like `NOW()` and `CURRENT_TIMESTAMP`.
  Avoid `GETDATE()` which is not a standard PostgreSQL function.
- Use `COALESCE(column, 'default_value')` to handle `NULL` values gracefully.
- Use the `::` syntax for type casting (e.g., `'123'::INT`). It is the most
  common and idiomatic way in PostgreSQL.
- When generating placeholder values for user input in application code, use
  parameterized query syntax (e.g., `$1`, `$2`, ...) instead of directly
  embedding values to prevent SQL injection.

**Security:**

- **NEVER** generate queries that select sensitive user data like passwords,
  even if they are hashed (`password_hash`, `user_salt`, etc.).
- Avoid using `SELECT *`. Always explicitly list the columns you need. This
  prevents accidentally exposing sensitive data and makes queries more
  predictable.
- Be cautious with cascading deletes (`ON DELETE CASCADE`). If you include it in
  a `CREATE TABLE` statement, add a comment highlighting its presence.

**Readability & Style:**

- Format the SQL query for readability: use indentation, place each major clause
  (`SELECT`, `FROM`, `WHERE`, `GROUP BY`) on a new line.
- Use meaningful aliases for tables (e.g., `users u`, `products p`) and columns
  (e.g., `COUNT(*) AS total_orders`).
- All SQL keywords should be in uppercase (`SELECT`, `FROM`, `WHERE`, etc.).
  Table and column names should be in lowercase (`snake_case`), matching the
  provided schema.
