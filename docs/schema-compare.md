# Schema Comparison: Embedded Validation System

## 1. Problem Statement

The Txlog Server uses [golang-migrate](https://github.com/golang-migrate/migrate)
for database schema management. This tool tracks only the **version number** of
the most recently applied migration in a `schema_migrations` table. It does not
validate whether the actual database schema matches the expected state.

This leads to several failure modes:

- A migration is marked as applied but **failed mid-execution** (dirty state
  with partial changes)
- A migration was applied to a **different database** than the one currently
  connected
- Manual `ALTER TABLE` statements were executed directly, **diverging** from the
  migration history
- The `schema_migrations` table shows version N, but a migration between 1 and
  N was **skipped** or only partially applied

Currently, the application discovers these issues at runtime through PostgreSQL
errors like `column X does not exist (42703)`, which result in 500 errors for
end users.

## 2. Proposed Solution

Embed a **schema snapshot** (JSON) into the compiled binary at build time. At
runtime, compare this expected schema against the actual production database
using `information_schema` and `pg_catalog` queries.

### 2.1 Architecture Overview

```
┌── CI/CD (GitHub Actions) ─────────────────────────────────┐
│                                                           │
│  1. Start ephemeral PostgreSQL (Docker service)           │
│  2. Run all migrations against it                         │
│  3. Run schema-snapshot tool → expected_schema.json       │
│  4. go build (binary embeds the JSON via //go:embed)      │
│  5. Build Docker image with the binary                    │
│                                                           │
└───────────────────────────────────────────────────────────┘
                           │
                 expected_schema.json
                 (embedded in binary)
                           │
                           ▼
┌── Production Runtime ─────────────────────────────────────┐
│                                                           │
│  On startup or on-demand (admin page):                    │
│    1. Query production DB's information_schema             │
│    2. Compare with embedded expected schema               │
│    3. Report differences (missing, extra, type mismatch)  │
│    4. Display results in Admin → Database Migrations      │
│                                                           │
└───────────────────────────────────────────────────────────┘
```

### 2.2 Why Not pgdiff/migra?

Both [pgdiff](https://github.com/joncrlsn/pgdiff) and
[migra](https://github.com/djrobstep/migra) require **two live database
connections simultaneously**. They cannot work with a schema snapshot file.
Since CI/CD does not have access to the production database, these tools cannot
be used in the build pipeline.

Our approach achieves the same result by splitting the comparison into two
phases: **capture** (CI) and **validate** (runtime).

## 3. Data Structures

### 3.1 Schema Snapshot Format

File: `database/expected_schema.json`

```json
{
  "version": "202602120002",
  "generated_at": "2026-02-20T15:30:00Z",
  "tables": {
    "assets": {
      "columns": [
        {
          "name": "id",
          "type": "integer",
          "nullable": false
        },
        {
          "name": "machine_id",
          "type": "text",
          "nullable": false
        },
        {
          "name": "hostname",
          "type": "text",
          "nullable": true
        },
        {
          "name": "agent_version",
          "type": "text",
          "nullable": true
        },
        {
          "name": "os",
          "type": "text",
          "nullable": true
        },
        {
          "name": "is_active",
          "type": "boolean",
          "nullable": true
        },
        {
          "name": "first_seen",
          "type": "timestamp with time zone",
          "nullable": true
        },
        {
          "name": "last_seen",
          "type": "timestamp with time zone",
          "nullable": true
        }
      ]
    },
    "executions": {
      "columns": [
        {
          "name": "id",
          "type": "integer",
          "nullable": false
        },
        {
          "name": "machine_id",
          "type": "text",
          "nullable": false
        },
        {
          "name": "agent_version",
          "type": "text",
          "nullable": true
        }
      ]
    }
  },
  "indices": [
    {
      "name": "idx_assets_agent_version",
      "table": "assets",
      "definition": "CREATE INDEX idx_assets_agent_version ON public.assets USING btree (agent_version) WHERE (agent_version IS NOT NULL)"
    }
  ],
  "materialized_views": [
    {
      "name": "mv_dashboard_agent_stats",
      "columns": ["agent_version", "num_machines"]
    },
    {
      "name": "mv_dashboard_os_stats",
      "columns": ["os", "num_machines"]
    }
  ]
}
```

### 3.2 Go Types

File: `database/schema.go`

```go
package database

import "time"

// SchemaSnapshot represents the expected database schema state.
// It is generated at build time by the schema-snapshot tool and
// embedded into the binary via //go:embed.
type SchemaSnapshot struct {
    Version      string                   `json:"version"`
    GeneratedAt  time.Time                `json:"generated_at"`
    Tables       map[string]TableSchema   `json:"tables"`
    Indices      []IndexSchema            `json:"indices"`
    MatViews     []MatViewSchema          `json:"materialized_views"`
}

// TableSchema describes the expected columns of a table.
type TableSchema struct {
    Columns []ColumnSchema `json:"columns"`
}

// ColumnSchema describes a single column.
type ColumnSchema struct {
    Name     string `json:"name"`
    Type     string `json:"type"`
    Nullable bool   `json:"nullable"`
}

// IndexSchema describes an expected index.
type IndexSchema struct {
    Name       string `json:"name"`
    Table      string `json:"table"`
    Definition string `json:"definition"`
}

// MatViewSchema describes an expected materialized view.
type MatViewSchema struct {
    Name    string   `json:"name"`
    Columns []string `json:"columns"`
}

// SchemaDiff represents a single difference between expected and actual schema.
type SchemaDiff struct {
    Object   string `json:"object"`   // e.g. "assets.agent_version", "idx_assets_agent_version"
    Type     string `json:"type"`     // "column", "index", "materialized_view", "table"
    Status   string `json:"status"`   // "missing", "extra", "type_mismatch", "nullable_mismatch"
    Expected string `json:"expected"` // expected value (empty for "extra")
    Actual   string `json:"actual"`   // actual value (empty for "missing")
}
```

## 4. Component: Schema Snapshot Tool

### 4.1 Purpose

A standalone Go program that connects to a database, queries the schema
metadata, and writes a JSON snapshot file. This runs in CI after all
migrations are applied to an ephemeral database.

### 4.2 Location

`tools/schema-snapshot/main.go`

### 4.3 Implementation

```go
package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"

    _ "github.com/lib/pq"
)

func main() {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "postgres://postgres:check@localhost:5433/txlog_expected?sslmode=disable"
    }

    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal("Failed to connect:", err)
    }
    defer db.Close()

    snapshot := SchemaSnapshot{
        GeneratedAt: time.Now().UTC(),
        Tables:      make(map[string]TableSchema),
    }

    // 1. Get migration version
    var version int64
    err = db.QueryRow("SELECT version FROM schema_migrations LIMIT 1").Scan(&version)
    if err == nil {
        snapshot.Version = fmt.Sprintf("%d", version)
    }

    // 2. Get all columns for all tables in public schema
    rows, err := db.Query(`
        SELECT table_name, column_name, data_type,
               CASE WHEN is_nullable = 'YES' THEN true ELSE false END AS nullable
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name NOT IN ('schema_migrations')
        ORDER BY table_name, ordinal_position
    `)
    if err != nil {
        log.Fatal("Failed to query columns:", err)
    }
    defer rows.Close()

    for rows.Next() {
        var tableName, colName, dataType string
        var nullable bool
        if err := rows.Scan(&tableName, &colName, &dataType, &nullable); err != nil {
            log.Fatal("Failed to scan column:", err)
        }

        table := snapshot.Tables[tableName]
        table.Columns = append(table.Columns, ColumnSchema{
            Name:     colName,
            Type:     dataType,
            Nullable: nullable,
        })
        snapshot.Tables[tableName] = table
    }

    // 3. Get all indices
    idxRows, err := db.Query(`
        SELECT indexname, tablename, indexdef
        FROM pg_indexes
        WHERE schemaname = 'public'
          AND tablename NOT IN ('schema_migrations')
          AND indexname NOT LIKE '%_pkey'
        ORDER BY tablename, indexname
    `)
    if err != nil {
        log.Fatal("Failed to query indices:", err)
    }
    defer idxRows.Close()

    for idxRows.Next() {
        var name, table, def string
        if err := idxRows.Scan(&name, &table, &def); err != nil {
            log.Fatal("Failed to scan index:", err)
        }
        snapshot.Indices = append(snapshot.Indices, IndexSchema{
            Name:       name,
            Table:      table,
            Definition: def,
        })
    }

    // 4. Get materialized views
    mvRows, err := db.Query(`
        SELECT m.matviewname, a.attname
        FROM pg_matviews m
        JOIN pg_class c ON c.relname = m.matviewname
        JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum > 0
        WHERE m.schemaname = 'public'
        ORDER BY m.matviewname, a.attnum
    `)
    if err != nil {
        log.Fatal("Failed to query materialized views:", err)
    }
    defer mvRows.Close()

    mvMap := make(map[string][]string)
    for mvRows.Next() {
        var mvName, colName string
        if err := mvRows.Scan(&mvName, &colName); err != nil {
            log.Fatal("Failed to scan matview:", err)
        }
        mvMap[mvName] = append(mvMap[mvName], colName)
    }

    for name, cols := range mvMap {
        snapshot.MatViews = append(snapshot.MatViews, MatViewSchema{
            Name:    name,
            Columns: cols,
        })
    }

    // 5. Write JSON
    output, err := json.MarshalIndent(snapshot, "", "  ")
    if err != nil {
        log.Fatal("Failed to marshal JSON:", err)
    }

    outFile := "database/expected_schema.json"
    if len(os.Args) > 1 {
        outFile = os.Args[1]
    }

    if err := os.WriteFile(outFile, output, 0644); err != nil {
        log.Fatal("Failed to write file:", err)
    }

    fmt.Printf("Schema snapshot written to %s (%d tables, %d indices, %d materialized views)\n",
        outFile, len(snapshot.Tables), len(snapshot.Indices), len(snapshot.MatViews))
}
```

### 4.4 Usage

```bash
# After migrations are applied to ephemeral DB
DATABASE_URL="postgres://postgres:check@localhost:5433/txlog_expected?sslmode=disable" \
  go run tools/schema-snapshot/main.go database/expected_schema.json
```

## 5. Component: Schema Validator

### 5.1 Purpose

Runs at application runtime. Loads the embedded JSON snapshot, queries the
production database's `information_schema`, and returns a list of differences.

### 5.2 Location

`database/schema_validator.go`

### 5.3 Embedding

```go
package database

import (
    _ "embed"
    "encoding/json"
)

//go:embed expected_schema.json
var expectedSchemaJSON []byte

// LoadExpectedSchema parses the embedded schema snapshot.
// Returns nil if no snapshot is embedded (development builds).
func LoadExpectedSchema() (*SchemaSnapshot, error) {
    if len(expectedSchemaJSON) == 0 {
        return nil, nil
    }

    var schema SchemaSnapshot
    if err := json.Unmarshal(expectedSchemaJSON, &schema); err != nil {
        return nil, err
    }
    return &schema, nil
}
```

### 5.4 Comparison Logic

```go
package database

import (
    "database/sql"
    "fmt"
)

// ValidateSchema compares the embedded expected schema against
// the actual production database and returns all differences.
func ValidateSchema(db *sql.DB) ([]SchemaDiff, error) {
    expected, err := LoadExpectedSchema()
    if err != nil {
        return nil, fmt.Errorf("failed to load expected schema: %w", err)
    }
    if expected == nil {
        return nil, nil // no snapshot embedded (dev build)
    }

    var diffs []SchemaDiff

    // --- Compare columns ---
    actualColumns, err := queryActualColumns(db)
    if err != nil {
        return nil, err
    }

    // Check for missing and mismatched columns
    for tableName, expectedTable := range expected.Tables {
        actualTable, tableExists := actualColumns[tableName]
        if !tableExists {
            diffs = append(diffs, SchemaDiff{
                Object:   tableName,
                Type:     "table",
                Status:   "missing",
                Expected: fmt.Sprintf("%d columns", len(expectedTable.Columns)),
            })
            continue
        }

        actualColMap := make(map[string]ColumnSchema)
        for _, col := range actualTable.Columns {
            actualColMap[col.Name] = col
        }

        for _, expectedCol := range expectedTable.Columns {
            actualCol, exists := actualColMap[expectedCol.Name]
            if !exists {
                diffs = append(diffs, SchemaDiff{
                    Object:   tableName + "." + expectedCol.Name,
                    Type:     "column",
                    Status:   "missing",
                    Expected: expectedCol.Type,
                })
                continue
            }

            if actualCol.Type != expectedCol.Type {
                diffs = append(diffs, SchemaDiff{
                    Object:   tableName + "." + expectedCol.Name,
                    Type:     "column",
                    Status:   "type_mismatch",
                    Expected: expectedCol.Type,
                    Actual:   actualCol.Type,
                })
            }

            if actualCol.Nullable != expectedCol.Nullable {
                diffs = append(diffs, SchemaDiff{
                    Object:   tableName + "." + expectedCol.Name,
                    Type:     "column",
                    Status:   "nullable_mismatch",
                    Expected: fmt.Sprintf("nullable=%v", expectedCol.Nullable),
                    Actual:   fmt.Sprintf("nullable=%v", actualCol.Nullable),
                })
            }
        }

        // Check for extra columns (exist in production but not expected)
        expectedColMap := make(map[string]bool)
        for _, col := range expectedTable.Columns {
            expectedColMap[col.Name] = true
        }
        for _, col := range actualTable.Columns {
            if !expectedColMap[col.Name] {
                diffs = append(diffs, SchemaDiff{
                    Object: tableName + "." + col.Name,
                    Type:   "column",
                    Status: "extra",
                    Actual: col.Type,
                })
            }
        }
    }

    // Check for extra tables (exist in production but not expected)
    for tableName := range actualColumns {
        if _, exists := expected.Tables[tableName]; !exists {
            diffs = append(diffs, SchemaDiff{
                Object: tableName,
                Type:   "table",
                Status: "extra",
            })
        }
    }

    // --- Compare indices ---
    actualIndices, err := queryActualIndices(db)
    if err != nil {
        return nil, err
    }

    actualIdxMap := make(map[string]bool)
    for _, idx := range actualIndices {
        actualIdxMap[idx.Name] = true
    }

    for _, expectedIdx := range expected.Indices {
        if !actualIdxMap[expectedIdx.Name] {
            diffs = append(diffs, SchemaDiff{
                Object:   expectedIdx.Name,
                Type:     "index",
                Status:   "missing",
                Expected: expectedIdx.Definition,
            })
        }
    }

    // --- Compare materialized views ---
    actualMVs, err := queryActualMatViews(db)
    if err != nil {
        return nil, err
    }

    actualMVMap := make(map[string]bool)
    for _, mv := range actualMVs {
        actualMVMap[mv.Name] = true
    }

    for _, expectedMV := range expected.MatViews {
        if !actualMVMap[expectedMV.Name] {
            diffs = append(diffs, SchemaDiff{
                Object:   expectedMV.Name,
                Type:     "materialized_view",
                Status:   "missing",
                Expected: fmt.Sprintf("columns: %v", expectedMV.Columns),
            })
        }
    }

    return diffs, nil
}

// queryActualColumns retrieves all columns from the production database.
func queryActualColumns(db *sql.DB) (map[string]TableSchema, error) {
    rows, err := db.Query(`
        SELECT table_name, column_name, data_type,
               CASE WHEN is_nullable = 'YES' THEN true ELSE false END
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name NOT IN ('schema_migrations')
        ORDER BY table_name, ordinal_position
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    result := make(map[string]TableSchema)
    for rows.Next() {
        var tableName, colName, dataType string
        var nullable bool
        if err := rows.Scan(&tableName, &colName, &dataType, &nullable); err != nil {
            return nil, err
        }
        table := result[tableName]
        table.Columns = append(table.Columns, ColumnSchema{
            Name: colName, Type: dataType, Nullable: nullable,
        })
        result[tableName] = table
    }
    return result, nil
}

// queryActualIndices retrieves all non-primary-key indices.
func queryActualIndices(db *sql.DB) ([]IndexSchema, error) {
    rows, err := db.Query(`
        SELECT indexname, tablename, indexdef
        FROM pg_indexes
        WHERE schemaname = 'public'
          AND tablename NOT IN ('schema_migrations')
          AND indexname NOT LIKE '%_pkey'
        ORDER BY tablename, indexname
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var result []IndexSchema
    for rows.Next() {
        var idx IndexSchema
        if err := rows.Scan(&idx.Name, &idx.Table, &idx.Definition); err != nil {
            return nil, err
        }
        result = append(result, idx)
    }
    return result, nil
}

// queryActualMatViews retrieves all materialized views.
func queryActualMatViews(db *sql.DB) ([]MatViewSchema, error) {
    rows, err := db.Query(`
        SELECT m.matviewname, a.attname
        FROM pg_matviews m
        JOIN pg_class c ON c.relname = m.matviewname
        JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum > 0
        WHERE m.schemaname = 'public'
        ORDER BY m.matviewname, a.attnum
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    mvMap := make(map[string][]string)
    for rows.Next() {
        var name, col string
        if err := rows.Scan(&name, &col); err != nil {
            return nil, err
        }
        mvMap[name] = append(mvMap[name], col)
    }

    var result []MatViewSchema
    for name, cols := range mvMap {
        result = append(result, MatViewSchema{Name: name, Columns: cols})
    }
    return result, nil
}
```

## 6. CI/CD Integration

### 6.1 GitHub Actions Changes

Add the following steps to `.github/workflows/build.yml`, **before** the
Go compile step:

```yaml
    - name: Start ephemeral PostgreSQL
      run: |
        docker run -d --name pg_schema \
          -e POSTGRES_DB=txlog_expected \
          -e POSTGRES_USER=postgres \
          -e POSTGRES_PASSWORD=check \
          -p 5433:5432 \
          postgres:17
        # Wait for PostgreSQL to be ready
        for i in $(seq 1 30); do
          docker exec pg_schema pg_isready -U postgres && break
          sleep 1
        done

    - name: Run migrations on ephemeral DB
      run: |
        go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
        migrate -path database/migrations \
          -database "postgres://postgres:check@localhost:5433/txlog_expected?sslmode=disable" up

    - name: Generate schema snapshot
      run: |
        DATABASE_URL="postgres://postgres:check@localhost:5433/txlog_expected?sslmode=disable" \
          go run tools/schema-snapshot/main.go database/expected_schema.json

    - name: Cleanup ephemeral PostgreSQL
      run: docker rm -f pg_schema
      if: always()
```

Alternatively, use GitHub Actions' built-in `services` for PostgreSQL:

```yaml
    services:
      postgres:
        image: postgres:17
        env:
          POSTGRES_DB: txlog_expected
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: check
        ports:
          - 5433:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
```

### 6.2 Build Order

The build order is critical. The schema snapshot must be generated **before**
`go build` so that `//go:embed` includes the JSON in the binary:

```
1. Build Tailwind CSS         (make css)
2. Start ephemeral PostgreSQL
3. Run all migrations
4. Generate expected_schema.json
5. Stop ephemeral PostgreSQL
6. go build                   (embeds expected_schema.json)
7. docker build + push
```

### 6.3 Development Builds

For local development, `expected_schema.json` won't exist unless the developer
runs the snapshot tool manually. The `LoadExpectedSchema()` function returns
`nil` in this case, and the validator is simply skipped. No impact on dev
workflow.

To generate it locally:

```bash
make schema-snapshot
```

### 6.4 Makefile Target

```makefile
schema-snapshot:
	@echo "Starting ephemeral PostgreSQL..."
	@docker run -d --name pg_schema_check \
	    -e POSTGRES_DB=txlog_expected \
	    -e POSTGRES_USER=postgres \
	    -e POSTGRES_PASSWORD=check \
	    -p 5433:5432 postgres:17
	@echo "Waiting for PostgreSQL..."
	@sleep 3
	@echo "Running migrations..."
	@migrate -path database/migrations \
	    -database "postgres://postgres:check@localhost:5433/txlog_expected?sslmode=disable" up
	@echo "Generating schema snapshot..."
	@DATABASE_URL="postgres://postgres:check@localhost:5433/txlog_expected?sslmode=disable" \
	    go run tools/schema-snapshot/main.go database/expected_schema.json
	@echo "Cleaning up..."
	@docker rm -f pg_schema_check
	@echo "Done. Schema snapshot saved to database/expected_schema.json"
```

## 7. Admin Page Integration

### 7.1 Controller Changes

In `controllers/admin_controller.go`, add schema validation to the admin
handler:

```go
// In the admin handler, after getting migration status:
schemaDiffs, err := database.ValidateSchema(db)
if err != nil {
    logger.Error("Schema validation failed: " + err.Error())
}

// Pass to template
data["schemaDiffs"] = schemaDiffs
data["schemaValidated"] = (err == nil && schemaDiffs != nil)
```

### 7.2 Template Changes

Add a new section in `templates/admin.html` after the Database Migrations
section:

```html
<!-- Schema Validation -->
{{ if .schemaValidated }}
<div class="bg-white rounded-3xl shadow-soft overflow-hidden mt-6">
  <div class="border-b border-txlog-lavender px-6 py-4 flex items-center justify-between">
    <h3 class="font-display font-semibold text-lg">Schema Validation</h3>
    <div class="flex gap-2">
      {{ if eq (len .schemaDiffs) 0 }}
      <span class="bg-txlog-leaf/10 text-txlog-leaf text-xs font-bold px-3 py-1 rounded-lg">
        Schema OK
      </span>
      {{ else }}
      <span class="bg-txlog-coral/10 text-txlog-coral text-xs font-bold px-3 py-1 rounded-lg">
        {{ len .schemaDiffs }} differences
      </span>
      {{ end }}
    </div>
  </div>
  <div class="p-6">
    {{ if eq (len .schemaDiffs) 0 }}
    <div class="bg-txlog-leaf/5 border border-txlog-leaf/20 rounded-2xl p-4 flex gap-3">
      <i data-lucide="shield-check" class="w-5 h-5 text-txlog-leaf flex-shrink-0 mt-0.5"></i>
      <div>
        <h4 class="font-semibold text-sm mb-1">Schema Verified</h4>
        <p class="text-sm text-txlog-indigo/60">
          The database schema matches the expected state from build
          {{ .schemaVersion }}.
        </p>
      </div>
    </div>
    {{ else }}
    <div class="bg-txlog-coral/5 border border-txlog-coral/20 rounded-2xl p-4 mb-4">
      <div class="flex gap-3">
        <i data-lucide="triangle-alert"
           class="w-5 h-5 text-txlog-coral flex-shrink-0 mt-0.5"></i>
        <div>
          <h4 class="font-semibold text-sm mb-1">Schema Divergence Detected</h4>
          <p class="text-sm text-txlog-indigo/60">
            The production database does not match the expected schema.
            This may indicate failed or missing migrations.
          </p>
        </div>
      </div>
    </div>
    <div class="overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-txlog-lavender/50 text-left">
            <th class="px-4 py-3 font-semibold text-txlog-indigo/70
                       text-xs uppercase tracking-wider">Status</th>
            <th class="px-4 py-3 font-semibold text-txlog-indigo/70
                       text-xs uppercase tracking-wider">Type</th>
            <th class="px-4 py-3 font-semibold text-txlog-indigo/70
                       text-xs uppercase tracking-wider">Object</th>
            <th class="px-4 py-3 font-semibold text-txlog-indigo/70
                       text-xs uppercase tracking-wider">Details</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-txlog-lavender/30">
          {{ range .schemaDiffs }}
          <tr>
            <td class="px-4 py-3">
              {{ if eq .Status "missing" }}
              <span class="bg-txlog-coral/10 text-txlog-coral text-xs
                          font-bold px-2 py-0.5 rounded-md">Missing</span>
              {{ else if eq .Status "extra" }}
              <span class="bg-txlog-golden/10 text-txlog-golden text-xs
                          font-bold px-2 py-0.5 rounded-md">Extra</span>
              {{ else }}
              <span class="bg-txlog-sky/10 text-txlog-sky text-xs
                          font-bold px-2 py-0.5 rounded-md">Mismatch</span>
              {{ end }}
            </td>
            <td class="px-4 py-3 text-txlog-indigo/60">{{ .Type }}</td>
            <td class="px-4 py-3 font-mono text-sm">{{ .Object }}</td>
            <td class="px-4 py-3 text-txlog-indigo/60 text-xs">
              {{ if .Expected }}Expected: {{ .Expected }}{{ end }}
              {{ if .Actual }}Actual: {{ .Actual }}{{ end }}
            </td>
          </tr>
          {{ end }}
        </tbody>
      </table>
    </div>
    {{ end }}
  </div>
</div>
{{ end }}
```

### 7.3 Visual Result

When schema matches:

```
┌─ Schema Validation ──────────────── [Schema OK] ─┐
│                                                   │
│  🛡️  Schema Verified                              │
│     The database schema matches the expected      │
│     state from build 202602120002.                 │
│                                                   │
└───────────────────────────────────────────────────┘
```

When differences are found:

```
┌─ Schema Validation ────────── [3 differences] ───┐
│                                                   │
│  ⚠️  Schema Divergence Detected                   │
│     The production database does not match...     │
│                                                   │
│  STATUS    TYPE     OBJECT                DETAILS │
│  Missing   column   assets.agent_version  text    │
│  Missing   index    idx_assets_agent_ver. CREATE..│
│  Extra     column   assets.legacy_field   text    │
│                                                   │
└───────────────────────────────────────────────────┘
```

## 8. Edge Cases

### 8.1 Development Builds Without Snapshot

When building locally without running the snapshot tool, the
`expected_schema.json` file won't exist. The `//go:embed` directive will cause
a **compile error**.

Solution: Always keep a placeholder file in the repository:

```json
{}
```

The `LoadExpectedSchema()` function detects empty/minimal JSON and returns
`nil`, disabling validation for dev builds. Only CI-generated snapshots with
populated data trigger validation.

Alternatively, use a build tag:

```go
//go:build schema_check

package database

//go:embed expected_schema.json
var expectedSchemaJSON []byte
```

```go
//go:build !schema_check

package database

var expectedSchemaJSON []byte // empty, validation disabled
```

CI builds with `-tags schema_check`, dev builds without.

### 8.2 Schema Changes Between Build and Deploy

If a migration is applied between the time the binary is built and when it's
deployed, the validator will report "extra" columns/indices. This is expected
and not harmful — it means the database is *ahead* of the binary.

### 8.3 Sequences and Constraints

The current design intentionally **excludes** sequences, constraints (foreign
keys, check constraints), and triggers. These are harder to compare
meaningfully and rarely the source of migration failures. They can be added
later if needed.

### 8.4 Column Order

PostgreSQL does not guarantee column order. The comparison should be done by
**column name**, not by position. The provided implementation already does
this (uses maps keyed by column name).

## 9. Files to Create

| File | Purpose |
|---|---|
| `database/schema.go` | Type definitions (SchemaSnapshot, SchemaDiff, etc.) |
| `database/schema_validator.go` | Comparison logic + embed directive |
| `database/expected_schema.json` | Placeholder (empty JSON `{}`) |
| `tools/schema-snapshot/main.go` | CLI tool to generate snapshot |

## 10. Files to Modify

| File | Change |
|---|---|
| `.github/workflows/build.yml` | Add PostgreSQL service + snapshot generation steps |
| `Makefile` | Add `schema-snapshot` target |
| `controllers/admin_controller.go` | Call validator, pass results to template |
| `templates/admin.html` | Add Schema Validation section |
| `.gitignore` | Optionally ignore `expected_schema.json` (generated file) |

## 11. Estimated Effort

| Component | Lines of Code | Complexity |
|---|---|---|
| Type definitions | ~50 | Low |
| Snapshot tool | ~120 | Low |
| Validator | ~180 | Medium |
| CI/CD changes | ~25 | Low |
| Admin template | ~60 | Low |
| Admin controller | ~15 | Low |
| **Total** | **~450** | **Medium** |

## 12. Future Enhancements

- **Constraint validation**: Compare foreign keys, unique constraints, check
  constraints
- **Auto-fix**: Generate and offer to run the SQL needed to fix divergences
- **API endpoint**: `/api/v1/schema/validate` for monitoring/alerting
- **Scheduled checks**: Run validation periodically and log warnings
- **Notification**: Send alert (webhook/email) when schema divergence is
  detected at startup
