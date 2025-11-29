# LDAP Authentication Guide

This document explains how to configure and use LDAP authentication with Txlog
Server.

## Overview

Txlog Server now supports LDAP (Lightweight Directory Access Protocol)
authentication in addition to the existing OIDC authentication. LDAP
authentication allows users to log in using their enterprise directory
credentials with group-based access control.

## Features

- **Username/Password Authentication**: Users authenticate with their LDAP
  credentials
- **Group-Based Authorization**: Define admin and viewer groups in LDAP
- **Flexible Configuration**: Support for various LDAP server configurations
- **TLS Support**: Optional TLS/LDAPS connections
- **Service Account Binding**: Optional service account for group membership
  lookups
- **Dual Authentication**: Can be enabled alongside OIDC, allowing users to
  choose their preferred method

## Configuration

### Required Environment Variables

The following environment variables are **required** to enable LDAP
authentication:

```bash
LDAP_HOST=ldap.example.com          # LDAP server hostname
LDAP_BASE_DN=ou=users,dc=example,dc=com  # Base DN for user searches
```

Additionally, at least one of these group DNs must be configured:

```bash
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com      # Admin group DN
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com    # Viewer group DN
```

### Optional Environment Variables

```bash
# Connection Settings
LDAP_PORT=389                       # Default: 389 (LDAP) or 636 (LDAPS)
LDAP_USE_TLS=false                  # Enable TLS connection (default: false)
LDAP_SKIP_TLS_VERIFY=false          # Skip TLS cert verification (default: false)

# Service Account (OPTIONAL - for group membership checks)
# If not provided, uses anonymous bind for search and authenticated user session for group checks
LDAP_BIND_DN=cn=admin,dc=example,dc=com          # Service account DN
LDAP_BIND_PASSWORD=your_bind_password            # Service account password

# Search Filters
LDAP_USER_FILTER=(uid=%s)           # User search filter (default: (uid=%s))
LDAP_GROUP_FILTER=(member=%s)       # Group membership filter (default: (member=%s))
```

## Group-Based Authorization

Txlog Server implements a two-tier authorization model:

### Admin Group

Users in the **admin group** have full access to:

- View all transaction data
- Manage assets and packages
- Access the admin panel
- Create and manage API keys
- Manage user accounts

### Viewer Group

Users in the **viewer group** have read-only access to:

- View transaction data
- View assets and packages
- View insights and statistics

**Note**: A user can be a member of both groups. If a user is in the admin
group, they will have admin privileges regardless of viewer group membership.

## Configuration Examples

### Example 1: Basic LDAP Setup (No TLS, No Service Account)

This configuration works when:

- Your LDAP allows anonymous bind for user searches
- Authenticated users can read their own group memberships

```bash
LDAP_HOST=ldap.company.com
LDAP_PORT=389
LDAP_BASE_DN=ou=employees,dc=company,dc=com
LDAP_USER_FILTER=(uid=%s)
LDAP_ADMIN_GROUP=cn=txlog-admins,ou=groups,dc=company,dc=com
LDAP_VIEWER_GROUP=cn=txlog-users,ou=groups,dc=company,dc=com
LDAP_GROUP_FILTER=(member=%s)
# Note: LDAP_BIND_DN and LDAP_BIND_PASSWORD are NOT set
```

**How it works without service account:**

1. Anonymous bind to search for user
2. User authentication via bind with user credentials
3. Group membership check using authenticated user session

### Example 2: LDAPS with Service Account

```bash
LDAP_HOST=ldaps.company.com
LDAP_PORT=636
LDAP_USE_TLS=true
LDAP_BIND_DN=cn=svc-txlog,ou=services,dc=company,dc=com
LDAP_BIND_PASSWORD=SecureServicePassword123
LDAP_BASE_DN=ou=employees,dc=company,dc=com
LDAP_USER_FILTER=(sAMAccountName=%s)
LDAP_ADMIN_GROUP=cn=IT-Admins,ou=security,dc=company,dc=com
LDAP_VIEWER_GROUP=cn=All-Staff,ou=security,dc=company,dc=com
LDAP_GROUP_FILTER=(member=%s)
```

### Example 3: Active Directory Configuration

```bash
LDAP_HOST=dc01.domain.local
LDAP_PORT=389
LDAP_BASE_DN=cn=Users,dc=domain,dc=local
LDAP_USER_FILTER=(sAMAccountName=%s)
LDAP_BIND_DN=cn=Service Account,cn=Users,dc=domain,dc=local
LDAP_BIND_PASSWORD=YourADPassword
LDAP_ADMIN_GROUP=cn=TxlogAdmins,ou=Groups,dc=domain,dc=local
LDAP_VIEWER_GROUP=cn=TxlogViewers,ou=Groups,dc=domain,dc=local
LDAP_GROUP_FILTER=(member=%s)
```

### Example 4: OpenLDAP with posixGroup (No Service Account)

For OpenLDAP using `posixGroup` schema with anonymous search:

```bash
LDAP_HOST=openldap.example.com
LDAP_PORT=389
LDAP_BASE_DN=ou=people,dc=example,dc=com
LDAP_USER_FILTER=(uid=%s)
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=users,ou=groups,dc=example,dc=com
LDAP_GROUP_FILTER=(memberUid=%s)
# No LDAP_BIND_DN or LDAP_BIND_PASSWORD
```

### Example 5: OpenLDAP with Service Account

For restricted OpenLDAP that requires authentication:

```bash
LDAP_HOST=openldap.example.com
LDAP_PORT=389
LDAP_BIND_DN=cn=readonly,dc=example,dc=com
LDAP_BIND_PASSWORD=readonly_password
LDAP_BASE_DN=ou=people,dc=example,dc=com
LDAP_USER_FILTER=(uid=%s)
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=users,ou=groups,dc=example,dc=com
LDAP_GROUP_FILTER=(memberUid=%s)
```

