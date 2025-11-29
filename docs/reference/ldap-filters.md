# Quick Guide: LDAP Filters

## TL;DR - Discovering Your Filters

### Method 1: Automatic Script (Recommended)

```bash
./ldap-discovery.sh
```

### Method 2: Manual with ldapsearch

**1. Find users:**

```bash
ldapsearch -H ldap://your-server:389 -x -D "cn=admin,dc=example,dc=com" -W -b "dc=example,dc=com" "(uid=username)"
```

**2. View user attributes:**

```bash
ldapsearch -H ldap://your-server:389 -x -D "cn=admin,dc=example,dc=com" -W -b "dc=example,dc=com" "(uid=john)" dn uid cn sAMAccountName
```

**3. View groups:**

```bash
ldapsearch -H ldap://your-server:389 -x -D "cn=admin,dc=example,dc=com" -W -b "dc=example,dc=com" "(cn=admins)" dn member uniqueMember memberUid
```

---

## Common Values by Server Type

### OpenLDAP (Standard)

```bash
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(member=%s)
```

### Active Directory

```bash
LDAP_USER_FILTER=(sAMAccountName=%s)
LDAP_GROUP_FILTER=(member=%s)
```

### FreeIPA

```bash
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(member=%s)
```

### OpenLDAP with posixGroup

```bash
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(memberUid=%s)
```

⚠️ **Attention:** posixGroup uses only the `uid` (e.g., `john`) and not the full DN.

---

## Quick Reference Table

### USER_FILTER - By System

| System | Attribute | Filter |
|--------|-----------|--------|
| OpenLDAP | `uid` | `(uid=%s)` |
| Active Directory | `sAMAccountName` | `(sAMAccountName=%s)` |
| AD (email login) | `userPrincipalName` | `(userPrincipalName=%s)` |
| FreeIPA | `uid` | `(uid=%s)` |
| Legacy | `cn` | `(cn=%s)` |
| Email login | `mail` | `(mail=%s)` |

### GROUP_FILTER - By Group Type

| ObjectClass | Member Attribute | Filter | Expected Value |
|-------------|------------------|--------|----------------|
| `groupOfNames` | `member` | `(member=%s)` | Full DN |
| `groupOfUniqueNames` | `uniqueMember` | `(uniqueMember=%s)` | Full DN |
| `posixGroup` | `memberUid` | `(memberUid=%s)` | uid only |
| `group` (AD) | `member` | `(member=%s)` | Full DN |

---

## How to Know Which One to Use?

### Step 1: Identify User Login Attribute

Search for a user and see which field contains the login name:

```bash
ldapsearch -x -D "cn=admin,dc=example,dc=com" -W -b "dc=example,dc=com" "(objectClass=person)" uid cn sAMAccountName
```

OpenLDAP output example:

```text
dn: uid=john.doe,ou=users,dc=example,dc=com
uid: john.doe          ← This is the login field!
cn: John Doe
```

Active Directory output example:

```text
dn: CN=John Doe,CN=Users,DC=example,DC=com
sAMAccountName: john.doe    ← This is the login field!
cn: John Doe
```

**Result:** Use the attribute name in the filter → `(uid=%s)` or `(sAMAccountName=%s)`

---

### Step 2: Identify Group Member Attribute

Search for a group and see which field lists the members:

```bash
ldapsearch -x -D "cn=admin,dc=example,dc=com" -W -b "dc=example,dc=com" "(cn=admins)" member uniqueMember memberUid
```

Example with `member`:

```text
dn: cn=admins,ou=groups,dc=example,dc=com
member: uid=john.doe,ou=users,dc=example,dc=com    ← Full DN
member: uid=jane.doe,ou=users,dc=example,dc=com
```

**Result:** `LDAP_GROUP_FILTER=(member=%s)`

Example with `memberUid`:

```text
dn: cn=admins,ou=groups,dc=example,dc=com
memberUid: john.doe    ← Just the uid, no DN
memberUid: jane.doe
```

**Result:** `LDAP_GROUP_FILTER=(memberUid=%s)`
⚠️ **Requires code modification to extract only the uid from the DN**

---

## Testing Before Configuring

### Test 1: Can the user be found?

```bash
# Replace %s with real username
ldapsearch -x -D "cn=admin,dc=example,dc=com" -W -b "dc=example,dc=com" "(uid=john.doe)"
```

✅ Should return **exactly 1 user**

### Test 2: Does the user belong to the group?

```bash
# Replace group DN and user DN
ldapsearch -x -D "cn=admin,dc=example,dc=com" -W \
  -b "cn=admins,ou=groups,dc=example,dc=com" \
  -s base \
  "(member=uid=john.doe,ou=users,dc=example,dc=com)"
```

✅ Should return the group if the user is a member

---

## Complete .env Example

```bash
# OpenLDAP
LDAP_HOST=ldap.company.com
LDAP_PORT=389
LDAP_USE_TLS=false
LDAP_BASE_DN=dc=company,dc=com
LDAP_BIND_DN=cn=readonly,dc=company,dc=com
LDAP_BIND_PASSWORD=readonly_password

LDAP_USER_FILTER=(uid=%s)
LDAP_ADMIN_GROUP=cn=txlog-admins,ou=groups,dc=company,dc=com
LDAP_VIEWER_GROUP=cn=txlog-viewers,ou=groups,dc=company,dc=com
LDAP_GROUP_FILTER=(member=%s)
```

```bash
# Active Directory
LDAP_HOST=ad.company.com
LDAP_PORT=636
LDAP_USE_TLS=true
LDAP_SKIP_TLS_VERIFY=false
LDAP_BASE_DN=DC=company,DC=com
LDAP_BIND_DN=CN=LDAP Service,OU=Service Accounts,DC=company,DC=com
LDAP_BIND_PASSWORD=service_account_password

LDAP_USER_FILTER=(sAMAccountName=%s)
LDAP_ADMIN_GROUP=CN=Txlog Admins,OU=Security Groups,DC=company,DC=com
LDAP_VIEWER_GROUP=CN=Txlog Users,OU=Security Groups,DC=company,DC=com
LDAP_GROUP_FILTER=(member=%s)
```

---

## Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| "user not found" | Wrong LDAP_USER_FILTER | Use `ldapsearch` to test the filter |
| "not a member of any authorized group" | Wrong LDAP_GROUP_FILTER or incorrect group | Check if user is in group and test filter |
| "failed to bind" | Incorrect LDAP_BIND_DN or password | Test bind manually |
| "connection refused" | Incorrect Host/Port or firewall | Check connectivity with `telnet` |

---

## Resources

- **Full Document:** `LDAP_FILTER_DISCOVERY.md`
- **Interactive Script:** `./ldap-discovery.sh`
- **Official LDAP Documentation:** <https://ldap.com/ldap-filters/>

---

## Final Tip

**Use the `ldap-discovery.sh` script** - it guides you step-by-step to discover all necessary values interactively!

```bash
chmod +x ldap-discovery.sh
./ldap-discovery.sh
```
