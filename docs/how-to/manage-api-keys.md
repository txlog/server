# How to Manage API Keys

API Keys are used by Agents to authenticate with the Txlog Server when pushing data. This guide explains how to create,
revoke, and delete them.

## Prerequisites

- You must be logged in as an **Admin** (via OIDC or LDAP).
- If no authentication is configured on the server, API keys are **not required** for agents, and this section is not applicable.

## Creating an API Key

1. Log in to the Txlog Server.
2. Navigate to the **Admin Panel** (usually `/admin`).
3. Locate the **API Keys** section.
4. Enter a **Name** for the new key (e.g., "Production Cluster A").
5. Click **Create API Key**.
6. **IMPORTANT**: A modal or message will appear showing the full API Key (e.g., `txlog_sk_...`). **Copy this key
    immediately**. It will never be shown again.

## Revoking an API Key

Revoking a key prevents it from being used but keeps the record in the database for audit purposes.

1. Go to the **Admin Panel**.
2. Find the key in the list.
3. Click the **Revoke** button.
4. The key status will change to `Inactive`. Agents using this key will immediately receive `401 Unauthorized` errors.

## Deleting an API Key

Deleting a key removes it permanently from the database.

1. Go to the **Admin Panel**.
2. Find the key in the list.
3. Click the **Delete** button.
4. Confirm the action.

## Using the API Key

Configure your Txlog Agent to use the key by setting the `TXLOG_API_KEY` environment variable or configuration option
on the agent side.
