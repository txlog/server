# LDAP Configuration Without Service Account - Practical Guide

## When to Use

Authentication **WITHOUT service account** works when:

✅ Your LDAP server allows anonymous bind for searches.
✅ Authenticated users can read their own groups.
✅ You are using OpenLDAP with default configuration.
✅ You want a simpler configuration with fewer credentials.

## When NOT to Use (Needs Service Account)

❌ Active Directory (usually requires authentication for searches).
❌ LDAP with restricted ACLs that block anonymous bind.
❌ Production environments with strict security policies.
❌ LDAP that does not allow users to read their own groups.

## Example 1: Basic OpenLDAP (No Service Account)

```bash
# Minimal configuration - only 4 variables!
LDAP_HOST=ldap.mycompany.com
LDAP_BASE_DN=ou=users,dc=mycompany,dc=com
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=mycompany,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=mycompany,dc=com
```

### How It Works

1. **User Search**: Anonymous bind → search user by `uid`.
2. **Authentication**: Bind with the user's own credentials.
3. **Group Check**: Uses the user's authenticated session to read groups.

## Example 2: OpenLDAP with TLS (No Service Account)

```bash
LDAP_HOST=ldap.mycompany.com
LDAP_PORT=636
LDAP_USE_TLS=true
LDAP_BASE_DN=ou=people,dc=mycompany,dc=com
LDAP_USER_FILTER=(uid=%s)
LDAP_ADMIN_GROUP=cn=txlog-admins,ou=groups,dc=mycompany,dc=com
LDAP_VIEWER_GROUP=cn=txlog-users,ou=groups,dc=mycompany,dc=com
```

## Example 3: Testing Without Service Account

### Test 1: Verify if anonymous bind works

```bash
# Try to search without authentication
ldapsearch -H ldap://ldap.mycompany.com:389 \
  -x \
  -b "ou=users,dc=mycompany,dc=com" \
  "(uid=myuser)"
```

**If it works**: ✅ Can use without service account.
**If fails with "No such object" or "Insufficient access"**: ❌ Needs service account.

### Test 2: Verify group reading

```bash
# Authenticate as user and try to read group
ldapsearch -H ldap://ldap.mycompany.com:389 \
  -D "uid=myuser,ou=users,dc=mycompany,dc=com" \
  -w "mypassword" \
  -b "cn=admins,ou=groups,dc=mycompany,dc=com" \
  "(member=uid=myuser,ou=users,dc=mycompany,dc=com)"
```

**If returns the group**: ✅ Group verification will work.
**If fails**: ❌ Needs service account with read permissions.

## OpenLDAP Configuration to Allow Anonymous Bind

If you manage the OpenLDAP server, configure to allow anonymous reads:

```ldif
# /etc/ldap/slapd.conf or via olcAccess

# Allow anonymous read for users and groups
olcAccess: {0}to dn.subtree="ou=users,dc=mycompany,dc=com"
  by anonymous read
  by * read

olcAccess: {1}to dn.subtree="ou=groups,dc=mycompany,dc=com"
  by anonymous read
  by * read
```

## Comparison: With vs Without Service Account

### WITHOUT Service Account (Simpler)

**Pros:**

- ✅ Simpler configuration (fewer variables).
- ✅ No need to create service account.
- ✅ Fewer credentials to manage.
- ✅ Works well with standard OpenLDAP.

**Cons:**

- ❌ Does not work with Active Directory (usually).
- ❌ Requires anonymous bind enabled.
- ❌ May not meet security policies.
- ❌ User needs permission to read groups.

**Configuration:**

```bash
LDAP_HOST=ldap.example.com
LDAP_BASE_DN=ou=users,dc=example,dc=com
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
```

### WITH Service Account (More Robust)

**Pros:**

- ✅ Works with Active Directory.
- ✅ Works with restrictive LDAP.
- ✅ More control over permissions.
- ✅ Better for production.

**Cons:**

- ❌ More configuration variables.
- ❌ Need to create and manage service account.
- ❌ One more password to store securely.

**Configuration:**

```bash
LDAP_HOST=ldap.example.com
LDAP_BIND_DN=cn=readonly,dc=example,dc=com
LDAP_BIND_PASSWORD=service_account_password
LDAP_BASE_DN=ou=users,dc=example,dc=com
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
```

