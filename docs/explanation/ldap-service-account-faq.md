# LDAP Service Account - FAQ

## Is it mandatory to use a service account?

**NO!** The service account is **optional**. Txlog Server works perfectly without it in many scenarios.

## When do I NOT need a service account?

You can authenticate **WITHOUT a service account** when:

1. **OpenLDAP with anonymous bind enabled**
   - This is the default OpenLDAP configuration.
   - Allows searches without authentication.
   - Authenticated users can read their own groups.

2. **Development/Test Environment**
   - Faster configuration.
   - Fewer credentials to manage.
   - Facilitates testing.

3. **LDAP with permissive ACLs**
   - Allows anonymous read on users and groups.
   - Allows users to read their own groups.

## When do I NEED a service account?

You **need** a service account when:

1. **Active Directory**
   - AD usually blocks anonymous bind.
   - Requires authentication for searches.
   - Microsoft default security policy.

2. **OpenLDAP with restricted ACLs**
   - Anonymous bind disabled.
   - Users cannot read groups.
   - Corporate security policies.

3. **Compliance Requirements**
   - Access auditing.
   - Tracking who performs searches.
   - Company security policies.

## How does it work WITHOUT a service account?

```text
Authentication Flow:

1. User types username + password in Txlog
   ↓
2. Txlog connects to LDAP (without authentication)
   ↓
3. Searches user via anonymous bind
   ↓
4. Authenticates user with bind using their credentials
   ↓
5. Checks groups using the user's authenticated session
   ↓
6. Creates session in Txlog
```

## How does it work WITH a service account?

```text
Authentication Flow:

1. User types username + password in Txlog
   ↓
2. Txlog connects to LDAP
   ↓
3. Txlog binds with service account
   ↓
4. Searches user using service account
   ↓
5. Authenticates user with bind using user credentials
   ↓
6. Re-binds with service account
   ↓
7. Checks groups using service account
   ↓
8. Creates session in Txlog
```

## Which is more secure?

**Depends on your environment:**

### WITH Service Account is more secure when

- ✅ You need to track all LDAP accesses.
- ✅ You want to limit exactly which objects can be read.
- ✅ You want to disable anonymous bind (best practice).
- ✅ You have compliance/auditing requirements.

### WITHOUT Service Account can be equally secure when

- ✅ Anonymous bind only allows read (not write).
- ✅ LDAP ACLs are well configured.
- ✅ You are in a private/trusted network.
- ✅ You have other security controls.

## Which is easier to configure?

**WITHOUT service account** is much simpler:

```bash
# Only 4 variables!
LDAP_HOST=ldap.example.com
LDAP_BASE_DN=ou=users,dc=example,dc=com
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
```

vs

```bash
# WITH service account: 6 variables
LDAP_HOST=ldap.example.com
LDAP_BIND_DN=cn=svc-txlog,dc=example,dc=com      # +1
LDAP_BIND_PASSWORD=secret_password                # +2
LDAP_BASE_DN=ou=users,dc=example,dc=com
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
```

## How to test which option works for me?

### Test 1: Does anonymous bind work?

```bash
ldapsearch -H ldap://your-ldap:389 -x \
  -b "ou=users,dc=example,dc=com" \
  "(uid=youruser)"
```

- **Works?** → You can use WITHOUT service account.
- **Access error?** → You need a service account.

### Test 2: Can user read groups?

```bash
ldapsearch -H ldap://your-ldap:389 \
  -D "uid=youruser,ou=users,dc=example,dc=com" \
  -w "yourpassword" \
  -b "cn=admins,ou=groups,dc=example,dc=com"
```

- **Returns groups?** → Verification will work.
- **Error?** → Need service account with permissions.

## Recommendations by server type

| LDAP Server | Recommendation | Reason |
|-------------|----------------|--------|
| **OpenLDAP** (default) | ✅ WITHOUT service account | Anonymous bind enabled by default |
| **OpenLDAP** (hardened) | ⚠️ WITH service account | Anonymous bind disabled |
| **Active Directory** | ⚠️ WITH service account | Requires authentication for searches |
| **FreeIPA** | ⚠️ WITH service account | More restrictive policies |
| **389 Directory** | ✅ WITHOUT service account | Usually allows anonymous |

## Can I change later?

**YES!** You can:

1. **Start WITHOUT service account**
   - Test if it works.
   - If it works, leave it.
   - If not, add service account.

2. **Start WITH service account**
   - Works in any scenario.
   - Remove later if you want to simplify.

**No impact on users** - it is just server configuration.

## Practical Example: My First Configuration

### Step 1: Start simple (WITHOUT service account)

```bash
# .env
LDAP_HOST=ldap.company.com
LDAP_BASE_DN=ou=users,dc=company,dc=com
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=company,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=company,dc=com
```

### Step 2: Test login

- ✅ **Worked?** Done! Leave it like this.
- ❌ **Error "user not found"?** Add service account (Step 3).

### Step 3: If necessary, add service account

```bash
# .env (add these 2 lines)
LDAP_BIND_DN=cn=readonly,dc=company,dc=com
LDAP_BIND_PASSWORD=service_account_password
```

## Security: Service Account vs Anonymous Bind

### Service Account

**Security advantages:**

- ✅ Logs show which account performed each search.
- ✅ Can audit specific accesses.
- ✅ Can revoke access easily.
- ✅ Can limit exactly what is accessible.

**Disadvantages:**

- ❌ One more credential to protect.
- ❌ Password can leak.
- ❌ Need to manage password rotation.

### Anonymous Bind

**Security advantages:**

- ✅ No credentials to leak.
- ✅ No password to manage.
- ✅ Simpler = less chance of error.

**Disadvantages:**

- ❌ Harder to audit accesses.
- ❌ Anyone can perform searches.
- ❌ May not meet corporate policies.

## Conclusion

| Criteria | WITHOUT Service Account | WITH Service Account |
|----------|-------------------------|----------------------|
| **Simplicity** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Security** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Compatibility** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Auditing** | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Maintenance** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |

**Quick Answer:**

- 🏠 **Homelab/Development?** → WITHOUT service account
- 🏢 **Production/Enterprise?** → WITH service account
- 💼 **Active Directory?** → WITH service account (mandatory)
- 🐧 **Simple OpenLDAP?** → WITHOUT service account
- 📋 **Have compliance?** → WITH service account

## Need help?

1. Consult [configure-ldap-anonymous.md](../how-to/configure-ldap-anonymous.md) for full guide.
2. See [ldap-cheatsheet.md](../reference/ldap-cheatsheet.md) for quick examples.
3. Read [ldap-deep-dive.md](../explanation/ldap-deep-dive.md) for full documentation.
