# Txlog Server Tests

## Overview

This directory contains comprehensive tests for the Assets feature implemented in the Txlog Server.

## Test Files

### 1. `models/asset_manager_test.go`
Tests for the `AssetManager` model including:
- ✅ Creating new assets
- ✅ Updating existing assets (`last_seen`)
- ✅ Replacing assets (same hostname, different machine_id)
- ✅ Reactivating inactive assets
- ✅ Deactivating old assets
- ✅ Getting active assets by hostname
- ✅ Getting assets by machine_id
- ✅ Ensuring only one active asset per hostname

### 2. `controllers/root_controller_test.go`
Tests for controller functions including:
- ✅ `getTotalActiveAssets()` - Real-time count of active assets
- ✅ `getAssetsByOS()` - Asset distribution by operating system
- ✅ `getAssetsByAgentVersion()` - Asset distribution by agent version
- ✅ `getDuplicatedAssets()` - Assets with multiple machine_ids
- ✅ `getMostUpdatedPackages()` - Package update statistics
- ✅ `getStatistics()` - General statistics
- ✅ Asset listing queries with `last_seen` column
- ✅ Unique active asset constraint validation

### 3. `tests/integration_test.go`
Integration tests for the complete asset lifecycle:
- ✅ Full asset lifecycle (create → update → replace)
- ✅ Asset listing with `last_seen` filtering
- ✅ Database migration verification
- ✅ Table structure and indexes validation
- ✅ Unique constraints validation

## Running Tests

### Prerequisites

1. **PostgreSQL Database**: Tests require a PostgreSQL instance
2. **Test Database**: Create a test database named `txlog_test`

```bash
# Create test database
createdb -U postgres txlog_test

# Run migrations on test database
# (migrations will be run automatically when server starts)
```

### Run All Tests

```bash
# From repository root
go test ./... -v
```

### Run Specific Test Suites

```bash
# Test AssetManager model only
go test ./models -v -run TestAssetManager

# Test Controllers only
go test ./controllers -v -run TestGet

# Test Integration tests only
go test ./tests -v
```

### Run Tests Without Database

Tests will automatically skip if PostgreSQL is not available:

```bash
go test ./... -v
# Output: "Skipping test: PostgreSQL not available"
```

### Test with Coverage

```bash
# Generate coverage report
go test ./models ./controllers ./tests -coverprofile=coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

## Test Database Configuration

Tests expect PostgreSQL to be accessible with these credentials:
- **Host**: localhost
- **Port**: 5432
- **User**: postgres
- **Password**: postgres
- **Database**: txlog_test

To use different credentials, modify the `connStr` in each test file's `setupTestDB()` function.

## Test Data Cleanup

All tests clean up their data automatically:
- Model tests: Delete assets with hostname `test-%`
- Integration tests: Delete assets with machine_id `integration-test-%`

## What Is Being Tested

### Core Functionality
1. **Asset Creation**: New assets are created with `is_active = TRUE`
2. **Asset Updates**: `last_seen` is updated on each execution/transaction
3. **Asset Replacement**: Old assets become inactive when hostname gets new machine_id
4. **Unique Constraint**: Only one active asset per hostname at any time
5. **Reactivation**: Inactive assets can be reactivated
6. **Historical Data**: Inactive assets are preserved with `deactivated_at` timestamp

### Database Schema
1. **Table Existence**: `assets` table exists
2. **Column Validation**: All required columns present
3. **Index Validation**: Performance indexes exist
4. **Constraint Validation**: Unique constraint on (hostname, machine_id)

### Query Performance
1. **Real-time Count**: Active assets count without statistics cache
2. **Last Seen Filtering**: Filtering by `last_seen` column (not executions)
3. **Active Assets Only**: Queries exclude inactive assets
4. **Lateral Join**: Efficient retrieval of latest execution data

## Continuous Integration

These tests can be integrated into CI/CD pipelines:

```yaml
# .github/workflows/test.yml example
- name: Run tests
  env:
    POSTGRES_HOST: localhost
    POSTGRES_PORT: 5432
    POSTGRES_USER: postgres
    POSTGRES_PASSWORD: postgres
    POSTGRES_DB: txlog_test
  run: go test ./... -v
```

## Troubleshooting

### Tests Skip: "PostgreSQL not available"
- Ensure PostgreSQL is running: `systemctl status postgresql`
- Check connection: `psql -U postgres -d txlog_test`
- Verify credentials in test files

### Tests Fail: "permission denied"
- Grant permissions: `GRANT ALL ON DATABASE txlog_test TO postgres;`
- Check pg_hba.conf for authentication method

### Tests Fail: "table does not exist"
- Run migrations first on test database
- Start server once with test database connection string

## Best Practices

1. **Always cleanup**: Tests clean their data in `defer` statements
2. **Use prefixes**: Test data uses `test-` or `integration-test-` prefixes
3. **Skip gracefully**: Tests skip if database unavailable (not fail)
4. **Parallel safe**: Tests use unique identifiers to avoid conflicts
5. **Comprehensive**: Tests cover happy path + edge cases

## Coverage Goals

Target coverage for new features:
- ✅ Models: 80%+ coverage
- ✅ Controllers: 60%+ coverage (UI logic harder to unit test)
- ✅ Integration: End-to-end scenarios

Current coverage:
```bash
# Check coverage
go test ./models ./controllers ./tests -cover
```

## Future Improvements

- [ ] Add benchmarks for query performance
- [ ] Add concurrent access tests
- [ ] Add stress tests (many assets)
- [ ] Mock database for faster unit tests
- [ ] Add API endpoint tests with httptest
