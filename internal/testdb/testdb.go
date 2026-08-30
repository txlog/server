// Package testdb prepares the database that the test suite expects.
//
// Every package with database-backed tests opens its own connection to
// txlog_test and queries tables directly, but nothing applied the migrations,
// so those tables did not exist and the tests failed as soon as PostgreSQL was
// actually reachable. Each such package now calls EnsureSchema from TestMain.
package testdb

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/txlog/server/database"
)

// ConnString is the connection the test suite expects. See tests/README.md for
// how to provide it.
const ConnString = "host=localhost port=5432 user=postgres password=postgres dbname=txlog_test sslmode=disable"

// EnsureSchema applies every migration to the test database.
//
// It reports no error when PostgreSQL is unreachable: the tests themselves skip
// in that case, and failing here would turn a missing local database into a
// suite-wide failure. It does report an error when the database is reachable
// but the migrations do not apply, since that is a real problem.
func EnsureSchema() error {
	db, err := sql.Open("postgres", ConnString)
	if err != nil {
		return nil
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return nil
	}

	return database.Migrate(db)
}
