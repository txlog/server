# How to Configure OIDC Authentication

This guide explains how to configure OpenID Connect (OIDC) authentication for the Txlog Server. This allows users to
log in using your existing identity provider (e.g., Google, Keycloak, Okta).

## Prerequisites

- An OIDC Identity Provider (IdP).
- A Client ID and Client Secret from your IdP.
- The Redirect URL registered in your IdP: `https://<your-server-domain>/auth/callback` (or
  `http://localhost:8080/auth/callback` for local dev).

## Configuration Steps

1. **Open your `.env` file** (or configure environment variables in your deployment).

2. **Set the OIDC variables**:

    ```bash
    # The base URL of your Identity Provider
    OIDC_ISSUER_URL=https://accounts.google.com

    # Your Client ID
    OIDC_CLIENT_ID=your-client-id

    # Your Client Secret
    OIDC_CLIENT_SECRET=your-client-secret

    # The callback URL (must match what is registered in the IdP)
    OIDC_REDIRECT_URL=http://localhost:8080/auth/callback

    # Optional: Skip TLS verification (only for testing with self-signed certs)
    OIDC_SKIP_TLS_VERIFY=false
    ```

3. **Restart the Server**.

4. **Verify**:
    - Go to the login page (`/login`).
    - You should see a "Login with OIDC" (or similar) button.
    - Click it to start the authentication flow.

## Troubleshooting

- **"Issuer URL mismatch"**: Ensure `OIDC_ISSUER_URL` exactly matches the `issuer` field in your IdP's discovery
  document (`/.well-known/openid-configuration`).
- **"Redirect URI mismatch"**: Ensure `OIDC_REDIRECT_URL` is exactly the same as registered in the IdP.
