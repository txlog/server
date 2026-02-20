# Changelog

<!-- markdownlint-disable MD024 -->

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic
Versioning](https://semver.org/spec/v2.0.0.html).

<!--

`Added` for new features.
`Changed` for changes in existing functionality.
`Deprecated` for soon-to-be removed features.
`Removed` for now removed features.
`Fixed` for any bug fixes.
`Security` in case of vulnerabilities.
-->

## [Unreleased]

### Added

- Expand the migration status grid to four columns and display the database
  migration state.
- Introduce a fallback query for the agent version filter to handle cases where
  the column is undefined in the assets table.
- Configure the API key page to automatically reload after closing the API Key
  creation modal.
- Add hover tooltips to action icons in the machine ID view.
- Add a new design document outlining an embedded database schema validation
  system.

### Changed

- Convert the user interface to use Tailwind CSS and integrate the build step
  into the CI workflow.
- Replace legacy CSS classes with Tailwind utilities for elements such as status
  indicators, UI components, and modals.
- Standardize link styling across all HTML templates and compact Go template
  expressions.
- Increase the width of the API Key creation modal to prevent content overflow
  and add a stronger shadow effect to dropdown menus.
- Update the Visual Identity documentation with refined color palettes, UI
  component specifications, and WCAG compliance details.
- Update OIDC and LDAP configuration links in the admin panel and remove
  outdated research documents.

### Fixed

- Adjust asset image styling by removing `overflow-hidden` from container and
  applying rounded corners directly to the image.

## [1.22.0] - 2026-02-16

### Added

- Add pagination (`limit` and `offset` parameters) to `GET /v1/machines`
  endpoint.
- Add pagination (`limit` and `offset` parameters) to `GET /v1/transactions`
  endpoint.

### Changed

- Update Go version to 1.26.0.
- Update Swagger/OpenAPI documentation with improved comment formatting.
- Bump `github.com/quic-go/quic-go` from 0.57.0 to 0.59.0.
- Bump `golang.org/x/crypto` from 0.45.0 to 0.48.0.
- Bump `golang.org/x/net` from 0.47.0 to 0.50.0.
- Bump `github.com/go-playground/validator/v10` from 10.27.0 to 10.30.1.
- Bump `github.com/klauspost/compress` from 1.18.0 to 1.18.4.
- Bump `github.com/redis/go-redis/v9` from 9.17.2 to 9.18.0.
- Bump `github.com/bytedance/sonic` from 1.14.0 to 1.15.0.
- Bump `github.com/gabriel-vasile/mimetype` from 1.4.8 to 1.4.13.
- Bump `go.mongodb.org/mongo-driver` from 1.17.7 to 1.17.9.

### Removed

- Remove `GET /v1/machines/ids` endpoint.

## [1.21.0] - 2026-02-11

### Added

- Implement detection and force-cleaning of dirty database migration states via
  the admin panel.
- Add `agent_version` to assets and create dashboard materialized views.
- Add pagination to `GetTransactions` endpoint.
- Add pagination to `GetExecutions` endpoint.
- Add `golang.org/x/sync` as a direct dependency.

### Changed

- Update Go version to 1.25.6.
- Bump `golang.org/x/oauth2` from 0.34.0 to 0.35.0.
- Bump `github.com/lib/pq` to 1.11.2.
- Bump `github.com/tavsec/gin-healthcheck`.
- **Performance**: Batch insert transaction items in `PostTransactions`.
- **Performance**: Add `LIMIT 50` to transactions query in `GetMachineID`.
- **Performance**: Simplify `getAssetsByOS` to query assets table directly.
- **Performance**: Execute dashboard queries in parallel.
- **Performance**: Add composite indexes for critical query patterns.
- **Performance**: Configure connection pool limits.

### Fixed

- Fix concurrent map write panic in `EnvironmentVariablesMiddleware` by creating
  a per-request copy of the environment map.
- Fix backward compatibility for `agent_version` column in models.

## [1.20.0] - 2026-01-25

### Added

- Add Reports UI pages (`/reports/anomalies`, `/reports/compare`,
  `/reports/progression`) and consolidate all analytics API endpoints under
  `/v1/reports/*` for consistent naming.
- Update Contributor Covenant badge to version 3.0.

### Fixed

- Fix GitHub Actions build workflow not triggering for branches with slashes
  (e.g., `feat/advanced-analysis`) by changing branch pattern from `*` to `**`.
- Fix Docker image tags containing invalid characters by sanitizing branch names
  (replacing `/` with `-`) before using them as tags.
- Fix package comparison (`/analytics/compare`) incorrectly showing packages as
  "Version Differences" when all assets have the same version. The comparison
  now uses original version strings instead of re-constructed values.
- Fix package comparison displaying only version without release, making it
  appear that packages with different releases had the same version. Now
  displays full `version-release` format (e.g., `3.1.5-1.el9`).
- Fix anomaly badges text color for better readability by adding explicit white
  text color.

### Changed

- Allow session-authenticated users to access `/v1` API endpoints directly,
  enabling UI pages to call API endpoints without requiring an API key header.
- Optimize API key middleware to check for API key header before session cookie,
  avoiding unnecessary database queries for API requests that include an API
  key.
- Consolidate duplicate `/web/*` endpoints into `/v1/*` endpoints:
  - `/web/machines` → `/v1/machines`
  - `/web/packages/:name/:version/:release/assets` →
    `/v1/packages/:name/:version/:release/assets`
  - `/web/items` → `/v1/items`

### Removed

- Remove duplicate `/web/*` routes and their corresponding `*Web` controller
  functions (`GetMachinesWeb`, `GetAssetsUsingPackageVersionWeb`).

## [1.19.1] - 2026-01-11

### Changed

- Add `VERSION` and `BUILD_DATE` build arguments to Docker build workflow,
  populating OCI image labels (`org.opencontainers.image.version` and
  `org.opencontainers.image.created`) with accurate build metadata.

## [1.19.0] - 2025-12-30

### Added

- Add `GET /v1/reports/monthly` API endpoint to expose management report data in
  JSON format. Returns `month`, `year`, `asset_count`, and a `packages` array
  containing `os_version`, `package_rpm`, and `assets_affected` for each package
  updated in the specified period.
- Add comprehensive unit tests for the new reports endpoint covering parameter
  validation, response structure, and edge cases.

## [1.18.7] - 2025-12-17

### Changed

- Replace magic numbers with HTTP status constants (`http.StatusOK`,
  `http.StatusBadRequest`, `http.StatusInternalServerError`) across all API
  controllers for improved code readability and maintainability.
- Optimize `GET /v1/items` endpoint by replacing `Query` with `QueryRow` for
  single-row transaction lookup, reducing database round-trips.
- Simplify code structure in `GET /v1/transactions/ids` with early returns
  instead of nested if-else blocks.

### Fixed

- Fix resource leak in `GET /v1/items` where `defer rows.Close()` was
  incorrectly placed after the loop instead of immediately after query success.
- Fix resource leak in `GET /v1/items/ids` with same `defer rows.Close()`
  positioning issue.
- Fix resource leak in `GET /v1/transactions/ids` with same `defer rows.Close()`
  positioning issue.

### Added

- Add comprehensive unit tests for `GET /v1/version` endpoint covering version
  formats, content-type, and closure behavior.
- Add comprehensive unit tests for `GET /v1/items` and `GET /v1/items/ids`
  endpoints covering validation, empty results, and data retrieval scenarios.
- Add comprehensive unit tests for `GET /v1/transactions/ids` endpoint covering
  JSON validation, filtering, and result ordering.
- Add comprehensive unit tests for `POST /v1/transactions` endpoint covering
  transaction creation, duplicate handling, and multi-item transactions.

## [1.18.6] - 2025-12-16

### Changed

- Optimize `/packages` endpoint performance with materialized view for faster
  package listing queries. Previously, each request executed multiple complex
  CTEs scanning the entire `transaction_items` table; now uses pre-computed data
  refreshed every 5 minutes.
- Add scheduled job to refresh `mv_package_listing` materialized view every 5
  minutes for near real-time data with dramatically improved query performance.
- Optimize `/assets` endpoint performance by storing the `os` field directly in
  the `assets` table. Previously, each request executed expensive `LEFT JOIN
  LATERAL` subqueries to fetch the OS from the latest execution for each asset;
  now uses a simple direct query on the `assets` table.
- Update `UpsertAsset` to save the OS with each execution, ensuring the assets
  table always has the latest OS information in real-time.

### Fixed

- Fix "Replaced assets" card on dashboard not showing any results. The previous
  query incorrectly looked for hostnames with multiple active assets, but by
  design only one asset can be active per hostname. Now correctly identifies
  hostnames with assets that were deactivated (replaced) in the last 30 days.

## [1.18.5] - 2025-12-15

### Security

- Fix HTTP/3 QPACK Header Expansion DoS vulnerability by updating
  `github.com/quic-go/quic-go` from 0.54.1 to 0.57.0.

## [1.18.4] - 2025-12-04

### Changed

- Update Go version to 1.25.5.
- Change version definition to `.version` file.
- Overhaul and expand documentation with new tutorials, how-to guides, and
  reference material.
- Translate LDAP documentation from Portuguese to English.
- Update copyright year range in LICENSE.
- Bump `github.com/golang-migrate/migrate/v4`.
- Bump `github.com/coreos/go-oidc/v3` from 3.16.0 to 3.17.0.
- Bump `github.com/tavsec/gin-healthcheck`.

### Removed

- Remove 'New' badge from Admin navigation link.

## [1.18.3] - 2025-11-19

### Added

- Empty states for the Package Progression screen when no data is available.

## [1.18.2] - 2025-11-19

### Fixed

- Fixed a bug where deleting an asset would incorrectly delete transaction items
  from other assets that shared the same transaction_id.

## [1.18.1] - 2025-11-13

### Changed

- Monthly package data retrieval now uses OS version from `executions` table
  instead of `transactions.release_version` for more accurate OS information.
- CSV output for management reports now includes full RPM package names in
  format `package-version-release.arch` (e.g., `nginx-1.20.1-14.el9.x86_64`).
- Enhanced `PostTransactions` endpoint to handle existing transactions and
  return appropriate status messages.

### Dependencies

- Bumped `golang.org/x/oauth2` from 0.32.0 to 0.33.0.

## [1.18.0] - 2025-11-07

### Added

- AI report generator for package updates on the Package Progression page with
  month and year selection.
- Automatic CVE research instruction in generated prompts using Red Hat errata
  as reference.
- Support for multiple AI assistants (ChatGPT, Claude, Gemini).
- Modal dialogs for displaying prompts and error messages.
- Assets table for centralized server identity and lifecycle management.
- Automatic detection and tracking of replaced servers (same hostname, different
  machine ID).
- Real-time asset count throughout the application.
- Comprehensive test coverage for assets, statistics, scheduler, and
  controllers.
- Time-based status indicators for server last activity.
- `needs_restarting` field to track servers requiring reboot.
- Complete database schema documentation with table and column comments.

### Changed

- Package update reports now prioritize servers affected over total transaction
  count.
- AI prompts translated to English for broader compatibility.
- Package filtering uses both version and release fields for precise
  identification.
- Asset queries now use dedicated assets table instead of complex window
  functions.
- All statistics and listings reflect only active servers.
- Package order changed to ascending in weekly progression view.
- API endpoints updated to `/packages/:name/:version/:release/assets` format.

### Fixed

- Consistent button styling across all modals.
- Package counts now exclude inactive servers from statistics.
- Dashboard cards show only active server data.
- NULL handling in asset queries using proper SQL types.
- Improved template readability with better formatting.

### Removed

- Cached server count statistics (now computed in real-time).
- `IGNORE_EMPTY_EXECUTION` environment variable and related logic.

## [1.17.0] - 2025-10-29

### Added

- Prioritize OIDC `sub` over LDAP `sub` for user authentication, ensuring
  consistent user identification across different identity providers.

### Changed

- Bump `github.com/tavsec/gin-healthcheck` from 1.7.9 to 1.8.0.
- Add new database indexes to improve query performance.

### Fixed

- Fix modal for asset details.
- Fix a bug related to the Tabler javascript library.
- Check for existing user by email before creating a new user to prevent
  duplicates.

## [1.16.2] - 2025-10-14

### Fixed

- Package listing page showing incorrect version counts due to 'Change ' prefix
  in package names.

### Security

- Fix CVE-2025-59530 in github.com/quic-go/quic-go via minor version upgrade
  from 0.54.0 to 0.54.1

## [1.16.1] - 2025-10-13

### Changed

- Show a full empty-state component on the package details page when no assets
  run the selected version, improving the user experience.

### Fixed

- Fixed package details page failing to load asset lists when an API key was
  missing by adding a web-only endpoint that uses session-based authentication.
- Fixed assets list endpoint returning `null` instead of `[]` when no assets
  were found, preventing frontend failures.

## [1.16.0] - 2025-10-11

### Added

- Add inactive filter to asset search functionality.

### Changed

- Update Go version to 1.25.2 across documentation and build files.
- Bump `golang.org/x/oauth2` from 0.31.0 to 0.32.0.

### Fixed

- Adjust formatting of note regarding `make run` command in `GEMINI.md`.
- Update Docker image tags handling and clean branch images in build workflow.

## [1.15.2] - 2025-10-08

### Changed

- New `formatDate` and `formatDateTime` template functions for consistent date
  and time formatting.

## [1.15.1] - 2025-10-08

### Fixed

- API key authentication is now only required when OIDC or LDAP authentication
  is enabled. When both are disabled, /v1 endpoints are accessible without API
  key

## [1.15.0] - 2025-10-07

### Added

- LDAP authentication support with comprehensive integration
- User avatar display with initials extraction from DN
- Markdownlint configuration file (`.markdownlint.json`)

### Changed

- Enhanced login interface to support LDAP authentication with improved error
  handling
- Updated authentication controller to handle both OIDC and LDAP login requests
- Modified authentication middleware to allow requests when neither OIDC nor
  LDAP is configured
- Improved admin panel with LDAP-specific user management features
- Updated header template with conditional LDAP/OIDC authentication state checks
- Updated dependencies in `go.mod` and `go.sum` for LDAP functionality
- Bumped `github.com/coreos/go-oidc/v3` from 3.15.0 to 3.16.0

### Fixed

- Conditional rendering for user avatar in admin template
- Markdown linting comments for consistency in README

## [1.14.0] - 2025-10-03

### Added

- API Key authentication system for /v1 endpoints
- Swagger/OpenAPI security definitions for API key authentication
- Admin panel accessible without OIDC authentication
- Server Configuration card in admin panel with comprehensive settings display
- First user auto-promotion to administrator when using OIDC
- Enhanced admin panel styling

### Changed

- Moved Settings functionality from footer offcanvas to admin panel
- Environment variables middleware expanded
- Admin middleware behavior refined
- Admin panel route structure reorganized

### Security

- API keys stored as bcrypt hashes (cost 10) in database
- API key secrets masked in admin interface
- Password and sensitive data consistently masked
- API endpoints protected by API key authentication
- Admin panel endpoints properly protected

### Fixed

- Admin panel card label contrast improved using `fw-bold` class
- OIDC authentication empty state provides link to documentation
- API key creation modal closes and refreshes list automatically
- Delete and revoke operations redirect properly (no 404 errors)

## [1.13.1] - 2025-10-01

### Fixed

- Enhance database migration scripts with error handling and conditional index
  creation

## [1.13.0] - 2025-10-01

### Added

- OIDC support for authentication
- Admin interface for user management and database migrations

## [1.12.1] - 2025-09-23

### Fixed

- Remove simulated delay in package history endpoint

## [1.12.0] - 2025-09-23

### Added

- Package listing and package history

### Changed

- `Details` button on asset and package listing
- bumps github.com/gin-gonic/gin from 1.10.1 to 1.11.0

## [1.11.1] - 2025-09-09

### Fixed

- Fix version badge display logic to handle version format mismatches

## [1.11.0] - 2025-09-01

### Changed

- Switch base image from `alpine` to `scratch` in `Dockerfile`
- Bumps github.com/swaggo/swag from 1.16.5 to 1.16.6
- Bumps github.com/golang-migrate/migrate/v4 from 4.18.3 to 4.19.0

### Added

- Asset data deletion
- `txlog` user on docker image

## [1.10.2] - 2025-07-17

### Changed

- Bump github.com/tavsec/gin-healthcheck from 1.7.8 to 1.7.9 by @dependabot[bot]
  in <https://github.com/txlog/server/pull/62>
- Bump github.com/swaggo/swag from 1.16.4 to 1.16.5 by @dependabot[bot] in
  <https://github.com/txlog/server/pull/63>

## [1.10.1] - 2025-07-14

### Changed

- Increased the time range for the "updated packages by week" graph from 12 to
  15 weeks for improved historical insight.
- Updated frontend dependencies:
  - Bootstrap upgraded from 5.3.6 to 5.3.7
  - Tabler upgraded from 1.3.2 to 1.4.0 (CSS & JS)
  - ApexCharts upgraded from 4.7.0 to 5.2.0 (CSS & JS)

## [1.10.0] - 2025-07-10

### Changed

- Bump github.com/tavsec/gin-healthcheck from 1.7.7 to 1.7.8 by @dependabot in
  <https://github.com/txlog/server/pull/58>
- Show number of updated packages by week by @rdeavila in
  <https://github.com/txlog/server/pull/59>

## [1.9.1] - 2025-06-05

### Changed

- Optimize asset restart query with ranked executions index

## [1.9.0] - 2025-05-29

### Changed

- Bump github.com/gin-gonic/gin from 1.10.0 to 1.10.1 by @dependabot in
  <https://github.com/txlog/server/pull/55>
- Bump github.com/tavsec/gin-healthcheck from 1.7.6 to 1.7.7 by @dependabot in
  <https://github.com/txlog/server/pull/56>
- Implement Restart Tracking for Assets (UI & API) by @rdeavila in
  <https://github.com/txlog/server/pull/57>

## [1.8.3] - 2025-05-20

### Security

- Added HTML escaping in `Text2HTML` function to prevent potential XSS
  vulnerabilities

### Changed

- Bump tabler version to 1.3.0

## [1.8.2] - 2025-05-16

### Changed

- Update brand detection for Red Hat
- Add utility functions for template processing
- Fix settings page

## [1.8.1] - 2025-05-16

### Changed

- Fix asset search
- Fix execution listing from agent

### Docker Image

```bash
docker pull ghcr.io/txlog/server:v1.8.1
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.8.0...v1.8.1>

## [1.8.0] - 2025-05-16

### Changed

- Implements restart detection by @rdeavila in
  <https://github.com/txlog/server/pull/52>

### Docker Image

```bash
docker pull ghcr.io/txlog/server:v1.8.0
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.7.4...v1.8.0>

## [1.7.4] - 2025-05-08

### Changed

- Fix version check URL

### Docker Image

```bash
docker pull ghcr.io/txlog/server:v1.7.4
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.7.3...v1.7.4>

## [1.7.3] - 2025-05-06

### Changed

- Move settings to a offcanvas by @rdeavila in
  <https://github.com/txlog/server/pull/44>
- Show new server version by @rdeavila in
  <https://github.com/txlog/server/pull/45>

### Docker Image

```bash
docker pull ghcr.io/txlog/server:v1.7.3
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.7.2...v1.7.3>

## [1.7.2] - 2025-05-06

### Changed

- List of Replaced assets and Most updated packages, by @rdeavila in
  <https://github.com/txlog/server/pull/41>

### Docker Image

```bash
docker pull ghcr.io/txlog/server:v1.7.2
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.7.1...v1.7.2>

## [1.7.1] - 2025-05-01

### Changed

- Enhance transaction display with user action icons and improved user handling

### Docker Image

```bash
docker pull ghcr.io/txlog/server:v1.7.1
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.7.0...v1.7.1>

## [1.7.0] - 2025-04-30

### Changed

- Add asset listing for OS and Agent distribution by @rdeavila in
  <https://github.com/txlog/server/pull/38>

### Docker Image

```bash
docker pull ghcr.io/txlog/server:v1.7.0
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.6.3...v1.7.0>

## [1.6.3] - 2025-04-24

### Changed

- Bump github.com/golang-migrate/migrate/v4 from 4.18.2 to 4.18.3 by @dependabot
  in <https://github.com/txlog/server/pull/37>

### Docker Image

```bash
docker pull ghcr.io/txlog/server:v1.6.3
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.6.2...v1.6.3>

## [1.6.2] - 2025-04-22

### Changed

- [Aikido] Fix CVE-2025-22872 in golang.org/x/net via minor version upgrade from
  0.36.0 to 0.38.0 by @aikido-autofix in
  <https://github.com/txlog/server/pull/36>

### Docker Image

```bash
docker pull ghcr.io/txlog/server:v1.6.2
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.6.1...v1.6.2>

## [1.6.1] - 2025-04-14

### Changed

### Enhancements

- Added descriptive Docker image labels, including source, description, and
  license information, for better container metadata.
- Enhanced the transaction modal in `machine_id.html` to handle missing
  command-line data gracefully and improved the visual display for empty items.

### Updates

- Upgraded to Go version **1.24.2*- to fix CVE-2025-22871

### Bug Fixes

- Fixed UI issues in the transaction details modal, ensuring better error
  handling and dynamic updates for missing or empty data.

### Docker Image

```bash
docker pull ghcr.io/txlog/server:v1.6.1
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.6.0...v1.6.1>

## [1.6.0] - 2025-04-11

### Changed

### Features

- We move the asset listing to a new **Assets Page*- (`/assets`) for searching
  assets by hostname or machine ID.
- Enhanced the `/` root page to display aggregated **Operating System*- and
  **Agent version distribution*- statistics.
- Minor UI improvements on the homepage and footer, introducing better styling
  and new icons.

### Bug Fixes

- Addressed error handling issues in controllers, ensuring better logging and
  error responses for database queries.
- Fixed inconsistencies in routing and removed unused paths.

### Docker Image

```bash
docker pull ghcr.io/txlog/server:v1.6.0
```

**Full Changelog**:
[v1.5.0...v1.6.0](<https://github.com/txlog/server/compare/v1.5.0...v1.6.0>)

## [1.5.0] - 2025-04-08

### Changed

### Features

- Added support for searching assets by hostname or machine ID in the root index
  page.
- Introduced a new `transactions` section in the machine details page,
  displaying transaction details for each asset.
- Added a new brand icon for AlmaLinux in the assets directory.
- Enhanced the execution controller to include transactions in the machine
  details page.

### Improvements

- Updated dependency versions in `go.mod` and `go.sum`:
  - `golang.org/x/sync` updated from v0.12.0 to v0.13.0
  - `github.com/tavsec/gin-healthcheck` updated from v1.7.5 to v1.7.6
- Replaced Red Hat brand icon with AlmaLinux brand icon throughout the
  templates.
- Improved SQL queries for better performance when searching assets.

### Bug Fixes

- Corrected environment variable names from `IGNORE_EMPTY_TRANSACTION` to
  `IGNORE_EMPTY_EXECUTION` in the documentation and codebase.
- Fixed an issue where empty transactions were not being ignored correctly in
  the executions controller.
- Improved error handling and logging in various controllers.

### Documentation

- Updated README.md to reflect the changes in environment variable names and new
  features.

## Docker image

```bash
docker pull ghcr.io/txlog/server:v1.5.0
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.4.0...v1.5.0>

## [1.4.0] - 2025-04-03

### Changed

- feat: implement `IGNORE_EMPTY_TRANSACTION` by @rdeavila in
  <https://github.com/txlog/server/pull/25>

## Docker image

```bash
docker pull ghcr.io/txlog/server:v1.4.0
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.3.2...v1.4.0>

## [1.3.2] - 2025-04-03

### Added

- Added `GetMachineID` function to handle machine details requests.
- Introduced a new `GetSponsorIndex` function for sponsor page requests.
- Added new templates: `execution_id.html`, `machine_id.html`, `sponsor.html`.

### Changed

- Refactored routes to include new endpoints for executions and machines.
- Improved query handling and limited total executions displayed to 1000.
- Updated existing templates with new styles and improved layouts.
- Updated `ExecutionID` struct field to include URI binding.

### Deprecated

- Removed unnecessary import of `strconv` in `executions_controller.go`.

## [1.3.1] - 2025-03-26

### Changed

- Increased Pagination Limit: The default pagination limit in
  `root_controller.go` was increased from 10 to 100, allowing for more items to
  be displayed per page.

### Added

- New template functions in `main.go`:
  - `formatPercentage`: Formats a float percentage to a string with comma as the
    decimal separator.
  - `formatInteger`: Formats an integer to a string with thousands separators.
- Tooltip functionality for execution details in `index.html`, showing details
  only when status is a failure, improving the user interface.

### Updated

- Footer link in `footer.html` to point to the specific version tag in the
  repository.
- Modified `index.html` to use the new `formatInteger` and `formatPercentage`
  functions for displaying statistics values.

## [1.3.0] - 2025-03-25

### Added

- Add Agent Version and OS Fields to Executions

## [1.2.0] - 2025-03-25

### Changed

- Bump `github.com/go-co-op/gocron/v2` from 2.15.0 to 2.16.0
- Bump `github.com/tavsec/gin-healthcheck` from 1.7.4 to 1.7.5
- Bump `github.com/go-co-op/gocron/v2` from 2.16.0 to 2.16.1

### Fixed

- Fix for 3rd party Github Actions should be pinned

### Added

- Add interface
- Add statistics

## [1.1.1] - 2025-02-24

### Fixed

- Fix for Potential SQL injection via string-based query concatenation

### Changed

- Bump `github.com/tavsec/gin-healthcheck` from 1.7.3 to 1.7.4

## [1.1] - 2025-02-14

### Fixed

- Fix typo on `macine_id` endpoint

## [1.0] - 2025-02-13

### Changed

- Changed base base image to Scratch
- Refactored API endpoints
- Refactor endpoints and documentation to use plural nouns for consistency

### Added

- Added endpoint to retrieve executions by machine ID and success status
- Added endpoint to retrieve machine IDs by hostname
- Added endpoint to retrieve saved transactions for a host by machine ID
- Added endpoints to retrieve saved items and item IDs for a transaction

## [0.4] - 2025-02-11

### Added

- Add scheduled housekeeping job to delete old executions

### Changed

- Bump Go version to 1.24.0

## [0.3] - 2025-02-08

### Added

- Add Swagger documentation
- Add execution endpoint for creating new executions

## [0.2] - 2025-02-06

### Changed

- Bump `github.com/tavsec/gin-healthcheck` from 1.7.2 to 1.7.3

### Fixed

- Refactor environment variable loading
