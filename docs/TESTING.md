# Testing Guide for Txlog Server

This document describes the test suite for the Txlog Server asset management system.

## Overview

The test suite covers the complete asset management lifecycle introduced in the v2 branch,
including asset creation, updates, replacements, and deactivation logic.

## Test Categories

### 1. Unit Tests (`models/asset_manager_test.go`)

Tests the core `AssetManager` business logic:

- **Asset Creation**: Creating new assets with initial state
- **Asset Updates**: Updating existing assets' `last_seen` timestamp
- **Asset Replacement**: Deactivating old assets when hostname reports with new `machine_id`
- **Asset Reactivation**: Reactivating previously inactive assets
- **Asset Queries**: Retrieving active assets by hostname or `machine_id`

**Coverage**: 9 test cases covering all AssetManager methods

**Run with**:

```bash
go test ./models -v
```

### 2. Controller Tests (`controllers/`)

#### Root Controller Tests (`controllers/root_controller_test.go`)

Tests dashboard and statistics functions:

- `getTotalActiveAssets()` - Real-time active asset count
- `getAssetsByOS()` - Asset distribution by operating system
- `getAssetsByAgentVersion()` - Asset distribution by agent version
- `getDuplicatedAssets()` - Detection of hostname duplicates
- `getMostUpdatedPackages()` - Top 10 most updated packages
- `getStatistics()` - Cached statistics retrieval
- Assets index query validation with various filters

**Coverage**: 7 test cases + query validation tests

**Run with**:

```bash
go test ./controllers -v
```

#### Assets Controller Tests (`controllers/assets_controller_test.go`)

Tests the assets web interface endpoints:

- `GetAssetsIndex()` - Asset listing with filters (search, restart, inactive)
- `DeleteMachineID()` - Asset and related data deletion
- `GetMachineID()` - Asset detail view with execution history

**Coverage**: 11 test cases covering all filter combinations

**Run with**:

```bash
go test ./controllers -v -run TestGetAssetsIndex
```

### 3. API Tests (`controllers/api/v1/`)

#### Executions Controller Tests (`executions_controller_test.go`)

Tests the POST `/v1/executions` endpoint:

- Asset creation on new execution
- Asset update on repeated execution
- Asset replacement when `machine_id` changes
- Error handling (invalid JSON)
- Null timestamp handling

**Coverage**: 6 test cases

**Run with**:

```bash
go test ./controllers/api/v1 -v -run TestPostExecutions
```

#### Machines Controller Tests (`machines_controller_test.go`)

Tests the GET `/v1/machines` endpoint:

- Listing all active assets
- Filtering by OS
- Filtering by agent version
- Combined filters
- Exclusion of inactive assets

**Coverage**: 6 test cases

**Run with**:

```bash
go test ./controllers/api/v1 -v -run TestGetMachines
```

### 4. Integration Tests (`tests/`)

#### Full Lifecycle Test (`integration_test.go`)

End-to-end test of asset lifecycle:

- Create initial asset
- Replace asset with new `machine_id`
- Verify old asset is inactive
- Verify new asset is active

**Coverage**: 1 comprehensive scenario

#### Asset Scenarios Tests (`asset_scenarios_test.go`)

Complex multi-step scenarios:

- **Multiple Replacements**: Sequential machine replacements
- **Concurrent Updates**: Simulating rapid successive updates
- **Asset Reactivation**: Moving assets between hostnames
- **History Preservation**: Verifying complete audit trail
- **Database Constraints**: Validating business rules
- **Edge Cases**: Older timestamps, non-existent assets

**Coverage**: 9 integration scenarios

**Run with**:

```bash
go test ./tests -v
```

## Database Requirements

All tests require a PostgreSQL test database:

```bash
# Connection string used by tests
host=localhost port=5432 user=postgres password=postgres dbname=txlog_test sslmode=disable
```

Tests will **skip gracefully** if the database is not available.

### Setting Up Test Database

```bash
# Create test database
createdb txlog_test

# Run migrations (from main code)
# Tests expect the assets table to exist
```

## Test Data Cleanup

All tests use unique prefixes to avoid conflicts:

- Unit tests: `test-*`
- API tests: `api-test-*`, `machines-test-*`
- Controller tests: `assets-ctrl-test-*`
- Integration tests: `integration-test-*`

Tests clean up their data using `defer` statements.

## Running All Tests

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./models ./controllers ./controllers/api/v1 ./tests -cover

# Run specific test
go test ./models -v -run TestAssetManager_UpsertAsset_NewAsset

# Run with short mode (skip integration tests)
go test ./... -short
```

## Test Patterns

### 1. Database Test Setup

```go
func setupTestDB(t *testing.T) *sql.DB {
    connStr := "host=localhost port=5432 user=postgres password=postgres dbname=txlog_test sslmode=disable"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        t.Skip("Skipping test: PostgreSQL not available")
    }
    if err := db.Ping(); err != nil {
        t.Skip("Skipping test: Cannot connect to PostgreSQL")
    }
    return db
}
```

### 2. Data Cleanup

```go
func cleanupTestData(t *testing.T, db *sql.DB) {
    _, err := db.Exec("DELETE FROM assets WHERE hostname LIKE 'test-%'")
    if err != nil {
        t.Logf("Warning: Failed to cleanup: %v", err)
    }
}

func TestSomething(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    defer cleanupTestData(t, db)
    // ... test code
}
```

### 3. Transaction-based Tests

```go
func TestAssetOperation(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    defer cleanupTestData(t, db)

    am := models.NewAssetManager(db)
    tx, _ := db.Begin()
    defer tx.Rollback()

    err := am.UpsertAsset(tx, "hostname", "machine_id", time.Now())
    if err != nil {
        t.Fatalf("Failed: %v", err)
    }

    if err := tx.Commit(); err != nil {
        t.Fatalf("Failed to commit: %v", err)
    }

    // Verify results
}
```

## Coverage Goals

Target coverage for features in v2 branch:

- ✅ Models: 80%+ coverage (9 tests)
- ✅ Controllers: 60%+ coverage (18 tests for assets)
- ✅ API: 70%+ coverage (12 tests)
- ✅ Integration: End-to-end scenarios (10 tests)

**Total**: 49 test cases covering asset management

## CI/CD Integration

Tests are designed to work in CI/CD pipelines:

- Skip gracefully if database unavailable
- Use unique test data prefixes
- Clean up data automatically
- Exit codes: 0 = pass/skip, 1 = fail

## Troubleshooting

### Tests are skipped

- Ensure PostgreSQL is running: `pg_isready`
- Verify test database exists: `psql -l | grep txlog_test`
- Check connection string in test files

### Tests fail with "relation does not exist"

- Run migrations on test database
- Ensure assets table is created

### Tests fail with "duplicate key"

- Check if test cleanup is working
- Manually clean test data: `DELETE FROM assets WHERE hostname LIKE 'test-%'`

### Template loading errors

- Ensure template files exist in `templates/` directory
- Run tests from repository root directory

## Future Improvements

- [ ] Add benchmarks for query performance
- [ ] Add concurrent access stress tests
- [ ] Mock database for faster unit tests
- [ ] Add API endpoint tests with `httptest.Server`
- [ ] Add property-based testing for edge cases

## Contributing

When adding new features:

1. Write tests first (TDD approach)
2. Use unique test data prefixes
3. Clean up in `defer` statements
4. Skip gracefully if dependencies unavailable
5. Document complex test scenarios
