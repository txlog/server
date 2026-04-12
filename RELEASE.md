# Release v1.27.0

## [1.27.0] - 2026-04-11

This release focuses on significant architectural improvements for performance and robustness, including full context propagation for database operations and dependency injection in background tasks.

### Changed

- Refactor the scheduler and vulnerability jobs to use explicit database dependency
  injection instead of global state, improving testability and architectural
  decoupling.
- Update all controller database operations to be context-aware, using
  `QueryContext`, `QueryRowContext`, and `BeginTx` to ensure request-scoped
  cancellation and prevent database connection pool exhaustion.

### Fixed

- Prevent goroutine leaks in API key middleware by implementing bounded timeouts
  using `context.WithTimeout` for asynchronous `last_used_at` updates.
- Improve panic recovery in asset deletion by logging the affected machine ID
  context before re-panicking to Gin's middleware.
- Fix test suite compatibility by updating function signatures in all existing
  test files to match the new context-aware and dependency-injected
  architecture.

---

**Full Changelog**: https://github.com/txlog/server/compare/v1.26.1...v1.27.0
