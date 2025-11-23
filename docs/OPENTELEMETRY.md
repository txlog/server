# Testing Guide - OpenTelemetry on Txlog Server

This guide shows how to test the OpenTelemetry implementation on Txlog Server.

## Quick Test with Jaeger

The easiest way to test is using Jaeger All-in-One locally.

### 1. Start Jaeger

```bash
docker run -d --name jaeger \
  -p 16686:16686 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

- **16686**: Jaeger UI
- **4318**: OTLP HTTP endpoint

### 2. Configure Txlog Server

Add to your `.env` file:

```bash
# OpenTelemetry Configuration
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
OTEL_SERVICE_NAME=txlog-server
OTEL_SERVICE_VERSION=dev
OTEL_RESOURCE_ATTRIBUTES=deployment.environment=development
```

> **Note**: The endpoint supports `http://` and `https://` prefixes. If `http://` is used, insecure mode is automatically enabled.
> If no prefix is provided, it defaults to `https://`.

### 3. Start Txlog Server

```bash
make run
```

You will see in the logs:

```text
INFO OpenTelemetry: initialized successfully
INFO OpenTelemetry: exporting to http://localhost:4318
```

### 4. Make Requests

Make some requests to generate traces:

```bash
# Main page
curl http://localhost:8080/

# API endpoint
curl http://localhost:8080/v1/version

# Swagger docs
curl http://localhost:8080/swagger/index.html
```

### 5. View Traces in Jaeger

Open your browser at: <http://localhost:16686>

1. In the "Service" field, select **txlog-server**
2. Click "Find Traces"
3. You will see all HTTP requests with their details:
   - Duration
   - HTTP Status
   - Route
   - Method
   - **Executed SQL Queries** (as child spans)

### 6. View SQL Queries in Traces

Each HTTP request that executes SQL queries will show:

**Captured information:**

- SQL query text (without sensitive parameters)
- Execution time
- Number of affected rows
- Errors (if any)
- Database connections

**Examples of SQL spans you will see:**

- `sql:query` - SELECT queries
- `sql:exec` - INSERT, UPDATE, DELETE
- `sql:prepare` - Prepared statements
- `sql:begin` - Transaction start
- `sql:commit` - Transaction commit
- `sql:rollback` - Transaction rollback

**Captured attributes:**

- `db.system`: "postgresql"
- `db.name`: database name
- `db.statement`: SQL query text
- `db.operation`: operation type (SELECT, INSERT, etc.)

### 7. View Correlated Logs

Logs are also sent to Jaeger with trace correlation:

- Click on a specific trace
- You will see the HTTP spans
- Logs will be correlated with trace_id and span_id

## Testing without OpenTelemetry

To test that the application works without OpenTelemetry:

1. **Remove** or **comment out** the `OTEL_EXPORTER_OTLP_ENDPOINT` variable from `.env`

2. Start the server:

```bash
make run
```

3. You will see in the logs:

```text
INFO OpenTelemetry: disabled (OTEL_EXPORTER_OTLP_ENDPOINT not set)
```

4. The application works normally, without sending telemetry

## Testing with OpenTelemetry Collector

For production environments, it is recommended to use an OpenTelemetry Collector.

### 1. Create file `otel-collector-config.yaml`

```yaml
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:

exporters:
  logging:
    loglevel: debug
  jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [logging, jaeger]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [logging]
```

### 2. Docker Compose

```yaml
version: '3.8'

services:
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"
      - "14250:14250"

  otel-collector:
    image: otel/opentelemetry-collector:latest
    command: ["--config=/etc/otel-collector-config.yaml"]
    volumes:
      - ./otel-collector-config.yaml:/etc/otel-collector-config.yaml
    ports:
      - "4318:4318"
    depends_on:
      - jaeger

  txlog-server:
    image: cr.rda.run/txlog/server:main
    environment:
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
      - OTEL_SERVICE_NAME=txlog-server
      - OTEL_RESOURCE_ATTRIBUTES=deployment.environment=production
      # ... other environment variables
    ports:
      - "8080:8080"
    depends_on:
      - otel-collector
```

