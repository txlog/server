# LDAP Error Codes - Troubleshooting Guide

## LDAP Result Code 32: No Such Object

### 🔍 What does it mean?

The error **"LDAP Result Code 32: No Such Object"** means that the LDAP server **could not find the object** (user, group, or DN) you are trying to access. It is like looking for a file that does not exist in a directory.

### 📍 Where Can It Occur?

This error can happen in **4 situations** in Txlog Server:

#### 1. **Incorrect Base DN** (Most Common)

```bash
# ❌ WRONG
LDAP_BASE_DN=ou=users,dc=example,dc=com

# ✅ CORRECT
LDAP_BASE_DN=dc=example,dc=com
```

**Problem:** The `LDAP_BASE_DN` is pointing to an OU that does not exist or is incorrect.

**How to Verify:**

```bash
# Test if the Base DN exists
ldapsearch -H ldap://server:389 -x -D "cn=admin,dc=example,dc=com" -W \
  -b "dc=example,dc=com" -s base "(objectClass=*)"

# If it returns error 32, the Base DN is wrong
```

**Solution:**

1. Discover the correct Base DN by exploring the server:

   ```bash
   ldapsearch -H ldap://server:389 -x -D "cn=admin,dc=example,dc=com" -W \
     -b "" -s base namingContexts
   ```

2. Update in `.env`:

   ```bash
   LDAP_BASE_DN=dc=example,dc=com  # Use the correct value
   ```

---

#### 2. **Incorrect Bind DN**

```bash
# ❌ WRONG
LDAP_BIND_DN=cn=readonly,dc=example,dc=com

# ✅ CORRECT
LDAP_BIND_DN=cn=readonly,ou=service-accounts,dc=example,dc=com
```

**Problem:** The service account (Bind DN) does not exist in the specified path.

**How to Verify:**

```bash
# Test the Bind DN
ldapsearch -H ldap://server:389 -x \
  -D "cn=readonly,ou=service-accounts,dc=example,dc=com" \
  -W -b "dc=example,dc=com" "(objectClass=*)"

# If it returns error 32, the Bind DN does not exist
```

**Solution:**

1. Search for the service account:

   ```bash
   # Search by CN
   ldapsearch -H ldap://server:389 -x -D "cn=admin,dc=example,dc=com" -W \
     -b "dc=example,dc=com" "(cn=readonly)" dn
   ```

2. Use the full DN returned in `.env`.

---

#### 3. **Incorrect Admin Group or Viewer Group**

```bash
# ❌ WRONG
LDAP_ADMIN_GROUP=cn=admins,ou=grupos,dc=example,dc=com

# ✅ CORRECT
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
```

**Problem:** The group DN does not exist.

**How to Verify:**

```bash
# Test if the group exists
ldapsearch -H ldap://server:389 -x -D "cn=admin,dc=example,dc=com" -W \
  -b "cn=admins,ou=groups,dc=example,dc=com" -s base "(objectClass=*)"

# If it returns error 32, the group does not exist at this path
```

**Solution:**

1. Search for the correct group:

   ```bash
   ldapsearch -H ldap://server:389 -x -D "cn=admin,dc=example,dc=com" -W \
     -b "dc=example,dc=com" "(cn=admins)" dn
   ```

2. Use the full group DN in `.env`:

   ```bash
   LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
   ```

---

#### 4. **User Not Found in Base DN**

```bash
# Base DN too restrictive
LDAP_BASE_DN=ou=employees,dc=example,dc=com

# But the user is at: uid=john,ou=contractors,dc=example,dc=com
```

**Problem:** The user exists in LDAP, but **outside** the configured Base DN.

**How to Verify:**

```bash
# Search for the user in the entire directory
ldapsearch -H ldap://server:389 -x -D "cn=admin,dc=example,dc=com" -W \
  -b "dc=example,dc=com" "(uid=john)" dn

# If you find the user in a different OU, expand the Base DN
```

**Solution:**

- Use a broader Base DN that includes all users:

  ```bash
  # Instead of:
  LDAP_BASE_DN=ou=employees,dc=example,dc=com

  # Use:
  LDAP_BASE_DN=dc=example,dc=com
  ```

---

## 🔧 How to Diagnose Error 32 in Txlog Server

### Step 1: Enable DEBUG Logs

In `.env`:

```bash
LOG_LEVEL=DEBUG
```

Restart the server and try to log in. You will see detailed logs:

```text
time=... level=DEBUG msg="LDAP user search: baseDN=ou=users,dc=example,dc=com, filter=(uid=john)"
time=... level=ERROR msg="LDAP search failed: LDAP Result Code 32 \"No Such Object\""
```

