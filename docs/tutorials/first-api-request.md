# Your First API Request

In this tutorial, you will learn how to interact with the Txlog Server API. We will check the server version and
retrieve a list of registered machines.

## Prerequisites

- You have completed the [Getting Started](getting-started.md) tutorial.
- The server is running at `http://localhost:8080`.
- You have `curl` installed (or another API client like Postman).

## Authentication Note

By default, when running in development mode (without OIDC or LDAP configured), the API endpoints are **open** and do
not require an API key.

If you were running in production with authentication enabled, you would need to include the `X-API-Key` header in
your requests.

## Step 1: Check Server Version

The simplest endpoint to test is `/v1/version`. This confirms the API is responsive.

Run the following command:

```bash
curl http://localhost:8080/v1/version
```

**Expected Response:**

```json
{
  "version": "v1.0.0"
}
```

*(Note: The actual version number may vary depending on the build).*

## Step 2: List Machines

Now let's try to fetch data. The `/v1/machines` endpoint lists all active machines (assets) registered with the server.

```bash
curl http://localhost:8080/v1/machines
```

**Expected Response (Empty):**

If you just started the server with a fresh database, the list will be empty:

```json
[]
```

**Expected Response (With Data):**

If you have agents reporting data, the response will look like this:

```json
[
  {
    "hostname": "server-01",
    "machine_id": "a1b2c3d4-e5f6-7890-1234-567890abcdef"
  },
  {
    "hostname": "laptop-dev",
    "machine_id": "f1e2d3c4-b5a6-0987-6543-210987fedcba"
  }
]
```

## Step 3: Filtering Results

You can filter machines by Operating System or Agent Version using query parameters.

Example: Filter by OS

```bash
curl "http://localhost:8080/v1/machines?os=Linux"
```

## Summary

You have successfully:

1. Connected to the API.
2. Retrieved the server version.
3. Queried the machines list.

For a complete list of available endpoints, visit the Swagger documentation at:
[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