### Example 6: Both OIDC and LDAP Enabled

You can enable both authentication methods simultaneously:

```bash
# OIDC Configuration
OIDC_ISSUER_URL=https://id.example.com
OIDC_CLIENT_ID=your_client_id
OIDC_CLIENT_SECRET=your_client_secret
OIDC_REDIRECT_URL=https://txlog.example.com/auth/callback

# LDAP Configuration
LDAP_HOST=ldap.example.com
LDAP_BASE_DN=ou=users,dc=example,dc=com
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
```

When both are enabled, users will see a login page where they can choose between LDAP (username/password) or OIDC authentication.

## LDAP Attribute Mapping

The server extracts the following user information from LDAP:

- **Username**: From the login form (used as unique identifier with `ldap:` prefix)
- **Email**: From `mail` attribute (fallback: `username@local`)
- **Display Name**: From `cn` or `displayName` attributes (fallback: username)

## Security Considerations

### TLS/LDAPS

- Always use TLS in production environments
- Set `LDAP_USE_TLS=true` or use port 636 for LDAPS
- Only use `LDAP_SKIP_TLS_VERIFY=true` for self-signed certificates in development

### Service Account (Optional)

- **Not required** if your LDAP allows anonymous searches and users can read group memberships
- **Recommended** for production environments with restricted LDAP access
- When used:
  - Use a dedicated service account with minimal privileges
  - The service account only needs read access to user and group objects
  - Restrict the service account to specific OUs if possible
- When **not** used:
  - Ensure LDAP allows anonymous bind for user searches
  - Ensure authenticated users can read their own group memberships
  - May not work with all LDAP configurations (e.g., Active Directory typically requires a service account)

### Password Handling

- Passwords are never stored; they're only used for authentication
- Failed authentication attempts are logged
- Consider implementing rate limiting at the reverse proxy level

## Troubleshooting

### Common Issues

1. **"Failed to connect to LDAP"**
   - Check `LDAP_HOST` and `LDAP_PORT` are correct
   - Verify network connectivity to LDAP server
   - Check firewall rules

2. **"User not found"**
   - Verify `LDAP_BASE_DN` is correct
   - Check `LDAP_USER_FILTER` matches your LDAP schema
   - Ensure user exists in the specified base DN

3. **"Invalid credentials"**
   - Username or password is incorrect
   - Check if user account is locked or disabled

4. **"User is not a member of any authorized group"**
   - User exists but doesn't belong to admin or viewer groups
   - Verify `LDAP_ADMIN_GROUP` and `LDAP_VIEWER_GROUP` DNs are correct
   - Check `LDAP_GROUP_FILTER` matches your LDAP group schema

5. **"Failed to check group membership"**
   - Service account (`LDAP_BIND_DN`) may not have permission to read groups
   - `LDAP_GROUP_FILTER` may be incorrect for your LDAP schema
   - Group DNs may be incorrect

### Testing LDAP Configuration

You can test your LDAP configuration using `ldapsearch`:

```bash
# Test user search
ldapsearch -H ldap://ldap.example.com:389 \
  -D "cn=admin,dc=example,dc=com" \
  -w "password" \
  -b "ou=users,dc=example,dc=com" \
  "(uid=username)"

# Test group membership
ldapsearch -H ldap://ldap.example.com:389 \
  -D "cn=admin,dc=example,dc=com" \
  -w "password" \
  -b "cn=admins,ou=groups,dc=example,dc=com" \
  "(member=uid=username,ou=users,dc=example,dc=com)"
```

## User Management

### User Creation

- Users are automatically created in the database upon first successful LDAP login
- User information (email, name) is synced from LDAP attributes
- Admin status is determined by group membership

### User Updates

- User information is updated on each login
- Admin status is re-evaluated on each login based on current group membership
- Users removed from all groups will be denied access on next login

### User Deletion

- Users can be disabled in the admin panel
- Disabled users cannot log in
- Admin users can also delete user accounts from the database

## Migration from OIDC

If you're currently using OIDC and want to add LDAP:

1. LDAP users will be created with a `ldap:` prefix in the `sub` field
2. OIDC users will continue to work with their existing OIDC `sub` values
3. The same user logging in via OIDC and LDAP will create two separate accounts
4. Consider which authentication method to keep enabled based on your organization's needs

## API Access

LDAP authentication is for web interface access only. API access continues to use API keys:

- Admin users (via LDAP or OIDC) can create API keys in the admin panel
- API keys are used for `/v1/*` endpoints
- Session cookies from LDAP login are not valid for API access

## Logging

LDAP authentication events are logged:

- Successful logins: `INFO` level
- Authentication failures: `ERROR` level
- Group membership checks: `ERROR` level (if failed)

Check application logs for troubleshooting authentication issues.

## Docker Configuration

When running in Docker, pass LDAP variables as environment variables:

```bash
docker run -d -p 8080:8080 \
  -e LDAP_HOST=ldap.example.com \
  -e LDAP_BASE_DN=ou=users,dc=example,dc=com \
  -e LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com \
  -e LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com \
  cr.rda.run/txlog/server:main
```

Or use Docker Compose:

```yaml
version: '3.8'
services:
  txlog-server:
    image: cr.rda.run/txlog/server:main
    ports:
      - "8080:8080"
    environment:
      - LDAP_HOST=ldap.example.com
      - LDAP_BASE_DN=ou=users,dc=example,dc=com
      - LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
      - LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
```

## Support

For issues or questions about LDAP authentication:

1. Check this documentation
2. Review application logs
3. Test LDAP connectivity with `ldapsearch`
4. Open an issue at <https://github.com/txlog/server/issues>
