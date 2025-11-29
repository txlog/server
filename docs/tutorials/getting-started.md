# Getting Started with Txlog Server

This tutorial will guide you through setting up the Txlog Server development environment on your local machine. By the
end of this guide, you will have a running server connected to a local database, ready to receive data.

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go** (version 1.22 or higher)
- **Docker** (to run the database)
- **Make** (to run build commands)
- **Git**

## Step 1: Clone the Repository

First, clone the repository to your local machine:

```bash
git clone https://github.com/txlog/server.git
cd server
```

## Step 2: Start the Database

Txlog Server requires a PostgreSQL database. The easiest way to run one for development is using Docker.

Run the following command to start a Postgres container:

```bash
docker run -d --name txlog-postgres \
  -e POSTGRES_USER=txlog \
  -e POSTGRES_PASSWORD=secret \
  -e POSTGRES_DB=txlog \
  -p 5432:5432 \
  postgres:15-alpine
```

This command:

- Starts a container named `txlog-postgres`.
- Sets the user, password, and database name to `txlog` (and password `secret`).
- Exposes port `5432` locally.

## Step 3: Configure the Environment

Create a `.env` file in the root of the project to configure the server. You can copy the following configuration,
which matches the database we just started:

```bash
# Server Configuration
INSTANCE=Local Dev
LOG_LEVEL=DEBUG
GIN_MODE=debug
PORT=8080

# Database Configuration
PGSQL_HOST=127.0.0.1
PGSQL_PORT=5432
PGSQL_USER=txlog
PGSQL_DB=txlog
PGSQL_PASSWORD=secret
PGSQL_SSLMODE=disable

# Cron Configuration (for background tasks)
CRON_RETENTION_DAYS=7
CRON_RETENTION_EXPRESSION=0 2 * * *
CRON_STATS_EXPRESSION=0 * * * *
```

> **Note**: We set `PGSQL_SSLMODE=disable` because the standard Postgres Docker image does not have SSL configured by default.

## Step 4: Install Dependencies and Tools

The project uses `air` for live reloading and `swag` for API documentation. Install them using the provided commands:

```bash
# Install Air (for live reload)
curl -fsSL https://install.rda.run/air-verse/air@latest! | bash

# Install Swaggo (for API docs)
curl -fsSL https://install.rda.run/swaggo/swag@latest! | bash
```

## Step 5: Run the Server

Now you are ready to start the server. Use the `make run` command, which uses `air` to watch for file changes and
restart the server automatically.

```bash
make run
```

You should see output indicating the server has started:

```text
...
[INFO] server/main.go:47: Logger initialized
[INFO] server/main.go:53: Database connected
[INFO] server/scheduler/main.go:31: Scheduler: started.
[INFO] server/main.go:80: No authentication configured - API endpoints accessible without API key
[GIN-debug] Listening and serving HTTP on :8080
```

## Step 6: Verify Installation

Open your web browser and navigate to:

[http://localhost:8080](http://localhost:8080)

You should see the Txlog Server dashboard. Since this is a fresh database, it will be empty, but it confirms the
application is running and connected to the database.

---

**Next Steps**: Now that your server is running, check out the [Your First API Request](first-api-request.md) tutorial
to learn how to interact with it programmatically.
