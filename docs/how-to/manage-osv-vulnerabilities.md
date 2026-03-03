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
3. **Fetching**: The underlying Go worker groups packages by their exact OS (e.g., *AlmaLinux:9*) and performs bulk
   REST API queries.
4. **Scoring Fallback**: The server extracts CVSS scores intrinsically (e.g., translating "Important" to `8.0`).
5. **Re-scoring**: It re-evaluates all historical package installations to map mitigating actions accurately back to
   your dashboard.
