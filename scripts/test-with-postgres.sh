#!/usr/bin/env bash
# Runs the test suite against a throwaway PostgreSQL container managed with
# Podman. The container is removed on exit, including when the tests fail.
#
# The suite applies the migrations itself (see internal/testdb), so an empty
# database is all this has to provide. TXLOG_TEST_REQUIRE_DB turns the tests'
# own "skip when PostgreSQL is unreachable" behaviour into a failure: here the
# database is guaranteed, so a skip would mean something went wrong.
set -euo pipefail

CONTAINER=${TXLOG_TEST_CONTAINER:-txlog-test-pg}
IMAGE=${TXLOG_TEST_IMAGE:-docker.io/library/postgres:17-alpine}

if ! command -v podman >/dev/null 2>&1; then
  echo "podman is required by this target; run 'go test ./...' to use a database of your own" >&2
  exit 1
fi

podman rm -f "$CONTAINER" >/dev/null 2>&1 || true

podman run -d --rm --name "$CONTAINER" \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=txlog_test \
  -p 5432:5432 \
  "$IMAGE" >/dev/null

cleanup() {
  podman rm -f "$CONTAINER" >/dev/null 2>&1 || true
}
trap cleanup EXIT

ready=
for _ in $(seq 1 30); do
  if podman exec "$CONTAINER" pg_isready -U postgres -d txlog_test >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 1
done

if [ -z "$ready" ]; then
  echo "PostgreSQL did not become ready in time" >&2
  podman logs "$CONTAINER" >&2 || true
  exit 1
fi

TXLOG_TEST_REQUIRE_DB=1 go test "${@:-./...}"
