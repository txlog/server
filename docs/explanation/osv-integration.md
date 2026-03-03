# OSV Integration Explained

The Txlog server integrates tightly with the **Open Source Vulnerabilities (OSV)** database schema distributed by
Google via `api.osv.dev`, providing real-time data on vulnerabilities affecting packages installed across your fleet.

This document details the mechanics of how CVE records are queried, parsed, and assigned CVSS risk scores
mathematically, particularly when facing the challenges of ecosystem discrepancies.

## The Challenge with Linux Ecosystems

Package names alone are not enough to query OSV securely because standard Linux packages (e.g., `curl` or
`openssl`) share names across entirely different ecosystems (e.g., Rocky Linux 8 vs. Alpine vs. Debian). If the
server requested pure names, false positives for CVEs that don't apply to a specific distribution would flood
the dashboard.

To resolve this, Txlog queries local tables to extract the `os` origin string (e.g., `"AlmaLinux 9.4"`) for every
package deployment.

1. **Mapping:** The internal function maps generic OS strings into precise OSV namespaces—for instance,
   `AlmaLinux:9` or `Rocky Linux:8`.
2. **Contextualization:** Packages from recent `Install` or `Upgrade` actions are contextualized with these
   namespaces.
3. **Filtering:** The payloads are then aggregated uniquely (`{Package, Ecosystem, Version}`) avoiding deduplication
   and generating concise batches for API traversal.

## The Two-Tier Querying System

Google OSV's Batch API (`/v1/querybatch`) optimizes traffic but severely limits semantic data payload by truncating
fields like summary and description vectors within arrays. To ensure precise mapping while keeping requests fast:

### 1. The Lightning Batch (The `POST` phase)

The scheduled job queries a collection of up to 500 combinations in a single `POST` request to
`api.osv.dev/v1/querybatch`. Since this request strips metadata like `Summary`, `Details`, and the inner array block
(`vuln.Affected`), Txlog strictly uses the index order returned by Google to know which package matched which
resulting CVE list.

### 2. The Detailed Fallback (The `GET` phase)

As the engine iterates over the batches, it builds a local, in-memory **RAM Cache**.

If a CVE (e.g., `ALSA-2022:1988`) returned from the batch doesn't exist deep within the Go RAM Cache from past
iterations, the engine initiates a singular isolated `GET` route back to the OSV (`api.osv.dev/v1/vulns/[ID]`). This
fetches the full un-truncated JSON schema containing full `Summary` strings needed for accurate CVSS abstraction.
The data is cached locally for the rest of the scan.

## Mathematical Semantic Extrapolation

A major limitation with distributions related to Red Hat (RHEL, Alma, Rocky, CentOS) is that the public OSV repository
omits native numeric CVSS metric arrays and strings, leaving Txlog conventionally "blind" to score calculation.

To counter this, Txlog employs a **Natural Language Semantic Engine** across strings during the save phase. Since
RHEL prepends summary strings with keywords, these are mathematically extrapolated:

- **"CRITICAL:"** translates to `CRITICAL` (Synthetic CVSS Score: 9.5)
- **"IMPORTANT:"** translates to `HIGH` (Synthetic CVSS Score: 8.0)
- **"MODERATE:"** translates to `MEDIUM` (Synthetic CVSS Score: 5.5)
- **"LOW:"** translates to `LOW` (Synthetic CVSS Score: 3.0)

If a keyword isn't matched properly, it defaults to `UNKNOWN` (0.0).

With these extrapolated values injected onto the table directly, Txlog successfully renders rich graphical Risk
Mitigated axes and exact Severity trackers correctly.
