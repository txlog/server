# How to Manage OSV Vulnerabilities

This guide explains how administrators can manually manage, update, or completely reset the Open Source
Vulnerabilities (OSV) database records inside the Txlog server.

## Overview

The Txlog server automatically pulls the latest CVEs directly from the Google OSV API on a scheduled interval
(e.g., using a CRON expression). However, there may be times when you need to force an update or completely rebuild
the local vulnerability database if the cached information is outdated or misaligned.

## Option 1: Trigger a Manual Update

If you want to pull down the newest vulnerabilities without waiting for the scheduled job:

1. Log in to your Txlog **Admin** panel as an Administrator.
2. In the OSV section, locate the **Run Manually Now** button.
3. Click the button. A standard confirmation prompt will appear asking if you wish to proceed.
4. Confirm the action. The server will launch a background process to fetch and process the current ecosystem
   mappings against the Google OSV API.
5. You can check the server logs, or monitor the `vulnerabilities` table to see the incoming records.

## Option 2: Reset All Data and Rebuild

If you suspect data corruption, missing mapping for specific Linux environments, or you want to force the Txlog
server to repopulate all CVSS scores from zero:

1. Log in to the **Admin** panel.
2. Under the OSV options, click **Reset All Data & Rebuild**.
3. A critical warning prompt will appear. **Proceeding will truncate the package vulnerabilities and vulnerabilities
   tables** and reset every transaction tracking score to `0`.
4. After confirmation, the database is wiped clean and an intensive API data-fetch job starts.
5. Wait a few moments depending on the size of your asset network. Go to the `Analytics > Security`
   dashboard to watch the scores regenerate incrementally.

### What happens under the hood during a Reset?

1. **Deletion**: `package_vulnerabilities` and `vulnerabilities` tables are wiped.
2. **Transaction Reset**: Scoreboard counters (`vulns_fixed`, `critical_vulns_fixed`, `risk_score_mitigated`, etc.)
   are zeroed out.
3. **Fetching**: The underlying Go worker groups packages by their exact OS ecosystem (e.g., *AlmaLinux:9*,
   *Rocky Linux:9*, or *Red Hat:enterprise_linux:9::baseos*) and performs bulk REST API queries.
4. **Scoring Fallback**: The server extracts CVSS scores intrinsically (e.g., translating "Important" to `8.0`).
5. **Re-scoring**: It re-evaluates all historical package installations to map mitigating actions accurately back to
   your dashboard.

## How Vulnerability Counts Work

### The Security Patch Badge

When a transaction is identified as a security patch, a red shield badge appears next to the transaction ID showing
the number of unique vulnerabilities fixed. This badge is visible on the asset details page.

A transaction is classified as a security patch when it results in a **net reduction** of known vulnerabilities — that
is, when the packages removed or upgraded had more CVEs than the packages installed or upgraded to.

### Counting Logic

Each vulnerability (ALSA/RHSA/RLSA/CVE) is counted **once**, regardless of how many packages it affects. For example,
if advisory `ALSA-2025:23279` (AlmaLinux) or `RHSA-2025:23279` (Red Hat) affects five kernel sub-packages (`kernel`,
`kernel-core`, `kernel-modules`, `kernel-modules-core`, `kernel-modules-extra`), it is counted as **1 vulnerability**,
not 5.

The scoreboard for each transaction is calculated as a delta:

```text
vulns_fixed = (unique CVEs in removed/upgraded-from packages) − (unique CVEs in installed/upgraded-to packages)
```

For a pure removal (e.g., `dnf remove kernel-6.12.0`), the formula becomes:

```text
vulns_fixed = (unique CVEs in removed packages) − 0
```

### Fixed vs. Introduced

- **Fixed**: The vulnerable package version was removed, upgraded, or obsoleted. The CVEs associated with the old
  version are no longer present on the system.
- **Introduced**: A package version with known vulnerabilities was installed or downgraded to. The CVEs associated
  with the new version are now present on the system.

### Severity Classification

Severity levels (Critical, High, Medium, Low) are inferred from the vulnerability metadata provided by the
[OSV database](https://osv.dev). When the OSV API does not provide an explicit severity, the server attempts to
infer it from keywords in the vulnerability summary and details fields.

### Viewing Vulnerability Details

On the asset details page, each transaction row has a split button. The main "Details" button shows the transaction
packages, while the dropdown arrow reveals a "Vulnerabilities" option that opens a dedicated modal listing every CVE
associated with the transaction, including its severity, the affected package, and whether it was fixed or introduced.
The "Vulnerabilities" option is disabled for transactions that have no associated vulnerability data.