### 3. Start

```bash
docker-compose up -d
```

## Trace Verification

### What to look for in traces

1. **HTTP Spans**:
   - Span name: HTTP route (e.g., `GET /v1/version`)
   - Attributes:
     - `http.method`: GET, POST, etc.
     - `http.status_code`: 200, 404, etc.
     - `http.route`: request route
     - `http.target`: full URL

2. **SQL Spans** (inside HTTP spans):
   - Span name: SQL operation (e.g., `sql:query SELECT`, `sql:exec INSERT`)
   - Attributes:
     - `db.system`: "postgresql"
     - `db.name`: database name
     - `db.statement`: SQL query text
     - `db.operation`: SELECT, INSERT, UPDATE, DELETE, etc.
   - Hierarchy:
     - HTTP Request (root span)
       - SQL Query 1 (child span)
       - SQL Query 2 (child span)
       - SQL Transaction (child span)
         - SQL Query 3 (grandchild span)
         - SQL Query 4 (grandchild span)

3. **Correlated Logs**:
   - Attribute `trace_id`: trace ID
   - Attribute `span_id`: span ID
   - Allows correlating logs with specific requests

## Troubleshooting

### "Failed to initialize telemetry"

Check:

- Is the OTLP endpoint accessible?
- Is the endpoint format correct? (e.g., `http://host:port` or `host:port`)
  - Prefixes `http://` and `https://` are supported and automatically configure secure/insecure mode.
- Do not include `/v1/traces` in the endpoint - the SDK adds it automatically

### Traces do not appear in Jaeger

1. Check server logs:
   - "OpenTelemetry: initialized successfully" should appear

2. Check if Jaeger is receiving data:

   ```bash
   docker logs jaeger
   ```

3. Test connectivity:

   ```bash
   curl -v http://localhost:4318/v1/traces
   ```

### Performance

OpenTelemetry overhead is typically < 5%:

- Traces are sent in batch (non-blocking)
- Logs are processed asynchronously
- Graceful shutdown ensures data is not lost

## Integration with Production Services

### Grafana Cloud

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=https://otlp-gateway-prod-us-central-0.grafana.net/otlp
OTEL_EXPORTER_OTLP_HEADERS=authorization=Basic $(echo -n "userid:api-key" | base64)
```

### New Relic

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=https://otlp.nr-data.net
OTEL_EXPORTER_OTLP_HEADERS=api-key=YOUR_LICENSE_KEY
```

### Datadog

First configure Datadog Agent to accept OTLP:

```yaml
# datadog.yaml
otlp_config:
  receiver:
    protocols:
      http:
        endpoint: 0.0.0.0:4318
```

Then configure Txlog Server:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://datadog-agent:4318
```

## Success Metrics

After configuring, you should see:

✅ Traces of all HTTP requests
✅ **Traces of all executed SQL queries**
✅ Logs correlated with trace_id
✅ Latency of HTTP and SQL requests
✅ Error rates (4xx, 5xx status and SQL errors)
✅ Throughput (requests per second)
✅ **Performance of individual queries**
✅ **Identification of slow queries (N+1 queries, etc.)**

### Examples of what you can monitor

**HTTP Requests:**

- Slowest endpoint
- Endpoints with most errors
- Throughput by endpoint

**SQL Queries:**

- Slowest queries
- Most frequent queries
- Queries with errors
- Average execution time by query type
- Number of queries per request (detect N+1)

**Transactions:**

- Commit/rollback time
- Transaction success rate
- Deadlocks and lock waits

## Next Steps

1. Configure latency-based alerts (HTTP and SQL)
2. Create dashboards for visualization
3. Use trace sampling in production to reduce cost
4. Configure service level objectives (SLOs)
5. **Identify and optimize slow queries using traces**
6. **Configure alerts for queries exceeding expected time**