### Step 2: Identify Which DN Is Incorrect

The logs show which operation failed:

| Log Message | Incorrect DN | .env Variable |
|-------------|--------------|---------------|
| "LDAP user search: baseDN=..." | Base DN | `LDAP_BASE_DN` |
| "Binding with service account: ..." | Bind DN | `LDAP_BIND_DN` |
| "LDAP search failed: ... filter=(uid=...)" | Base DN | `LDAP_BASE_DN` |
| "Failed to check admin group membership" | Admin Group | `LDAP_ADMIN_GROUP` |
| "Failed to check viewer group membership" | Viewer Group | `LDAP_VIEWER_GROUP` |

### Step 3: Validate the Correct DN

Use `ldapsearch` or the `ldap-discovery.sh` script:

```bash
./ldap-discovery.sh
# Option 1: Explore directory structure
# Option 2: Search users
# Option 3: Search groups
```

### Step 4: Fix and Test

1. Update `.env` with the correct DN.
2. Restart the server.
3. Try logging in again.

---

## 📋 Verification Checklist for Error 32

When encountering **"LDAP Result Code 32"**, check:

- [ ] **LDAP_BASE_DN** exists and is accessible?

  ```bash
  ldapsearch -H ldap://... -x -D "..." -W -b "dc=example,dc=com" -s base "(objectClass=*)"
  ```

- [ ] **LDAP_BIND_DN** exists (if configured)?

  ```bash
  ldapsearch -H ldap://... -x -D "cn=readonly,dc=example,dc=com" -W -b "dc=example,dc=com" -s base "(objectClass=*)"
  ```

- [ ] **LDAP_ADMIN_GROUP** exists?

  ```bash
  ldapsearch -H ldap://... -x -D "..." -W -b "cn=admins,ou=groups,dc=example,dc=com" -s base "(objectClass=*)"
  ```

- [ ] **LDAP_VIEWER_GROUP** exists (if configured)?

  ```bash
  ldapsearch -H ldap://... -x -D "..." -W -b "cn=viewers,ou=groups,dc=example,dc=com" -s base "(objectClass=*)"
  ```

- [ ] Users are within the **LDAP_BASE_DN**?

  ```bash
  ldapsearch -H ldap://... -x -D "..." -W -b "dc=example,dc=com" "(uid=user)"
  ```

---

## 🌟 Examples of Correct Configuration

### Typical OpenLDAP

```bash
LDAP_BASE_DN=dc=company,dc=com
LDAP_BIND_DN=cn=readonly,ou=service-accounts,dc=company,dc=com
LDAP_ADMIN_GROUP=cn=txlog-admins,ou=groups,dc=company,dc=com
LDAP_VIEWER_GROUP=cn=txlog-users,ou=groups,dc=company,dc=com
```

### Active Directory

```bash
LDAP_BASE_DN=DC=company,DC=com
LDAP_BIND_DN=CN=LDAP Service,OU=Service Accounts,DC=company,DC=com
LDAP_ADMIN_GROUP=CN=Txlog Admins,OU=Security Groups,DC=company,DC=com
LDAP_VIEWER_GROUP=CN=Txlog Users,OU=Security Groups,DC=company,DC=com
```

### FreeIPA

```bash
LDAP_BASE_DN=dc=company,dc=com
LDAP_BIND_DN=uid=readonly,cn=sysaccounts,cn=etc,dc=company,dc=com
LDAP_ADMIN_GROUP=cn=txlog-admins,cn=groups,cn=accounts,dc=company,dc=com
LDAP_VIEWER_GROUP=cn=txlog-users,cn=groups,cn=accounts,dc=company,dc=com
```

---

## 🔍 Other Common LDAP Error Codes

### Code 34: Invalid DN Syntax

**What it means:** The DN format is incorrect.

**Example:**

```bash
# ❌ WRONG (missing comma)
LDAP_BASE_DN=ou=usersdc=example,dc=com

# ✅ CORRECT
LDAP_BASE_DN=ou=users,dc=example,dc=com
```

### Code 49: Invalid Credentials

**What it means:** Incorrect username/password.

**Common causes:**

1. Wrong password in `LDAP_BIND_PASSWORD`.
2. The password of the user trying to log in is incorrect.
3. Service account expired or disabled.

**How to verify:**

```bash
# Test the Bind DN
ldapsearch -H ldap://server:389 -x \
  -D "cn=readonly,dc=example,dc=com" \
  -w "your_password" \
  -b "dc=example,dc=com" "(objectClass=*)"
```

### Code 50: Insufficient Access Rights

