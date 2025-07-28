# Changelog

<!-- markdownlint-disable MD024 -->

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog], and this project adheres to [Semantic
Versioning].

## [Unreleased]

- chore: switch base image from alpine to scratch in Dockerfile

## [1.10.2] - 2025-07-17

## What's Changed

- chore(deps): bump github.com/tavsec/gin-healthcheck from 1.7.8 to 1.7.9 by
  @dependabot[bot] in <https://github.com/txlog/server/pull/62>
- chore(deps): bump github.com/swaggo/swag from 1.16.4 to 1.16.5 by
  @dependabot[bot] in <https://github.com/txlog/server/pull/63>

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.10.2
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.10.1...v1.10.2>

## [1.10.1] - 2025-07-14

## What's Changed in v1.10.1

- Increased the time range for the "updated packages by week" graph from 12 to
  15 weeks for improved historical insight.
- Updated frontend dependencies:
  - Bootstrap upgraded from 5.3.6 to 5.3.7
  - Tabler upgraded from 1.3.2 to 1.4.0 (CSS & JS)
  - ApexCharts upgraded from 4.7.0 to 5.2.0 (CSS & JS)

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.10.1
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.10.0...v1.10.1>

## [1.10.0] - 2025-07-10

## What's Changed

- chore(deps): bump github.com/tavsec/gin-healthcheck from 1.7.7 to 1.7.8 by
  @dependabot in <https://github.com/txlog/server/pull/58>
- Show number of updated packages by week by @rdeavila in
  <https://github.com/txlog/server/pull/59>

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.10.0
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.9.1...v1.10.0>

## [1.9.1] - 2025-06-05

## What's Changed

- feat: optimize asset restart query with ranked executions index

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.9.1
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.9.0...v1.9.1>

## [1.9.0] - 2025-05-29

## What's Changed

- chore(deps): bump github.com/gin-gonic/gin from 1.10.0 to 1.10.1 by
  @dependabot in <https://github.com/txlog/server/pull/55>
- chore(deps): bump github.com/tavsec/gin-healthcheck from 1.7.6 to 1.7.7 by
  @dependabot in <https://github.com/txlog/server/pull/56>
- Implement Restart Tracking for Assets (UI & API) by @rdeavila in
  <https://github.com/txlog/server/pull/57>

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.9.0
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.8.3...v1.9.0>

## [1.8.3] - 2025-05-20

## What's Changed

- Added HTML escaping in `Text2HTML` function to prevent potential XSS
  vulnerabilities
- Bump tabler version to 1.3.0

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.8.3
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.8.2...v1.8.3>

## [1.8.2] - 2025-05-16

## What's Changed

- Update brand detection for Red Hat
- Add utility functions for template processing
- Fix settings page

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.8.2
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.8.1...v1.8.2>

## [1.8.1] - 2025-05-16

## What's Changed

- Fix asset search
- Fix execution listing from agent

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.8.1
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.8.0...v1.8.1>

## [1.8.0] - 2025-05-16

## What's Changed

- Implements restart detection by @rdeavila in
  <https://github.com/txlog/server/pull/52>

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.8.0
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.7.4...v1.8.0>

## [1.7.4] - 2025-05-08

## What's Changed

- Fix version check URL

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.7.4
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.7.3...v1.7.4>

## [1.7.3] - 2025-05-06

## What's Changed

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

## What's Changed

- List of Replaced assets and Most updated packages, by @rdeavila in
  <https://github.com/txlog/server/pull/41>

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.7.2
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.7.1...v1.7.2>

## [1.7.1] - 2025-05-01

## What's Changed

- Enhance transaction display with user action icons and improved user handling

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.7.1
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.7.0...v1.7.1>

## [1.7.0] - 2025-04-30

## What's Changed

- Add asset listing for OS and Agent distribution by @rdeavila in
  <https://github.com/txlog/server/pull/38>

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.7.0
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.6.3...v1.7.0>

## [1.6.3] - 2025-04-24

## What's Changed

- chore(deps): bump github.com/golang-migrate/migrate/v4 from 4.18.2 to 4.18.3
  by @dependabot in <https://github.com/txlog/server/pull/37>

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.6.3
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.6.2...v1.6.3>

## [1.6.2] - 2025-04-22

## What's Changed

- [Aikido] Fix CVE-2025-22872 in golang.org/x/net via minor version upgrade from
  0.36.0 to 0.38.0 by @aikido-autofix in
  <https://github.com/txlog/server/pull/36>

### Docker Image

```bash
docker pull cr.rda.run/txlog/server:v1.6.2
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.6.1...v1.6.2>

## [1.6.1] - 2025-04-14

## What's Changed

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

## What's Changed

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

## What's Changed

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

## What's Changed

- feat: implement `IGNORE_EMPTY_TRANSACTION` by @rdeavila in
  <https://github.com/txlog/server/pull/25>

## Docker image

```bash
docker pull cr.rda.run/txlog/server:v1.4.0
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.3.2...v1.4.0>

## [1.3.2] - 2025-04-03

## What's Changed

- **Major Changes**
  - Added `GetMachineID` function to handle machine details requests.
  - Introduced a new `GetSponsorIndex` function for sponsor page requests.
  - Refactored routes to include new endpoints for executions and machines.
  - Improved query handling and limited total executions displayed to 1000.

- **Template Changes**
  - Added new templates: `execution_id.html`, `machine_id.html`, `sponsor.html`.
  - Updated existing templates with new styles and improved layouts.