## Detailed Authentication Flow

### Without Service Account

```text
1. Client sends username + password
   ↓
2. Server connects to LDAP (anonymous)
   ↓
3. Searches user: ldapsearch -x "(uid=username)"
   ↓
4. Finds: uid=username,ou=users,dc=example,dc=com
   ↓
5. Authenticates: bind with uid=username + user password
   ↓
6. Checks groups using user's authenticated session
   ↓
7. Creates session in Txlog Server
   ↓
8. User logged in!
```

### With Service Account

```text
1. Client sends username + password
   ↓
2. Server connects to LDAP
   ↓
3. Bind with service account
   ↓
4. Searches user: ldapsearch "(uid=username)"
   ↓
5. Finds: uid=username,ou=users,dc=example,dc=com
   ↓
6. Authenticates: bind with uid=username + user password
   ↓
7. Re-bind with service account
   ↓
8. Checks groups using service account
   ↓
9. Creates session in Txlog Server
   ↓
10. User logged in!
```

## Troubleshooting

### Error: "User not found"

**Without service account:**

```bash
# Test anonymous search
ldapsearch -H ldap://your-ldap:389 -x \
  -b "ou=users,dc=example,dc=com" \
  "(uid=testuser)"
```

**Solution:** If fails, you need a service account.

### Error: "Failed to check group membership"

**Without service account:**

```bash
# Test if user can read groups
ldapsearch -H ldap://your-ldap:389 \
  -D "uid=testuser,ou=users,dc=example,dc=com" \
  -w "password" \
  -b "cn=admins,ou=groups,dc=example,dc=com"
```

**Solution:** If fails, configure service account with read permission on groups.

### Error: "Failed to connect to LDAP"

Same problem with or without service account - check:

- Correct host and port.
- Firewall allowed.
- LDAP server running.

## Recommendations

### Development/Test

✅ **Use WITHOUT service account** if possible.

- Faster to configure.
- Less complex.
- Local OpenLDAP usually allows it.

### Production

✅ **Use WITH service account**.

- More secure.
- More control.
- Works with Active Directory.
- Meets security policies.

### Mixed Environments

✅ **Start WITHOUT service account**.

- Test basic connectivity.
- If it works, decide if you will add service account.
- If it doesn't work, add service account.

## Complete Example: Docker Compose

### Without Service Account (OpenLDAP)

```yaml
version: '3.8'
services:
  txlog-server:
    image: ghcr.io/txlog/server:main
    ports:
      - "8080:8080"
    environment:
      # Database
      - PGSQL_HOST=postgres
      - PGSQL_DB=txlog
      - PGSQL_USER=txlog
      - PGSQL_PASSWORD=txlog_password

      # LDAP - WITHOUT service account
      - LDAP_HOST=openldap
      - LDAP_BASE_DN=ou=users,dc=example,dc=com
      - LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
      - LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
```

### With Service Account (Active Directory)

```yaml
version: '3.8'
services:
  txlog-server:
    image: ghcr.io/txlog/server:main
    ports:
      - "8080:8080"
    environment:
      # Database
      - PGSQL_HOST=postgres
      - PGSQL_DB=txlog
      - PGSQL_USER=txlog
      - PGSQL_PASSWORD=txlog_password

      # LDAP - WITH service account
      - LDAP_HOST=ad.company.local
      - LDAP_BIND_DN=CN=SvcTxlog,OU=ServiceAccounts,DC=company,DC=local
      - LDAP_BIND_PASSWORD=service_password
      - LDAP_BASE_DN=CN=Users,DC=company,DC=local
      - LDAP_USER_FILTER=(sAMAccountName=%s)
      - LDAP_ADMIN_GROUP=CN=TxlogAdmins,OU=Groups,DC=company,DC=local
      - LDAP_VIEWER_GROUP=CN=TxlogUsers,OU=Groups,DC=company,DC=local
```

## Conclusion

Authentication **WITHOUT service account** is:

- ✅ **Simpler** to configure.
- ✅ **Perfectly functional** for OpenLDAP.
- ✅ **Ideal for development** and less restrictive environments.
- ❌ **Does not work** with typical Active Directory.
- ❌ **May not meet** corporate security policies.

**Recommendation:** Start without service account. If it works and meets your security needs, great! If it doesn't work or if you need more control, add the service account.