**What it means:** The account does not have permission to perform the operation.

**Solution:** The service account needs:

- Read permission on the Base DN.
- Read permission on the configured groups.

### Code 52: Unavailable

**What it means:** LDAP server is not available.

**Causes:**

1. LDAP server down.
2. Port blocked by firewall.
3. Network issues.

**How to verify:**

```bash
# Test connectivity
telnet ldap.server.com 389

# Or with LDAPS
openssl s_client -connect ldap.server.com:636
```

### Code 53: Unwilling to Perform

**What it means:** Server refused to execute the operation.

**Common causes:**

1. Attempt to modify data in read-only mode.
2. Server policy violation.
3. Operation not allowed (e.g., anonymous bind disabled).

---

## 🛠️ Diagnostic Tools

### 1. ldap-discovery.sh Script

```bash
./ldap-discovery.sh
# Use menu options to test each component
```

### 2. Manual ldapsearch

```bash
# Complete test template
ldapsearch -H ldap://SERVER:PORT \
  -x \
  -D "BIND_DN" \
  -W \
  -b "BASE_DN" \
  "FILTER" \
  attributes

# Real example
ldapsearch -H ldap://ldap.company.com:389 \
  -x \
  -D "cn=readonly,dc=company,dc=com" \
  -W \
  -b "dc=company,dc=com" \
  "(uid=john)" \
  dn uid cn mail
```

### 3. Apache Directory Studio (GUI)

- Download: <https://directory.apache.org/studio/>
- Allows visual browsing of the LDAP tree.
- Shows errors in a more user-friendly way.

### 4. Txlog Server Logs

```bash
# Enable DEBUG in .env
LOG_LEVEL=DEBUG

# Run the server
make run

# Logs will show:
# - Base DN used in searches
# - Applied filters
# - Results of each operation
# - Detailed errors
```

---

## 📊 LDAP Codes Summary Table

| Code | Name | Meaning | Common Solution |
|------|------|---------|-----------------|
| 0 | Success | Operation successful | N/A |
| 32 | No Such Object | DN does not exist | Check DNs in .env |
| 34 | Invalid DN Syntax | Incorrect DN format | Check commas and format |
| 49 | Invalid Credentials | Incorrect username/password | Check credentials |
| 50 | Insufficient Access | No permission | Adjust account ACLs |
| 52 | Unavailable | Server unavailable | Check connectivity |
| 53 | Unwilling to Perform | Operation not allowed | Check server policies |
| 65 | Object Class Violation | Issue with objectClass | Check schema |

---

## 🚨 Step-by-Step Troubleshooting

### When receiving "LDAP Result Code 32"

```bash
# 1. Enable detailed logs
echo "LOG_LEVEL=DEBUG" >> .env

# 2. Restart the server and try to log in
make run

# 3. In logs, identify which DN failed:
#    - "LDAP user search: baseDN=..." → problem in LDAP_BASE_DN
#    - "failed to bind with service account" → problem in LDAP_BIND_DN
#    - "Failed to check ... group membership" → problem in group

# 4. Test the DN manually:
ldapsearch -H ldap://server:389 \
  -x -D "cn=admin,dc=example,dc=com" -W \
  -b "SUSPICIOUS_DN" -s base "(objectClass=*)"

# 5. If it returns error 32, the DN is wrong
#    If it returns success, the DN exists (problem elsewhere)

# 6. Use the script to discover the correct DN:
./ldap-discovery.sh
# Option 1: Explore structure
# Option 2 or 3: Search for the correct object

# 7. Update .env with the correct DN

# 8. Restart and test again
```

---

## 📞 Need Help?

1. ✅ Use `./ldap-discovery.sh` to explore your LDAP.
2. ✅ Enable `LOG_LEVEL=DEBUG` to see details.
3. ✅ Test each DN manually with `ldapsearch`.
4. ✅ Consult `LDAP_FILTER_DISCOVERY.md` for a complete guide.

---

## ✨ Summary

**"LDAP Result Code 32: No Such Object"** = **DN does not exist**

**Always check:**

1. ✅ `LDAP_BASE_DN` - The starting point for searches.
2. ✅ `LDAP_BIND_DN` - The service account (if used).
3. ✅ `LDAP_ADMIN_GROUP` - The administrators group.
4. ✅ `LDAP_VIEWER_GROUP` - The viewers group.

**Use tools:**

- `./ldap-discovery.sh` - Interactive discovery.
- `ldapsearch` - Manual tests.
- `LOG_LEVEL=DEBUG` - Detailed logs.

🎯 In most cases, error 32 is caused by an **Incorrect Base DN**!