- **Minor Changes**
  - Removed unnecessary import of `strconv` in `executions_controller.go`.
  - Updated `ExecutionID` struct field to include URI binding.

## Docker image

```bash
docker pull cr.rda.run/txlog/server:v1.3.2
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.3.1...v1.3.2>

## [1.3.1] - 2025-03-26

## What's Changed

- Increased Pagination Limit: The default pagination limit in
  `root_controller.go` was increased from 10 to 100, allowing for more items to
  be displayed per page.
- Template Functions: Added new template functions in `main.go`:
  - `formatPercentage`: Formats a float percentage to a string with comma as the
    decimal separator.
  - `formatInteger`: Formats an integer to a string with thousands separators.
- Footer Links: Updated the footer link in `footer.html` to point to the
  specific version tag in the repository.
- Statistics Formatting: Modified `index.html` to use the new `formatInteger`
  and `formatPercentage` functions for displaying statistics values.
- Tooltip for Execution Details: Added tooltip functionality for execution
  details in `index.html`, showing details only when status is a failure,
  improving the user interface.

## Docker image

```bash
docker pull cr.rda.run/txlog/server:v1.3.1
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.3.0...v1.3.1>

## [1.3.0] - 2025-03-25

## What's Changed

- Add Agent Version and OS Fields to Executions by @rdeavila in
  <https://github.com/txlog/server/pull/21>

## Docker image

```bash
docker pull cr.rda.run/txlog/server:v1.3.0
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.2.0...v1.3.0>

## [1.2.0] - 2025-03-25

## What's Changed

- Bump github.com/go-co-op/gocron/v2 from 2.15.0 to 2.16.0 by @dependabot in
  <https://github.com/txlog/server/pull/12>
- Bump github.com/tavsec/gin-healthcheck from 1.7.4 to 1.7.5 by @dependabot in
  <https://github.com/txlog/server/pull/14>
- Bump github.com/go-co-op/gocron/v2 from 2.16.0 to 2.16.1 by @dependabot in
  <https://github.com/txlog/server/pull/15>
- [Aikido AI] Fix for 3rd party Github Actions should be pinned by
  @aikido-autofix in <https://github.com/txlog/server/pull/18>
- Add interface by @rdeavila in <https://github.com/txlog/server/pull/13>
- Add statistics by @rdeavila in <https://github.com/txlog/server/pull/20>

## Docker image

```bash
docker pull cr.rda.run/txlog/server:v1.2.0
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.1.1...v1.2.0>

## [1.1.1] - 2025-02-24

## What's Changed

- Fix for Potential SQL injection via string-based query concatenation by
  @aikido-autofix in <https://github.com/txlog/server/pull/10>
- Bump github.com/tavsec/gin-healthcheck from 1.7.3 to 1.7.4 by @dependabot in
  <https://github.com/txlog/server/pull/11>

## New Contributors

- @aikido-autofix made their first contribution in
  <https://github.com/txlog/server/pull/10>

## Docker image

```bash
docker pull cr.rda.run/txlog/server:v1.1.1
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.1...v1.1.1>

## [1.1] - 2025-02-14

## What's Changed

- Fix typo on `macine_id` endpoint

## Docker image

```bash
docker pull cr.rda.run/txlog/server:v1.1
```

**Full Changelog**: <https://github.com/txlog/server/compare/v1.0...v1.1>

## [1.0] - 2025-02-13

## What's Changed

- Changed base base image to Scratch
- Refactored API endpoints
- Added endpoint to retrieve executions by machine ID and success status
- Added endpoint to retrieve machine IDs by hostname
- Added endpoint to retrieve saved transactions for a host by machine ID
- Added endpoints to retrieve saved items and item IDs for a transaction
- Refactor endpoints and documentation to use plural nouns for consistency

## Docker image

```bash
docker pull cr.rda.run/txlog/server:v1.0
```

**Full Changelog**: <https://github.com/txlog/server/compare/v0.4...v1.0>

## [0.4] - 2025-02-11

- [Add scheduled housekeeping job to delete old
  executions](https://github.com/txlog/server/commit/864ce48e7dd44003d3846282cb2bfb47fdbc97d2)
- [Bump Go version to
  1.24.0](https://github.com/txlog/server/commit/289c492f7f32aae70de7befb82cc001d8786357e)

## Docker image

```bash
docker pull cr.rda.run/txlog/server:v0.4
```

**Full Changelog**: <https://github.com/txlog/server/compare/v0.3...v0.4>

## [0.3] - 2025-02-08

- Add Swagger documentation
- Add execution endpoint for creating new executions

## Docker image

```bash
docker pull cr.rda.run/txlog/server:v0.3
```

**Full Changelog**: <https://github.com/txlog/server/compare/v0.2...v0.3>

## [0.2] - 2025-02-06

## What's Changed

- Bump github.com/tavsec/gin-healthcheck from 1.7.2 to 1.7.3 by @dependabot in
  <https://github.com/txlog/server/pull/7>
- [Refactor environment variable
  loading](https://github.com/txlog/server/commit/2952563ddf8c1258099b9d8e5718dcbb862fc431).
  Fixes <https://github.com/txlog/server/issues/6>

## Docker image

```bash
docker pull cr.rda.run/txlog/server:v0.2
```

**Full Changelog**: <https://github.com/txlog/server/compare/v0.1...v0.2>
