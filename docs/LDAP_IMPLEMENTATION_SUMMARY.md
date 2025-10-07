# LDAP Authentication Implementation Summary

## Overview

LDAP authentication has been successfully added to Txlog Server. The implementation provides username/password
authentication with group-based authorization, working alongside the existing OIDC authentication system.

## Changes Made

### 1. New Files Created

#### `auth/ldap.go`

- Complete LDAP authentication service implementation
- Features:
  - Connection handling (LDAP and LDAPS)
  - User authentication via bind
  - Group membership checking
  - User creation/update in database
  - Session management
  - Configurable via environment variables

#### `LDAP_AUTHENTICATION.md`

- Comprehensive documentation for LDAP setup
- Configuration examples for various LDAP servers
- Troubleshooting guide
- Security best practices

### 2. Files Modified

#### `controllers/auth_controller.go`

- Updated `GetLogin()` to accept both OIDC and LDAP services
- Added `PostLDAPLogin()` handler for LDAP authentication
- Updated `PostLogout()` to support both authentication methods
- Login page now shows which authentication methods are available

#### `middleware/auth.go`

- Updated `AuthMiddleware()` to check both OIDC and LDAP configuration
- Updated `AdminMiddleware()` to check both authentication methods
- Authentication is bypassed if neither OIDC nor LDAP is configured

#### `main.go`

- Initialize LDAP service alongside OIDC service
- Register LDAP authentication routes
- Add LDAP environment variables to template context
- Log authentication status for both methods

#### `templates/login.html`

- Complete redesign to support multiple authentication methods
- Shows LDAP login form with username/password fields
- Shows OIDC login button
- Allows user to choose between methods when both are enabled
- JavaScript to toggle between authentication forms
- Improved error messages for LDAP authentication

#### `README.md`

- Added LDAP configuration section
- Documented all LDAP environment variables
- Added authentication mode explanation
- Included LDAP configuration examples

#### `go.mod` / `go.sum`

- Added dependency: `github.com/go-ldap/ldap/v3 v3.4.12`

### 3. Environment Variables Added

#### Required for LDAP

- `LDAP_HOST`: LDAP server hostname
- `LDAP_BASE_DN`: Base DN for user searches
- `LDAP_ADMIN_GROUP` or `LDAP_VIEWER_GROUP`: At least one must be configured

#### Optional LDAP Variables

- `LDAP_PORT`: Server port (default: 389 or 636)
- `LDAP_USE_TLS`: Enable TLS (default: false)
- `LDAP_SKIP_TLS_VERIFY`: Skip TLS verification (default: false)
- `LDAP_BIND_DN`: Service account DN
- `LDAP_BIND_PASSWORD`: Service account password
- `LDAP_USER_FILTER`: User search filter (default: `(uid=%s)`)
- `LDAP_GROUP_FILTER`: Group membership filter (default: `(member=%s)`)

## Features Implemented

### 1. Authentication Modes

The server now supports three authentication modes:

- **No Authentication**: Default when neither OIDC nor LDAP is configured
- **OIDC Only**: When only OIDC variables are set
- **LDAP Only**: When only LDAP variables are set
- **Both OIDC and LDAP**: Users can choose their preferred method

### 2. Group-Based Authorization

- **Admin Group**: Full access to all features including admin panel
- **Viewer Group**: Read-only access to data
- Users must be in at least one group to authenticate
- Admin group members have full privileges regardless of viewer group membership

### 3. User Management

- Users are automatically created on first LDAP login
- User information is synced from LDAP attributes (email, name)
- Admin status is determined by group membership
- Users are updated on each login
- Session management (7-day cookies)

### 4. Security Features

- TLS/LDAPS support
- Service account for group lookups
- Password verification via LDAP bind
- Configurable search filters
- Self-signed certificate support for development

### 5. Compatibility

- Works alongside existing OIDC authentication
- Same user database schema (uses `ldap:` prefix in `sub` field)
- Same session management system
- API keys continue to work independently

## Technical Details

### User Identification

- LDAP users are stored with `sub` field as `ldap:username`
- OIDC users keep their OIDC `sub` identifier
- This prevents conflicts between authentication methods

### Database Schema

- No database migration required
- Reuses existing `users` and `user_sessions` tables
- Compatible with existing OIDC user data

### Authentication Flow

