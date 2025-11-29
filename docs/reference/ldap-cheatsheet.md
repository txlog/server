# LDAP Authentication - Quick Reference

## Minimal Configuration WITHOUT Service Account

```bash
# Simplest setup - works if LDAP allows anonymous search
LDAP_HOST=ldap.example.com
LDAP_BASE_DN=ou=users,dc=example,dc=com
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
```

## Minimal Configuration WITH Service Account

```bash
# For restricted LDAP servers (like Active Directory)
LDAP_HOST=ldap.example.com
LDAP_BASE_DN=ou=users,dc=example,dc=com
LDAP_BIND_DN=cn=readonly,dc=example,dc=com
LDAP_BIND_PASSWORD=your_password
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
```

## Full Configuration (All Options)

```bash
# Connection
LDAP_HOST=ldap.example.com              # Required
LDAP_PORT=389                           # Default: 389 (636 for LDAPS)
LDAP_USE_TLS=false                      # Default: false
LDAP_SKIP_TLS_VERIFY=false              # Default: false

# Service Account (Optional but recommended)
LDAP_BIND_DN=cn=admin,dc=example,dc=com
LDAP_BIND_PASSWORD=your_password

# User Search
LDAP_BASE_DN=ou=users,dc=example,dc=com # Required
LDAP_USER_FILTER=(uid=%s)               # Default: (uid=%s)

# Authorization Groups (at least one required)
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
LDAP_GROUP_FILTER=(member=%s)           # Default: (member=%s)
```

## Common LDAP Server Configurations

### Active Directory

```bash
LDAP_USER_FILTER=(sAMAccountName=%s)
LDAP_GROUP_FILTER=(member=%s)
```

### OpenLDAP (standard)

```bash
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(member=%s)
```

### OpenLDAP (posixGroup)

```bash
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(memberUid=%s)
```

## Environment Variable Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `LDAP_HOST` | Yes | - | LDAP server hostname |
| `LDAP_PORT` | No | 389/636 | LDAP server port |
| `LDAP_USE_TLS` | No | false | Enable TLS connection |
| `LDAP_SKIP_TLS_VERIFY` | No | false | Skip TLS cert verification |
| `LDAP_BIND_DN` | No | - | Service account DN (optional) |
| `LDAP_BIND_PASSWORD` | No | - | Service account password (optional) |
| `LDAP_BASE_DN` | Yes | - | Base DN for user searches |
| `LDAP_USER_FILTER` | No | (uid=%s) | User search filter |
| `LDAP_ADMIN_GROUP` | Yes* | - | Admin group DN |
| `LDAP_VIEWER_GROUP` | Yes* | - | Viewer group DN |
| `LDAP_GROUP_FILTER` | No | (member=%s) | Group membership filter |

\* At least one of `LDAP_ADMIN_GROUP` or `LDAP_VIEWER_GROUP` is required

**Note about Service Account**: `LDAP_BIND_DN` and `LDAP_BIND_PASSWORD` are **optional**. If not provided:

- Anonymous bind will be used for user searches
- Group membership checks will use the authenticated user's session
- Works with OpenLDAP and other servers that allow anonymous searches
- **Active Directory typically requires a service account**

## User Roles

### Admin Group Members

- View all data
- Manage assets and packages
- Access admin panel
- Create API keys
- Manage users

### Viewer Group Members

- View transaction data
- View assets and packages
- View insights
- **No admin access**

## Testing LDAP Connection

```bash
# Test with ldapsearch
ldapsearch -H ldap://ldap.example.com:389 \
  -D "cn=admin,dc=example,dc=com" \
  -w "password" \
  -b "ou=users,dc=example,dc=com" \
  "(uid=testuser)"
```

## Troubleshooting Quick Checks

1. **Can't connect**: Check `LDAP_HOST` and `LDAP_PORT`, verify network access
2. **User not found**: Verify `LDAP_BASE_DN` and `LDAP_USER_FILTER`
3. **Invalid credentials**: Check username/password in LDAP
4. **Not authorized**: User not in `LDAP_ADMIN_GROUP` or `LDAP_VIEWER_GROUP`
5. **Group check fails**: Verify `LDAP_BIND_DN` has read access to groups

## Authentication Flow

```text
WITHOUT Service Account:
1. User enters username/password
2. Server connects to LDAP
3. Server searches for user (anonymous bind)
4. Server authenticates via LDAP bind with user credentials
5. Server checks group membership (using authenticated session)
6. Server creates/updates user in database
7. Server creates session cookie
8. User redirected to dashboard

WITH Service Account:
1. User enters username/password
2. Server connects to LDAP
3. Server binds with service account
4. Server searches for user
5. Server authenticates user via LDAP bind
6. Server re-binds with service account
7. Server checks group membership
8. Server creates/updates user in database
9. Server creates session cookie
10. User redirected to dashboard
```

## Docker Quick Start

### Without Service Account (OpenLDAP with anonymous search)

```bash
docker run -d -p 8080:8080 \
  -e LDAP_HOST=ldap.example.com \
  -e LDAP_BASE_DN=ou=users,dc=example,dc=com \
  -e LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com \
  -e LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com \
  cr.rda.run/txlog/server:main
```

### With Service Account (Active Directory or restricted LDAP)

```bash
docker run -d -p 8080:8080 \
  -e LDAP_HOST=ldap.example.com \
  -e LDAP_BASE_DN=ou=users,dc=example,dc=com \
  -e LDAP_BIND_DN=cn=readonly,dc=example,dc=com \
  -e LDAP_BIND_PASSWORD=your_password \
  -e LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com \
  -e LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com \
  cr.rda.run/txlog/server:main
```

## Documentation Links

- Full Guide: [LDAP_AUTHENTICATION.md](LDAP_AUTHENTICATION.md)
- Implementation Details: [LDAP_IMPLEMENTATION_SUMMARY.md](LDAP_IMPLEMENTATION_SUMMARY.md)
- General Setup: [README.md](README.md)
