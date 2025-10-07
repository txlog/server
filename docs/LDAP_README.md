# LDAP Documentation - Master Index

This directory contains complete documentation for LDAP authentication in Txlog Server.

## 📚 Available Documentation

### 🚀 Getting Started

1. **[LDAP_AUTHENTICATION.md](LDAP_AUTHENTICATION.md)**  
   📖 Complete LDAP authentication guide  
   - Feature overview
   - Step-by-step configuration
   - Environment variables
   - Practical examples
   - Troubleshooting

### 🔍 Discovering Your LDAP Filters

**Each LDAP server is different!** Use these resources to discover the correct values:

1. **[LDAP_FILTERS_QUICK.md](LDAP_FILTERS_QUICK.md)** ⭐ **START HERE**  
   ⚡ Quick and practical guide  
   - Reference table by server type
   - Ready-to-use commands
   - Common values (OpenLDAP, AD, FreeIPA)
   - How to test your filters

2. **[LDAP_FILTER_DISCOVERY.md](LDAP_FILTER_DISCOVERY.md)**  
   📘 Complete and detailed guide  
   - Step-by-step LDAP exploration
   - Explanation of each filter type
   - Using ldapsearch and Apache Directory Studio
   - Advanced troubleshooting

3. **[ldap-discovery.sh](ldap-discovery.sh)** 🛠️ **Interactive Script**  

   ```bash
   chmod +x ldap-discovery.sh
   ./ldap-discovery.sh
   ```

   - Interactive menu to explore your LDAP
   - Automatically tests connection
   - Discovers users and groups
   - Tests filters in real-time
   - Generates recommended configuration

### 📋 Quick References

4. **[LDAP_QUICK_REFERENCE.md](LDAP_QUICK_REFERENCE.md)**  
   📄 Configuration cheatsheet  
   - Summary of environment variables
   - .env examples by scenario
   - Useful commands

### 🔐 Service Account

5. **[LDAP_SERVICE_ACCOUNT_FAQ.md](LDAP_SERVICE_ACCOUNT_FAQ.md)**  
   ❓ Service account FAQ  
   - When to use service account vs anonymous bind
   - How to create service account
   - Required permissions
   - Security best practices

6. **[LDAP_SEM_SERVICE_ACCOUNT.md](LDAP_SEM_SERVICE_ACCOUNT.md)**  
   🔓 How to use without service account (anonymous bind)  
   - Anonymous bind configuration
   - Limitations and considerations
   - When it's possible to use

### 🏗️ Technical Information

7. **[LDAP_IMPLEMENTATION_SUMMARY.md](LDAP_IMPLEMENTATION_SUMMARY.md)**  
   🔧 Implementation details  
   - Code architecture
   - Authentication flow
   - Database structure
   - API endpoints

### 🚨 Troubleshooting

8. **[LDAP_ERROR_CODES.md](LDAP_ERROR_CODES.md)**  
   🔍 LDAP error codes explained  
   - **LDAP Result Code 32: No Such Object** (most common)
   - LDAP Result Code 49: Invalid Credentials
   - LDAP Result Code 50: Insufficient Access
   - How to diagnose each error
   - Practical solutions

---

## 🎯 Recommended Configuration Flow

