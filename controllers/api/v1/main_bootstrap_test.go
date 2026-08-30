package v1

import (
	"fmt"
	"os"
	"testing"

	"github.com/txlog/server/internal/testdb"
)

// TestMain applies the migrations once before the package's tests run, so that
// the tables they query exist. Tests still skip themselves when PostgreSQL is
// unreachable.
func TestMain(m *testing.M) {
	if err := testdb.EnsureSchema(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to prepare the test schema: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
