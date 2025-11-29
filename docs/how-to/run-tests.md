# How to Run Tests

Txlog Server has a comprehensive test suite covering models, controllers, and integration scenarios.

## Running All Tests

To run the entire test suite:

```bash
go test ./...
```

## Running Specific Tests

To run tests for a specific package (e.g., models):

```bash
go test ./models -v
```

To run a specific test function:

```bash
go test ./models -v -run TestAssetManager_UpsertAsset
```

## Test Database

The tests require a PostgreSQL database. By default, they expect a database named `txlog_test` on localhost.

1. **Create the Test Database**:

    ```bash
    createdb txlog_test
    ```

    *(Or use your preferred method to create a Postgres DB)*.

2. **Connection String**:
    The tests use the following default connection string:
    `host=localhost port=5432 user=postgres password=postgres dbname=txlog_test sslmode=disable`

    If your local setup is different, you may need to modify the `setupTestDB` helper in the test files.

## Integration Tests

Integration tests in the `tests/` directory simulate full user scenarios. They are slower but verify the system end-to-end.

```bash
go test ./tests -v
```

For more details on the testing strategy, see [TESTING.md](../TESTING.md).
