# Using Reports Features

Txlog Server provides reports endpoints to gain insights into package management across your fleet.

## Overview

| Feature | Description | UI Path | API Endpoint |
| :--- | :--- | :--- | :--- |
| Package Comparison | Compare installed packages between assets | `/analytics/compare` | `GET /v1/reports/compare-packages` |
| Package Freshness | Analyze how current deployed versions are | `/analytics/freshness` | `GET /v1/reports/package-freshness` |
| Package Adoption | See how widely packages are adopted | `/analytics/adoption` | `GET /v1/reports/package-adoption` |
| Anomaly Detection | Detect unusual transaction patterns | `/analytics/anomalies` | `GET /v1/reports/anomalies` |

## Accessing via UI

1. Navigate to **Reports** in the top navigation menu.
2. Select the desired report.
3. Configure filters and view results.

### Quick Access

- **From Dashboard**: The home page shows recent anomalies with a "View All" link.
- **From Asset Page**: Click "Compare with..." to compare packages with other assets.

## Using the API

### Compare Packages

Compare packages between 2-20 assets:

```bash
curl "http://localhost:8080/v1/reports/compare-packages?machine_ids=id1,id2,id3"
```

**Response includes:**

- `common`: Packages present on all assets with identical versions.
- `different`: Packages with version differences across assets.
- `only_in`: Packages unique to specific assets.

### Package Freshness

Analyze package age based on first-seen date:

```bash
# System-wide analysis
curl "http://localhost:8080/v1/reports/package-freshness?limit=50"

# Specific asset
curl "http://localhost:8080/v1/reports/package-freshness?machine_id=abc123&limit=20"
```

**Response includes:**

- `average_age_days`: Average package version age.
- `oldest_package` / `newest_package`: Extremes.
- `packages`: List with age details.

### Package Adoption

See how widely packages are adopted across active assets:

```bash
curl "http://localhost:8080/v1/reports/package-adoption?limit=50&min_assets=5"
```

**Response includes:**

- `total_active_assets`: Count of active assets.
- `packages`: List with adoption percentage and update frequency.

### Anomaly Detection

Detect unusual patterns like high-volume transactions or rapid package changes:

```bash
# Last 7 days, all severities
curl "http://localhost:8080/v1/reports/anomalies?days=7"

# Last 30 days, high severity only
curl "http://localhost:8080/v1/reports/anomalies?days=30&severity=high"
```

**Anomaly types detected:**

| Type | Description | Threshold |
| :--- | :--- | :--- |
| `high_volume` | Transaction with many packages | >50 packages |
| `rapid_change` | Same package changed frequently | >3 times in 24h |
| `downgrade` | Package version downgrade | Any occurrence |

**Severity levels:**

- **low**: Downgrades (informational).
- **medium**: High volume (50-100 packages) or rapid changes (4-5 times).
- **high**: Very high volume (>100 packages) or extreme rapid changes (>5 times).

## Use Cases

1. **Drift Detection**: Compare package sets between similar servers to find configuration drift.
2. **Update Planning**: Use freshness analysis to identify outdated packages needing updates.
3. **Rollout Monitoring**: Track package adoption to verify deployment success.
4. **Incident Detection**: Monitor anomalies for unusual activity that may indicate problems.