1. User enters username and password on login page
2. Server connects to LDAP server
3. Binds with service account (if configured)
4. Searches for user in base DN
5. Authenticates user via LDAP bind with their credentials
6. Checks group membership (admin and/or viewer)
7. Creates/updates user in database
8. Creates session and sets cookie
9. Redirects to dashboard

### LDAP Attribute Mapping

- `uid` or `sAMAccountName` → Username (login identifier)
- `mail` → Email (fallback: `username@local`)
- `cn` or `displayName` → Display Name (fallback: username)

### Group Membership Check

- Supports standard `member` attribute (Active Directory, OpenLDAP)
- Supports `memberUid` attribute (posixGroup)
- Configurable via `LDAP_GROUP_FILTER`

## Testing Results

### Build and Tests

- ✅ Code formatted with `make fmt`
- ✅ Static analysis passed with `make vet`
- ✅ Production build successful with `make build`
- ✅ All existing tests pass
- ✅ No breaking changes to existing functionality

### Compatibility

- ✅ Works with OIDC enabled
- ✅ Works with LDAP enabled
- ✅ Works with both enabled
- ✅ Works with neither enabled (no auth)
- ✅ Existing OIDC users not affected

## Usage Examples

### Example 1: Enable LDAP with Active Directory

```bash
LDAP_HOST=dc01.domain.local
LDAP_PORT=389
LDAP_BASE_DN=cn=Users,dc=domain,dc=local
LDAP_USER_FILTER=(sAMAccountName=%s)
LDAP_BIND_DN=cn=ServiceAccount,cn=Users,dc=domain,dc=local
LDAP_BIND_PASSWORD=YourPassword
LDAP_ADMIN_GROUP=cn=TxlogAdmins,ou=Groups,dc=domain,dc=local
LDAP_VIEWER_GROUP=cn=TxlogViewers,ou=Groups,dc=domain,dc=local
LDAP_GROUP_FILTER=(member=%s)
```

### Example 2: Enable LDAP with OpenLDAP (LDAPS)

```bash
LDAP_HOST=ldap.example.com
LDAP_PORT=636
LDAP_USE_TLS=true
LDAP_BASE_DN=ou=people,dc=example,dc=com
LDAP_USER_FILTER=(uid=%s)
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=users,ou=groups,dc=example,dc=com
```

### Example 3: Both OIDC and LDAP Enabled

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

## Migration Path

### For New Deployments

- Choose OIDC, LDAP, or both based on infrastructure
- Configure appropriate environment variables
- Users will be created on first login

### For Existing OIDC Deployments

- Add LDAP configuration alongside OIDC
- Existing OIDC users continue to work
- New LDAP users can be added
- Users can use either method

### For Deployments Without Authentication

- Add LDAP configuration
- First LDAP user from admin group becomes admin
- All subsequent users follow group membership rules

## Documentation

### User-Facing Documentation

- `README.md`: Updated with LDAP configuration section
- `LDAP_AUTHENTICATION.md`: Complete LDAP setup and troubleshooting guide

### Developer Documentation

- Code comments in `auth/ldap.go` explain all functions
- Environment variable documentation in comments
- Examples for various LDAP server types

## Known Limitations

1. **Nested Groups**: Only direct group membership is checked
2. **Password Policies**: Enforced by LDAP server, not the application
3. **Account Lockout**: Handled by LDAP server
4. **User Provisioning**: Users must exist in LDAP before login
5. **Profile Pictures**: LDAP doesn't provide profile pictures (blank for LDAP users)

## Future Enhancements (Not Implemented)

These could be added in future updates:

- Nested group membership support
- LDAP connection pooling
- Account lockout tracking in application
- Failed login attempt rate limiting
- LDAP group caching for performance
- Support for multiple LDAP servers (failover)

## Deployment Checklist

- [ ] Configure LDAP environment variables
- [ ] Test LDAP connectivity with `ldapsearch`
- [ ] Create admin and viewer groups in LDAP
- [ ] Add users to appropriate groups
- [ ] Test LDAP authentication with a test user
- [ ] Verify admin privileges work correctly
- [ ] Verify viewer privileges work correctly
- [ ] Test session management (cookie expiry)
- [ ] Review application logs for authentication events

## Support

For issues or questions:

1. Review `LDAP_AUTHENTICATION.md` documentation
2. Check application logs for error messages
3. Test LDAP connectivity independently
4. Open an issue at <https://github.com/txlog/server/issues>
