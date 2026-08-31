# LDAP Error Codes

The result codes an LDAP directory returns, what each one means for Txlog
Server, and which setting to look at. For the settings themselves see the
[LDAP configuration reference](ldap-configuration.md).

## Code summary

| Code | Name                   | What it means                        | Where to look                                 |
| ---- | ---------------------- | ------------------------------------ | --------------------------------------------- |
| 0    | Success                | The operation worked                 | —                                             |
| 32   | No Such Object         | A DN does not exist                  | `LDAP_BASE_DN`, `LDAP_BIND_DN`, the group DNs |
| 34   | Invalid DN Syntax      | A DN is malformed                    | Missing or stray commas in a DN               |
| 49   | Invalid Credentials    | A bind was refused                   | The password, or a locked account             |
| 50   | Insufficient Access    | The bound identity may not read that | The service account's ACLs                    |
| 52   | Unavailable            | The server is not answering          | `LDAP_HOST`, `LDAP_PORT`, firewall            |
| 53   | Unwilling to Perform   | The server refuses the operation     | Server policy, e.g. plaintext binds           |
| 65   | Object Class Violation | The entry breaks the schema          | The directory's schema                        |

## Code 32: No Such Object

By far the most common. It means one of the DNs in your configuration points at
something that is not there. Four settings can produce it, and the message alone
does not say which — enable `LOG_LEVEL=DEBUG` and match the log line:

| Log message                                | Wrong DN     | Variable            |
| ------------------------------------------ | ------------ | ------------------- |
| `LDAP user search: baseDN=...`             | Base DN      | `LDAP_BASE_DN`      |
| `Binding with service account: ...`        | Bind DN      | `LDAP_BIND_DN`      |
| `LDAP search failed: ... filter=(uid=...)` | Base DN      | `LDAP_BASE_DN`      |
| `Failed to check admin group membership`   | Admin group  | `LDAP_ADMIN_GROUP`  |
| `Failed to check viewer group membership`  | Viewer group | `LDAP_VIEWER_GROUP` |

Then confirm the DN exists, using the same credentials the server uses:

```bash
# Does the base DN exist?
ldapsearch -x -H ldap://ldap.example.com:389 -b "ou=users,dc=example,dc=com" -s base

# Does the service account's own DN exist?
ldapsearch -x -H ldap://ldap.example.com:389 -b "cn=readonly,dc=example,dc=com" -s base

# Does the group exist at that exact path?
ldapsearch -x -H ldap://ldap.example.com:389 -b "cn=admins,ou=groups,dc=example,dc=com" -s base
```

A base DN that exists but is *too narrow* gives the same symptom for some users
and not others: if `LDAP_BASE_DN=ou=employees,dc=example,dc=com` but a
contractor lives under `ou=contractors`, only the contractors fail. Search the
whole tree to find where the user really is:

```bash
ldapsearch -x -H ldap://ldap.example.com:389 -b "dc=example,dc=com" "(uid=john)" dn
```

Then widen `LDAP_BASE_DN` to a common ancestor.

## Code 34: Invalid DN Syntax

A DN with a missing comma or a stray space between components:

```bash
LDAP_BASE_DN=ou=users dc=example,dc=com     # wrong
LDAP_BASE_DN=ou=users,dc=example,dc=com     # right
```

## Code 49: Invalid Credentials

The bind was refused. Which bind matters:

- During the **user search**, it is the service account: `LDAP_BIND_DN` or
  `LDAP_BIND_PASSWORD` is wrong.
- During **authentication**, it is the user's own password — or their account
  is locked, disabled or expired in the directory.

Test the service account on its own:

```bash
ldapsearch -H ldap://ldap.example.com:389 \
  -D "cn=readonly,dc=example,dc=com" -w "password" \
  -b "dc=example,dc=com" -s base
```

## Code 50: Insufficient Access Rights

The DN exists and the bind succeeded, but the bound identity may not read it.
Usually the service account has no read access to the group subtree, so
authentication works and authorization then fails for everyone. Grant it read
on the groups, or use a directory-appropriate service account.

## Code 52: Unavailable

The server is unreachable rather than unhappy. Check connectivity from the
Txlog Server host:

```bash
nc -vz ldap.example.com 389
openssl s_client -connect ldap.example.com:636   # for LDAPS
```

If the LDAPS handshake fails on certificate verification, the fix is to trust
the issuing CA on the Txlog Server host. Certificate verification cannot be
disabled.

## Diagnosing anything else

1. Set `LOG_LEVEL=DEBUG` and restart. The logs print the base DN, the expanded
   filter and the outcome of every LDAP operation.
2. Reproduce the failing login.
3. Reproduce the same operation with `ldapsearch`, using the credentials from
   the log line. `ldapsearch` reports the numeric result code directly.
4. Apache Directory Studio (<https://directory.apache.org/studio/>) is worth
   installing if you need to browse the tree visually to find the right DNs.
