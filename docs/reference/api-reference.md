# API Reference

This document provides a high-level reference for the Txlog Server API. For interactive documentation and testing, use
the Swagger UI at `/swagger/index.html`.

## Base URL

`/v1`

## Authentication

- **Header**: `X-API-Key`
- **Required**: Only if OIDC or LDAP is enabled on the server.

## Endpoints

### Assets (Machines)

| Method | Path | Description | Query Params |
| :--- | :--- | :--- | :--- |
| `GET` | `/machines` | List active machines. | `os`, `agent_version` |
| `GET` | `/machines/ids` | Get machine IDs for a hostname. | `hostname` (Required) |
| `GET` | `/assets/requiring-restart` | List assets flagged for restart. | - |

### Executions

| Method | Path | Description | Body |
| :--- | :--- | :--- | :--- |
| `GET` | `/executions` | List recent executions. | - |
| `POST` | `/executions` | Report a new execution. | JSON (Execution object) |

### Transactions

| Method | Path | Description | Body |
| :--- | :--- | :--- | :--- |
| `GET` | `/transactions` | List transactions. | - |
| `GET` | `/transactions/ids` | Get transaction IDs. | - |
| `POST` | `/transactions` | Upload transaction data. | JSON (Transaction object) |

### Packages

| Method | Path | Description | Query Params |
| :--- | :--- | :--- | :--- |
| `GET` | `/packages/:name/:version/:release/assets` | List assets with specific package. | - |

### System

| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/version` | Get server version. |
