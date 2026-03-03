# How to Detect Transaction Anomalies

Txlog Server analyzes transaction data to identify unusual patterns, such as bulk package
updates or rapid, repeated changes to identical packages. This functionality allows you
to detect operational issues or security concerns early.

## Identifying Anomalies via the UI

1. Open the Txlog Server dashboard and navigate to **Reports** > **Anomaly Detection**.
2. From the time filter, select the period you wish to analyze (e.g., Last 7 Days, Last 30 Days).
3. Optionally, filter by **Severity** (Low, Medium, High).
4. Results are presented in a table detailing:
   * **Host**: The asset experiencing the anomaly.
   * **Type**: The class of anomaly (`high_volume`, `rapid_change`, or `downgrade`).
   * **Details**: Additional context, such as the number of packages altered in a single transaction.

## Retrieving Anomalies via the API

You can programmatically fetch anomalies using the `/v1/reports/anomalies` endpoint.

### Basic Query

To check for all anomalies within a specific number of days:

```bash
curl "http://localhost:8080/v1/reports/anomalies?days=7" -H "X-API-Key: YOUR_API_KEY"
```

### Filtering by Severity

You can narrow your analysis to focus exclusively on highly severe events:

```bash
curl "http://localhost:8080/v1/reports/anomalies?days=30&severity=high" -H "X-API-Key: YOUR_API_KEY"
```

## Understanding Anomaly Types

* **`high_volume`**: Occurs when a single transaction modifies an unusually large number
  of packages (e.g., >50 for medium severity, >100 for high severity). Use this to track bulk operations.
* **`rapid_change`**: Triggered when the same package is updated or reinstalled repeatedly
  over a short time frame (e.g., >3 times in 24 hours). This may indicate configuration flapping
  or failing upgrades.
* **`downgrade`**: Fired when a package configuration intentionally regresses to an older
  version. Downgrades may signify an attempt to circumvent security patches or a quick rollback
  of a failing application deployment.
