# OSV Integration Explained

The Txlog server integrates tightly with the **Open Source Vulnerabilities (OSV)** database schema distributed by
Google via `api.osv.dev`, providing real-time data on vulnerabilities affecting packages installed across your fleet.

This document details how vulnerability records are queried, parsed, and assigned severity scores, with particular
attention to how each Linux distribution is mapped to its correct OSV ecosystem.

## Ecosystem Isolation by Distribution

The OSV database maintains **separate advisory databases** for each Linux distribution. Each distribution publishes
its own security advisories with distinct identifiers:

| Distribution | OSV Ecosystem | Advisory Prefix | Example |
| --- | --- | --- | --- |
| AlmaLinux | `AlmaLinux:<major>` | ALSA | `ALSA-2024:5529` |
| Rocky Linux | `Rocky Linux:<major>` | RLSA | `RLSA-2024:5529` |
| Red Hat Enterprise Linux | `Red Hat:<CPE>` | RHSA | `RHSA-2024:5529` |

### Why This Matters

Package names alone are not enough to query OSV securely because standard Linux packages (e.g., `curl` or
`openssl`) share names across entirely different ecosystems. If the server requested names without the correct
ecosystem, vulnerabilities from the wrong distribution would appear — for example, ALSA (AlmaLinux) advisories
showing up on Red Hat systems, or RHSA (Red Hat) advisories on AlmaLinux systems.

### How Txlog Maps Each Distribution

Txlog queries the `assets.os` field (e.g., `"AlmaLinux 9.4"`, `"Red Hat Enterprise Linux 9.2"`) and the
`transaction_items.repo` field (e.g., `"baseos"`, `"appstream"`) to build the correct ecosystem identifier:

1. **AlmaLinux** → `AlmaLinux:<major>` (e.g., `AlmaLinux:9`)
2. **Rocky Linux** → `Rocky Linux:<major>` (e.g., `Rocky Linux:9`)
3. **Red Hat / CentOS / Oracle Linux** → `Red Hat:enterprise_linux:<major>::<channel>`
   (e.g., `Red Hat:enterprise_linux:9::baseos`)

The Red Hat ecosystem in OSV uses CPE-based identifiers that include a repository channel suffix (`baseos`,
`appstream`, or `crb`). Txlog derives the correct channel from the `repo` column in `transaction_items`.

CentOS and Oracle Linux do not have their own ecosystems in OSV and return `"Invalid ecosystem"` when queried
directly. Since they are binary-compatible with RHEL, Txlog maps them to the Red Hat advisories (RHSA) as the
closest match.

## The Two-Tier Querying System

Google OSV's Batch API (`/v1/querybatch`) optimizes traffic but severely limits semantic data payload by truncating
fields like summary and description vectors within arrays. To ensure precise mapping while keeping requests fast:

### 1. The Lightning Batch (The `POST` phase)

The scheduled job queries a collection of up to 500 combinations in a single `POST` request to
`api.osv.dev/v1/querybatch`. Since this request strips metadata like `Summary`, `Details`, and the inner array block
(`vuln.Affected`), Txlog strictly uses the index order returned by Google to know which package matched which
resulting advisory list.

### 2. The Detailed Fallback (The `GET` phase)

As the engine iterates over the batches, it builds a local, in-memory **RAM Cache**.

If an advisory (e.g., `ALSA-2022:1988` for AlmaLinux, `RHSA-2022:1988` for Red Hat, or `RLSA-2022:1988` for Rocky
Linux) returned from the batch doesn't exist in the Go RAM Cache from past iterations, the engine initiates an
isolated `GET` to `api.osv.dev/v1/vulns/[ID]`. This fetches the full un-truncated JSON schema containing full
`Summary` strings needed for accurate severity extraction. The data is cached locally for the rest of the scan.

## Severity and Score Extraction

Txlog uses a multi-tier approach to determine severity and CVSS scores:

### 1. CVSS Vector Parsing

When the OSV record includes a `severity[]` array with a CVSS 3.x vector string, Txlog computes an approximate
CVSS base score from the vector components (AV, AC, PR, UI, S, C, I, A).

### 2. Distribution-Specific Severity

The `database_specific.severity` field is used as a primary source for distributions like AlmaLinux and Rocky Linux,
which include labels such as `"Important"`, `"Critical"`, `"Moderate"`, or `"Low"`.

### 3. Text-Based Fallback

When neither structured severity data nor CVSS vectors are available, Txlog employs keyword matching in the
`summary` and `details` fields. Since RHEL-family advisories prepend summary strings with severity keywords, these
are mapped to standard labels:

- **"CRITICAL"** → `CRITICAL` (Synthetic CVSS Score: 9.5)
- **"IMPORTANT"** → `HIGH` (Synthetic CVSS Score: 8.0)
- **"MODERATE"** → `MEDIUM` (Synthetic CVSS Score: 5.5)
- **"LOW"** → `LOW` (Synthetic CVSS Score: 3.0)

If no keyword is matched, it defaults to `UNKNOWN` (0.0).

With these extracted values stored in the database, Txlog successfully renders Risk Mitigated charts and
Severity trackers on the dashboard.
