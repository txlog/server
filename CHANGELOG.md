# Changelog

<!-- markdownlint-disable MD024 -->

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<!--
`Added` for new features.
`Changed` for changes in existing functionality.
`Deprecated` for soon-to-be removed features.
`Removed` for now removed features.
`Fixed` for any bug fixes.
`Security` in case of vulnerabilities.
-->

## [Unreleased]

### Changed

- Switch base image from `alpine` to `scratch` in `Dockerfile`
- Added `txlog` user on docker image
- Bumps github.com/swaggo/swag from 1.16.5 to 1.16.6
- Bumps github.com/golang-migrate/migrate/v4 from 4.18.3 to 4.19.0
- Add asset data deletion

## [1.10.2] - 2025-07-17

### Changed

- Bump github.com/tavsec/gin-healthcheck from 1.7.8 to 1.7.9 by
  @dependabot[bot] in <https://github.com/txlog/server/pull/62>
- Bump github.com/swaggo/swag from 1.16.4 to 1.16.5 by
  @dependabot[bot] in <https://github.com/txlog/server/pull/63>

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

- Bump github.com/tavsec/gin-healthcheck from 1.7.7 to 1.7.8 by
  @dependabot in <https://github.com/txlog/server/pull/58>
- Show number of updated packages by week by @rdeavila in
  <https://github.com/txlog/server/pull/59>

## [1.9.1] - 2025-06-05

### Changed

- Optimize asset restart query with ranked executions index

## [1.9.0] - 2025-05-29

### Changed

- Bump github.com/gin-gonic/gin from 1.10.0 to 1.10.1 by
  @dependabot in <https://github.com/txlog/server/pull/55>
- Bump github.com/tavsec/gin-healthcheck from 1.7.6 to 1.7.7 by
  @dependabot in <https://github.com/txlog/server/pull/56>
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
docker pull cr.rda.run/txlog/server:v1.8.1
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.8.0...v1.8.1>

## [1.8.0] - 2025-05-16

### Changed

- Implements restart detection by @rdeavila in
  <https://github.com/txlog/server/pull/52>

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.8.0
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.7.4...v1.8.0>

## [1.7.4] - 2025-05-08

### Changed

- Fix version check URL

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.7.4
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
docker pull cr.rda.run/txlog/server:v1.7.3
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.7.2...v1.7.3>

## [1.7.2] - 2025-05-06

### Changed

- List of Replaced assets and Most updated packages, by @rdeavila in
  <https://github.com/txlog/server/pull/41>

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.7.2
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.7.1...v1.7.2>

## [1.7.1] - 2025-05-01

### Changed

- Enhance transaction display with user action icons and improved user handling

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.7.1
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.7.0...v1.7.1>

## [1.7.0] - 2025-04-30

### Changed

- Add asset listing for OS and Agent distribution by @rdeavila in
  <https://github.com/txlog/server/pull/38>

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.7.0
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.6.3...v1.7.0>

## [1.6.3] - 2025-04-24

### Changed

- Bump github.com/golang-migrate/migrate/v4 from 4.18.2 to 4.18.3
  by @dependabot in <https://github.com/txlog/server/pull/37>

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.6.3
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.6.2...v1.6.3>

## [1.6.2] - 2025-04-22

### Changed

- [Aikido] Fix CVE-2025-22872 in golang.org/x/net via minor version upgrade from
  0.36.0 to 0.38.0 by @aikido-autofix in
  <https://github.com/txlog/server/pull/36>

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.6.2
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
docker pull cr.rda.run/txlog/server:v1.6.1
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
docker pull cr.rda.run/txlog/server:v1.6.0
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
docker pull cr.rda.run/txlog/server:v1.5.0
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.4.0...v1.5.0>

## [1.4.0] - 2025-04-03

### Changed

- feat: implement `IGNORE_EMPTY_TRANSACTION` by @rdeavila in
  <https://github.com/txlog/server/pull/25>

## Docker image

```bash
docker pull cr.rda.run/txlog/server:v1.4.0
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
