# Discovering LDAP Filters for Your Server

This guide helps you discover the correct values for `LDAP_USER_FILTER` and `LDAP_GROUP_FILTER` in your specific LDAP environment.

## Index

- [Required Tools](#required-tools)
- [Step 1: Connect to LDAP Server](#step-1-connect-to-ldap-server)
- [Step 2: Explore Structure](#step-2-explore-structure)
- [Step 3: Find Users](#step-3-find-users)
- [Step 4: Find Groups](#step-4-find-groups)
- [Step 5: Determine Filters](#step-5-determine-filters)
- [Common Examples](#common-examples)

---

## Required Tools

### Linux/Mac

```bash
# Install ldap-utils (Debian/Ubuntu)
sudo apt-get install ldap-utils

# Install ldap-utils (Red Hat/CentOS/AlmaLinux)
sudo yum install openldap-clients

# Install ldap-utils (Mac)
brew install openldap
```

### Windows

- Download and install **Apache Directory Studio** (GUI): <https://directory.apache.org/studio/>
- Or use **ldp.exe** (already included in Windows Server)

---

## Step 1: Connect to LDAP Server

### Using ldapsearch (Command Line)

```bash
# Basic connection test (no TLS)
ldapsearch -H ldap://your-ldap-server.com:389 \
  -x \
  -b "dc=example,dc=com" \
  -D "cn=admin,dc=example,dc=com" \
  -W \
  "(objectClass=*)" dn

# With TLS/LDAPS
ldapsearch -H ldaps://your-ldap-server.com:636 \
  -x \
  -b "dc=example,dc=com" \
  -D "cn=admin,dc=example,dc=com" \
  -W \
  "(objectClass=*)" dn
```

**Parameters:**

- `-H`: LDAP server URL
- `-x`: Simple authentication
- `-b`: Base DN (search starting point)
- `-D`: Bind DN (user for authentication)
- `-W`: Prompt for password interactively
- `-w password`: Password on command line (not recommended)

### Using Apache Directory Studio (GUI)

1. Open Apache Directory Studio
2. Click **"New Connection"**
3. Configure:
   - **Connection name**: Descriptive name
   - **Hostname**: LDAP server address
   - **Port**: 389 (LDAP) or 636 (LDAPS)
   - **Encryption**: None/LDAPS/StartTLS
4. On **"Authentication"** tab:
   - **Authentication Method**: Simple Authentication
   - **Bind DN**: cn=admin,dc=example,dc=com
   - **Bind Password**: your password
5. Click **"Check Network Parameter"** to test
6. Click **"Finish"**

---

## Step 2: Explore Structure

### Discover Base Structure

```bash
# List all entries at root level
ldapsearch -H ldap://your-ldap-server.com:389 \
  -x \
  -b "dc=example,dc=com" \
  -D "cn=admin,dc=example,dc=com" \
  -W \
  -s one \
  "(objectClass=*)" dn

# View complete structure (use with caution on large directories)
ldapsearch -H ldap://your-ldap-server.com:389 \
  -x \
  -b "dc=example,dc=com" \
  -D "cn=admin,dc=example,dc=com" \
  -W \
  -LLL \
  "(objectClass=organizationalUnit)" dn
```

**Common structures:**

```text
dc=example,dc=com
├── ou=users          ← Users usually here
├── ou=people         ← Or here
├── ou=grupos         ← Groups usually here
├── ou=groups         ← Or here
└── ou=departments    ← Organizational structure
```

---

## Step 3: Find Users

### 3.1 Search All Users

```bash
# Search by person
ldapsearch -H ldap://your-ldap-server.com:389 \
  -x \
  -b "dc=example,dc=com" \
  -D "cn=admin,dc=example,dc=com" \
  -W \
  -LLL \
  "(objectClass=person)"

# Search by inetOrgPerson (most common)
ldapsearch -H ldap://your-ldap-server.com:389 \
  -x \
  -b "dc=example,dc=com" \
  -D "cn=admin,dc=example,dc=com" \
  -W \
  -LLL \
  "(objectClass=inetOrgPerson)"

# Search by posixAccount (Unix/Linux systems)
ldapsearch -H ldap://your-ldap-server.com:389 \
  -x \
  -b "dc=example,dc=com" \
  -D "cn=admin,dc=example,dc=com" \
  -W \
  -LLL \
  "(objectClass=posixAccount)"
```

### 3.2 Examine a Specific User

```bash
# Search user by uid
ldapsearch -H ldap://your-ldap-server.com:389 \
  -x \
  -b "dc=example,dc=com" \
  -D "cn=admin,dc=example,dc=com" \
  -W \
  -LLL \
  "(uid=john.doe)"

# Search user by cn (common name)
ldapsearch -H ldap://your-ldap-server.com:389 \
  -x \
  -b "dc=example,dc=com" \
  -D "cn=admin,dc=example,dc=com" \
  -W \
  -LLL \
  "(cn=John Doe)"

# Search user by sAMAccountName (Active Directory)
ldapsearch -H ldap://your-ldap-server.com:389 \
  -x \
  -b "dc=example,dc=com" \
  -D "cn=admin,dc=example,dc=com" \
  -W \
  -LLL \
  "(sAMAccountName=john.doe)"
```

### 3.3 Identify Login Attribute

Examine the output and look for:

```ldif
dn: uid=john.doe,ou=users,dc=example,dc=com
objectClass: inetOrgPerson
objectClass: posixAccount
uid: john.doe              ← Login attribute (common in OpenLDAP)
cn: John Doe
mail: john.doe@example.com
```

or

```ldif
dn: CN=John Doe,CN=Users,DC=example,DC=com
objectClass: user
sAMAccountName: john.doe   ← Login attribute (Active Directory)
cn: John Doe
userPrincipalName: john.doe@example.com
mail: john.doe@example.com
```

**Common login attributes:**

- `uid`: OpenLDAP, FreeIPA, 389 Directory Server
- `sAMAccountName`: Active Directory
- `cn`: Some older systems
- `mail`: Some systems use email as login

---

## Step 4: Find Groups

### 4.1 Search All Groups

```bash
# Search by groupOfNames (Standard LDAP)
ldapsearch -H ldap://your-ldap-server.com:389 \
  -x \
  -b "dc=example,dc=com" \
  -D "cn=admin,dc=example,dc=com" \
  -W \
  -LLL \
  "(objectClass=groupOfNames)"

# Search by groupOfUniqueNames
ldapsearch -H ldap://your-ldap-server.com:389 \
  -x \
  -b "dc=example,dc=com" \
  -D "cn=admin,dc=example,dc=com" \
  -W \
  -LLL \
  "(objectClass=groupOfUniqueNames)"

# Search by posixGroup (Unix/Linux systems)
ldapsearch -H ldap://your-ldap-server.com:389 \
  -x \
  -b "dc=example,dc=com" \
  -D "cn=admin,dc=example,dc=com" \
  -W \
  -LLL \
  "(objectClass=posixGroup)"

# Search by group (Active Directory)
ldapsearch -H ldap://your-ldap-server.com:389 \
  -x \
  -b "dc=example,dc=com" \
  -D "cn=admin,dc=example,dc=com" \
  -W \
  -LLL \
  "(objectClass=group)"
```

### 4.2 Examine a Specific Group

```bash
# Search specific group
ldapsearch -H ldap://your-ldap-server.com:389 \
  -x \
  -b "dc=example,dc=com" \
  -D "cn=admin,dc=example,dc=com" \
  -W \
  -LLL \
  "(cn=admins)"
```

### 4.3 Identify Member Attribute

Examine the output and look for:

#### Type 1: groupOfNames (OpenLDAP)

```ldif
dn: cn=admins,ou=groups,dc=example,dc=com
objectClass: groupOfNames
cn: admins
member: uid=john.doe,ou=users,dc=example,dc=com    ← User's full DN
member: uid=jane.doe,ou=users,dc=example,dc=com
```

#### Type 2: groupOfUniqueNames

```ldif
dn: cn=admins,ou=groups,dc=example,dc=com
objectClass: groupOfUniqueNames
cn: admins
uniqueMember: uid=john.doe,ou=users,dc=example,dc=com  ← Full DN
uniqueMember: uid=jane.doe,ou=users,dc=example,dc=com
```

#### Type 3: posixGroup

```ldif
dn: cn=admins,ou=groups,dc=example,dc=com
objectClass: posixGroup
cn: admins
gidNumber: 1000
memberUid: john.doe      ← Just the uid, not the full DN
memberUid: jane.doe
```

#### Type 4: Active Directory

```ldif
dn: CN=Admins,CN=Users,DC=example,DC=com
objectClass: group
cn: Admins
member: CN=John Doe,CN=Users,DC=example,DC=com     ← Full DN
member: CN=Jane Doe,CN=Users,DC=example,DC=com
```

---

## Step 5: Determine Filters

### LDAP_USER_FILTER

Based on login attribute identified in **Step 3.3**:

| Login Attribute | LDAP_USER_FILTER | System |
|-----------------|------------------|--------|
| `uid` | `(uid=%s)` | OpenLDAP, FreeIPA, 389 DS |
| `sAMAccountName` | `(sAMAccountName=%s)` | Active Directory |
| `cn` | `(cn=%s)` | Legacy systems |
| `mail` | `(mail=%s)` | Email login |
| `userPrincipalName` | `(userPrincipalName=%s)` | AD (email login) |

**The `%s` will be replaced by the username typed at login.**

### LDAP_GROUP_FILTER

Based on member attribute identified in **Step 4.3**:

| Member Attribute | LDAP_GROUP_FILTER | System |
|------------------|-------------------|--------|
| `member` | `(member=%s)` | groupOfNames, AD |
| `uniqueMember` | `(uniqueMember=%s)` | groupOfUniqueNames |
| `memberUid` | `(memberUid=%s)` | posixGroup |

**The `%s` will be replaced by the user's full DN** (e.g., `uid=john.doe,ou=users,dc=example,dc=com`)

**EXCEPTION:** For `posixGroup` with `memberUid`, Txlog Server needs to extract only the `uid` from the user's DN.

---

## Common Examples

### OpenLDAP with groupOfNames

```bash
# Structure
ou=users: uid=john.doe
ou=groups: cn=admins with member=uid=john.doe,ou=users,dc=example,dc=com

# Configuration
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(member=%s)
```

### OpenLDAP with posixGroup

```bash
# Structure
ou=users: uid=john.doe
ou=groups: cn=admins with memberUid=john.doe

# Configuration
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(memberUid=%s)
```

**⚠️ IMPORTANT:** For posixGroup, you need to modify Txlog Server code to extract only the `uid` from the DN before doing the group search. Currently, it passes the full DN.

### Active Directory

```bash
# Structure
CN=Users: sAMAccountName=john.doe
CN=Groups: cn=Admins with member=CN=John Doe,CN=Users,DC=example,DC=com

# Configuration
LDAP_USER_FILTER=(sAMAccountName=%s)
LDAP_GROUP_FILTER=(member=%s)
```

### FreeIPA

```bash
# Structure
cn=users: uid=john.doe
cn=groups: cn=admins with member=uid=john.doe,cn=users,cn=accounts,dc=example,dc=com

# Configuration
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(member=%s)
```

### 389 Directory Server

```bash
# Structure similar to OpenLDAP
ou=People: uid=john.doe
ou=Groups: cn=admins with member=uid=john.doe,ou=People,dc=example,dc=com

# Configuration
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(member=%s)
```

---

## Testing Filters

### Test LDAP_USER_FILTER

```bash
# Replace %s with real username
ldapsearch -H ldap://your-ldap-server.com:389 \
  -x \
  -b "ou=users,dc=example,dc=com" \
  -D "cn=admin,dc=example,dc=com" \
  -W \
  -LLL \
  "(uid=john.doe)"

# If it returns exactly 1 user, the filter is correct
```

### Test LDAP_GROUP_FILTER

```bash
# First, get the user's full DN
USER_DN="uid=john.doe,ou=users,dc=example,dc=com"

# Replace %s with user DN
ldapsearch -H ldap://your-ldap-server.com:389 \
  -x \
  -b "cn=admins,ou=groups,dc=example,dc=com" \
  -D "cn=admin,dc=example,dc=com" \
  -W \
  -s base \
  -LLL \
  "(member=uid=john.doe,ou=users,dc=example,dc=com)"

# If it returns the group, the filter is correct
```

---

## Advanced Filters

### Combining Multiple Attributes

```bash
# Search user by uid OR email
LDAP_USER_FILTER=(|(uid=%s)(mail=%s))

# Search user by sAMAccountName OR userPrincipalName (AD)
LDAP_USER_FILTER=(|(sAMAccountName=%s)(userPrincipalName=%s))
```

### Filter by ObjectClass

```bash
# Ensure it is an inetOrgPerson with specific uid
LDAP_USER_FILTER=(&(objectClass=inetOrgPerson)(uid=%s))

# Ensure it is a specific group with member
LDAP_GROUP_FILTER=(&(objectClass=groupOfNames)(member=%s))
```

---

## Troubleshooting

### Error: "user not found"

1. Verify if `LDAP_BASE_DN` is correct.
2. Test `LDAP_USER_FILTER` manually with `ldapsearch`.
3. Verify if user actually exists in the directory.

### Error: "not a member of any authorized group"

1. Verify if `LDAP_ADMIN_GROUP` or `LDAP_VIEWER_GROUP` is correct (must be full group DN).
2. Test `LDAP_GROUP_FILTER` manually with `ldapsearch`.
3. Verify if user is actually a member of the group in LDAP.

### Error: "failed to connect to LDAP"

1. Verify if host and port are correct.
2. Test connectivity: `telnet ldap-server 389` or `openssl s_client -connect ldap-server:636`.
3. Verify firewall and network rules.

### Error: "failed to bind with service account"

1. Verify if `LDAP_BIND_DN` is correct (full DN format).
2. Verify if password in `LDAP_BIND_PASSWORD` is correct.
3. Test bind manually with `ldapsearch`.

---

## Additional Resources

- **OpenLDAP Documentation**: <https://www.openldap.org/doc/>
- **Active Directory LDAP**: <https://docs.microsoft.com/en-us/windows/win32/adsi/search-filter-syntax>
- **FreeIPA Documentation**: <https://www.freeipa.org/page/Documentation>
- **Apache Directory Studio**: <https://directory.apache.org/studio/>
- **LDAP Filter Syntax**: <https://ldap.com/ldap-filters/>

---

## Complete Configuration Example

```bash
# OpenLDAP with groupOfNames
LDAP_HOST=ldap.example.com
LDAP_PORT=389
LDAP_USE_TLS=false
LDAP_BASE_DN=dc=example,dc=com
LDAP_BIND_DN=cn=admin,dc=example,dc=com
LDAP_BIND_PASSWORD=secret_password
LDAP_USER_FILTER=(uid=%s)
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
LDAP_GROUP_FILTER=(member=%s)
```

```bash
# Active Directory
LDAP_HOST=ad.example.com
LDAP_PORT=636
LDAP_USE_TLS=true
LDAP_BASE_DN=DC=example,DC=com
LDAP_BIND_DN=CN=Service Account,CN=Users,DC=example,DC=com
LDAP_BIND_PASSWORD=secret_password
LDAP_USER_FILTER=(sAMAccountName=%s)
LDAP_ADMIN_GROUP=CN=Txlog Admins,CN=Users,DC=example,DC=com
LDAP_VIEWER_GROUP=CN=Txlog Viewers,CN=Users,DC=example,DC=com
LDAP_GROUP_FILTER=(member=%s)
```

---

## Conclusion

Each LDAP server can have a different structure. Use exploration tools (`ldapsearch` or Apache Directory Studio) to:

1. ✅ Identify where users are stored.
2. ✅ Identify which attribute is used for login (uid, sAMAccountName, etc.).
3. ✅ Identify where groups are stored.
4. ✅ Identify which attribute stores members (member, uniqueMember, memberUid).
5. ✅ Test filters manually before configuring Txlog Server.

With this information, you can correctly configure LDAP filters!
