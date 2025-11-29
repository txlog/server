# Database Schema Reference

This document provides a reference for the Txlog Server database schema.

## Tables

### `assets`

Central registry of all managed assets. Tracks asset replacements by linking `hostname` (logical) with `machine_id` (physical).

| Column | Type | Nullable | Description |
| :--- | :--- | :--- | :--- |
| `asset_id` | SERIAL | No | Primary Key (Surrogate). |
| `hostname` | TEXT | No | Logical name (e.g., `web-01`). |
| `machine_id` | TEXT | No | Unique OS/Hardware ID. |
| `first_seen` | TIMESTAMP | No | When first reported. |
| `last_seen` | TIMESTAMP | No | Most recent activity. |
| `is_active` | BOOLEAN | No | `TRUE` if currently active for this hostname. |
| `deactivated_at` | TIMESTAMP | Yes | When replaced by a new asset. |

### `executions`

Records of agent executions (check-ins).

| Column | Type | Nullable | Description |
| :--- | :--- | :--- | :--- |
| `id` | SERIAL | No | Primary Key. |
| `machine_id` | TEXT | No | Foreign Key to Asset (logical). |
| `hostname` | TEXT | No | Hostname at time of execution. |
| `executed_at` | TIMESTAMP | No | Time of execution. |
| `success` | BOOLEAN | No | Whether execution succeeded. |
| `details` | TEXT | Yes | Error messages or details. |
| `os` | TEXT | Yes | Operating System name/version. |
| `agent_version` | TEXT | Yes | Version of Txlog Agent. |

### `transactions`

RPM transactions (install, update, remove).

| Column | Type | Nullable | Description |
| :--- | :--- | :--- | :--- |
| `transaction_id` | INTEGER | No | RPM DB Transaction ID. |
| `machine_id` | TEXT | No | Composite PK with `transaction_id`. |
| `hostname` | TEXT | No | Hostname. |
| `begin_time` | TIMESTAMP | Yes | Start of transaction. |
| `end_time` | TIMESTAMP | Yes | End of transaction. |
| `actions` | TEXT | Yes | Summary of actions. |
| `user` | TEXT | Yes | User who ran the command. |
| `command_line` | TEXT | Yes | Command executed (e.g., `dnf update`). |
| `return_code` | TEXT | Yes | Exit code. |

### `transaction_items`

Individual packages affected in a transaction.

| Column | Type | Nullable | Description |
| :--- | :--- | :--- | :--- |
| `item_id` | SERIAL | No | Primary Key. |
| `transaction_id` | INTEGER | No | FK to `transactions`. |
| `machine_id` | TEXT | No | FK to `transactions`. |
| `package` | TEXT | Yes | Package name. |
| `version` | TEXT | Yes | Package version. |
| `release` | TEXT | Yes | Package release. |
| `action` | TEXT | Yes | Action (Install, Upgrade, Remove). |

### `users`

Admin panel users (OIDC/LDAP).

| Column | Type | Nullable | Description |
| :--- | :--- | :--- | :--- |
| `id` | SERIAL | No | Primary Key. |
| `email` | TEXT | No | User email. |
| `is_admin` | BOOLEAN | No | Full access flag. |
| `is_active` | BOOLEAN | No | Login permission flag. |

### `api_keys`

API keys for agent authentication.

| Column | Type | Nullable | Description |
| :--- | :--- | :--- | :--- |
| `id` | SERIAL | No | Primary Key. |
| `name` | TEXT | No | Human-readable name. |
| `key_prefix` | TEXT | No | First few chars of key. |
| `key_hash` | TEXT | No | Hashed key (SHA-256). |
| `is_active` | BOOLEAN | No | Valid for use. |