```text
┌─────────────────────────────────────────────────────────────┐
│ 1. Read LDAP_AUTHENTICATION.md                              │
│    └─ Understand how it works and what's needed             │
└─────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. Use LDAP_FILTERS_QUICK.md or ldap-discovery.sh          │
│    └─ Discover the correct values for your LDAP             │
└─────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. Configure .env with discovered values                    │
│    └─ Follow examples from LDAP_QUICK_REFERENCE.md         │
└─────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. Start server and test login                              │
│    └─ Use troubleshooting if needed                         │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔍 Which Document to Use?

### Need to discover LDAP_USER_FILTER and LDAP_GROUP_FILTER?

➡️ **Use:** `LDAP_FILTERS_QUICK.md` or `./ldap-discovery.sh`

### First time configuring LDAP?

➡️ **Use:** `LDAP_AUTHENTICATION.md` → then `LDAP_FILTERS_QUICK.md`

### My LDAP doesn't have a service account?

➡️ **Use:** `LDAP_SEM_SERVICE_ACCOUNT.md`

### Need a quick .env example?

➡️ **Use:** `LDAP_QUICK_REFERENCE.md`

### Have questions about security/permissions?

➡️ **Use:** `LDAP_SERVICE_ACCOUNT_FAQ.md`

### Having authentication problems?

➡️ **Use:** Troubleshooting section of `LDAP_AUTHENTICATION.md` or `LDAP_ERROR_CODES.md`

### Getting "LDAP Result Code 32: No Such Object"?

➡️ **Use:** `LDAP_ERROR_CODES.md` - Explains this error in detail

### Want to understand how it works internally?

➡️ **Use:** `LDAP_IMPLEMENTATION_SUMMARY.md`

### Need all details about filters?

➡️ **Use:** `LDAP_FILTER_DISCOVERY.md`

---

## 🌟 Configuration Examples by Server

### OpenLDAP

```bash
LDAP_HOST=ldap.company.com
LDAP_PORT=389
LDAP_BASE_DN=dc=company,dc=com
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(member=%s)
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=company,dc=com
```

### Active Directory

```bash
LDAP_HOST=ad.company.com
LDAP_PORT=636
LDAP_USE_TLS=true
LDAP_BASE_DN=DC=company,DC=com
LDAP_USER_FILTER=(sAMAccountName=%s)
LDAP_GROUP_FILTER=(member=%s)
LDAP_ADMIN_GROUP=CN=Txlog Admins,OU=Groups,DC=company,DC=com
```

### FreeIPA

```bash
LDAP_HOST=ipa.company.com
LDAP_PORT=389
LDAP_BASE_DN=dc=company,dc=com
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(member=%s)
LDAP_ADMIN_GROUP=cn=admins,cn=groups,cn=accounts,dc=company,dc=com
```

---

## 🛠️ Useful Tools

### Discovery Script (Recommended)

```bash
./ldap-discovery.sh
```

### ldapsearch Commands

```bash
# Test connection
ldapsearch -H ldap://server:389 -x -b "dc=example,dc=com" -D "cn=admin,dc=example,dc=com" -W "(objectClass=*)"

# Search users
ldapsearch -H ldap://server:389 -x -b "dc=example,dc=com" -D "cn=admin,dc=example,dc=com" -W "(uid=user)"

# Search groups
ldapsearch -H ldap://server:389 -x -b "dc=example,dc=com" -D "cn=admin,dc=example,dc=com" -W "(cn=admins)"
```

### GUI Tools

- **Apache Directory Studio**: <https://directory.apache.org/studio/>
- **JXplorer**: <http://jxplorer.org/>
- **ldp.exe** (Windows Server - built-in)

---

## 📞 Support

If you're having problems:

1. ✅ Check the **Troubleshooting** section in `LDAP_AUTHENTICATION.md`
2. ✅ Use the `ldap-discovery.sh` script to validate your configuration
3. ✅ Check server logs (DEBUG level shows more details)
4. ✅ Test filters manually with `ldapsearch`

---

## 🔐 Security

**Best practices:**

- ✅ Use TLS/LDAPS in production (`LDAP_USE_TLS=true`)
- ✅ Use service account with minimal permissions (read-only)
- ✅ Never use `LDAP_SKIP_TLS_VERIFY=true` in production
- ✅ Keep passwords in `.env` and add `.env` to `.gitignore`
- ✅ Configure separate groups for admins and viewers

---

## 📊 Implementation Status

✅ Functional LDAP authentication  
✅ Support for multiple LDAP server types  
✅ Web session integration  
✅ Group-based access control (admin/viewer)  
✅ User synchronization with local database  
✅ Unified login interface (OIDC + LDAP)  
✅ Configuration via environment variables  
✅ Complete documentation  
✅ Interactive discovery script  

---

## 🚀 Version

Documentation updated for Txlog Server v1.14.0+

---

## Happy configuring! 🎉

If you have questions or suggestions, open an issue in the repository.
