// Package testdb prepares the database that the test suite expects.
//
// Every package with database-backed tests opens its own connection to
// txlog_test and queries tables directly, but nothing applied the migrations,
// so those tables did not exist and the tests failed as soon as PostgreSQL was
// actually reachable. Each such package now calls EnsureSchema from TestMain.
package testdb

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/txlog/server/database"
)

// ConnString is the connection the test suite expects. See tests/README.md for
// how to provide it.
const ConnString = "host=localhost port=5432 user=postgres password=postgres dbname=txlog_test sslmode=disable"

// RequireDBEnv names the variable that turns an unreachable database from a
// reason to skip into a failure. Set it wherever a database is guaranteed —
// make test and CI both do — so that a database that failed to start is
// reported instead of quietly reducing the suite to nothing.
const RequireDBEnv = "TXLOG_TEST_REQUIRE_DB"

// EnsureSchema applies every migration to the test database.
//
// It reports no error when PostgreSQL is unreachable, unless RequireDBEnv is
// set: the tests themselves skip in that case, and failing here would turn a
// missing local database into a suite-wide failure. It always reports an error
// when the database is reachable but the migrations do not apply.
func EnsureSchema() error {
	required := os.Getenv(RequireDBEnv) != ""

	db, err := sql.Open("postgres", ConnString)
	if err != nil {
		if required {
			return fmt.Errorf("%s is set but the connection could not be opened: %w", RequireDBEnv, err)
		}
		return nil
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		if required {
			return fmt.Errorf("%s is set but PostgreSQL is unreachable: %w", RequireDBEnv, err)
		}
		return nil
	}

	return database.Migrate(db)
}
