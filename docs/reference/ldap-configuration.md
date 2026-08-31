# LDAP Configuration Reference

Every LDAP setting Txlog Server reads, and the values that work for the common
directory servers. For the setup procedure see
[How to Configure LDAP Authentication](../how-to/configure-ldap.md); for how the
pieces fit together see
[LDAP Authentication](../explanation/ldap-authentication.md).

## Environment variables

| Variable             | Required | Default       | Description                                                     |
| -------------------- | -------- | ------------- | --------------------------------------------------------------- |
| `LDAP_HOST`          | Yes      | —             | Directory server hostname.                                      |
| `LDAP_PORT`          | No       | `389`/`636`   | Port. Defaults to 636 when `LDAP_USE_TLS=true`.                 |
| `LDAP_USE_TLS`       | No       | `false`       | `true` connects over LDAPS. The certificate is always verified. |
| `LDAP_BIND_DN`       | No       | —             | Service account DN. Omit for anonymous bind.                    |
| `LDAP_BIND_PASSWORD` | No       | —             | Service account password.                                       |
| `LDAP_BASE_DN`       | Yes      | —             | Subtree searched for users.                                     |
| `LDAP_USER_FILTER`   | No       | `(uid=%s)`    | User search filter. `%s` is the submitted username.             |
| `LDAP_ADMIN_GROUP`   | Yes\*    | —             | DN of the group granting admin access.                          |
| `LDAP_VIEWER_GROUP`  | Yes\*    | —             | DN of the group granting read-only access.                      |
| `LDAP_GROUP_FILTER`  | No       | `(member=%s)` | Membership filter. `%s` is the user's DN or uid.                |

\* At least one of `LDAP_ADMIN_GROUP` and `LDAP_VIEWER_GROUP` must be set, or
no one can log in.

There is no `LDAP_SKIP_TLS_VERIFY`. TLS certificates are always verified; trust
your private CA on the Txlog Server host instead.

## Minimal configurations

Without a service account — the directory must allow anonymous searches:

```bash
LDAP_HOST=ldap.example.com
LDAP_BASE_DN=ou=users,dc=example,dc=com
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
```

With a service account — required for Active Directory and hardened
directories:

```bash
LDAP_HOST=ldap.example.com
LDAP_BASE_DN=ou=users,dc=example,dc=com
LDAP_BIND_DN=cn=readonly,dc=example,dc=com
LDAP_BIND_PASSWORD=your_password
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
```

## Filters by directory server

| Directory                | `LDAP_USER_FILTER`       | `LDAP_GROUP_FILTER` |
| ------------------------ | ------------------------ | ------------------- |
| OpenLDAP                 | `(uid=%s)`               | `(member=%s)`       |
| OpenLDAP with posixGroup | `(uid=%s)`               | `(memberUid=%s)`    |
| Active Directory         | `(sAMAccountName=%s)`    | `(member=%s)`       |
| Active Directory (email) | `(userPrincipalName=%s)` | `(member=%s)`       |
| FreeIPA                  | `(uid=%s)`               | `(member=%s)`       |

### User filter, by login attribute

| Attribute           | Filter                   | Used by                          |
| ------------------- | ------------------------ | -------------------------------- |
| `uid`               | `(uid=%s)`               | OpenLDAP, FreeIPA, 389 Directory |
| `sAMAccountName`    | `(sAMAccountName=%s)`    | Active Directory                 |
| `userPrincipalName` | `(userPrincipalName=%s)` | Active Directory, email as login |
| `mail`              | `(mail=%s)`              | Any directory, email as login    |
| `cn`                | `(cn=%s)`                | Legacy schemas                   |

### Group filter, by group objectClass

| objectClass          | Member attribute | Filter              | `%s` expands to |
| -------------------- | ---------------- | ------------------- | --------------- |
| `groupOfNames`       | `member`         | `(member=%s)`       | the full DN     |
| `groupOfUniqueNames` | `uniqueMember`   | `(uniqueMember=%s)` | the full DN     |
| `group` (AD)         | `member`         | `(member=%s)`       | the full DN     |
| `posixGroup`         | `memberUid`      | `(memberUid=%s)`    | the uid alone   |

`posixGroup` is the one that takes a bare uid (`john`) rather than a DN. Using
`(member=%s)` against a posixGroup silently matches nothing, which surfaces as
"not a member of any authorized group" for every user.

If you do not know which of these your directory uses, run
[`scripts/ldap-discovery.sh`](../../scripts/ldap-discovery.sh) or follow
[Discover LDAP Filters](../how-to/discover-ldap-filters.md).

## Container example

```bash
docker run -d -p 8080:8080 \
  -e LDAP_HOST=ldap.example.com \
  -e LDAP_BASE_DN=ou=users,dc=example,dc=com \
  -e LDAP_BIND_DN=cn=readonly,dc=example,dc=com \
  -e LDAP_BIND_PASSWORD=your_password \
  -e LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com \
  -e LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com \
  ghcr.io/txlog/server:main
```

Drop `LDAP_BIND_DN` and `LDAP_BIND_PASSWORD` for a directory that allows
anonymous searches.

## Verifying a configuration by hand

`ldapsearch` answers the same questions the server asks, in the same order. If
these succeed, the server will too:

```bash
# 1. Does the user search work, with the credentials the server would use?
ldapsearch -H ldap://ldap.example.com:389 \
  -D "cn=readonly,dc=example,dc=com" -w "password" \
  -b "ou=users,dc=example,dc=com" "(uid=testuser)"

# 2. Can the user themselves bind? (This is the authentication step.)
ldapsearch -H ldap://ldap.example.com:389 \
  -D "uid=testuser,ou=users,dc=example,dc=com" -w "user_password" \
  -b "ou=users,dc=example,dc=com" "(uid=testuser)" dn

# 3. Is the user in the group?
ldapsearch -H ldap://ldap.example.com:389 \
  -D "cn=readonly,dc=example,dc=com" -w "password" \
  -b "cn=admins,ou=groups,dc=example,dc=com" \
  "(member=uid=testuser,ou=users,dc=example,dc=com)"
```

Drop `-D` and `-w` and add `-x` to test whether anonymous bind works, which is
what decides whether you need a service account at all.

## When something fails

| Symptom                                       | Look at                                                |
| --------------------------------------------- | ------------------------------------------------------ |
| Connection refused                            | `LDAP_HOST`, `LDAP_PORT`, firewall rules               |
| "user not found"                              | `LDAP_BASE_DN` and `LDAP_USER_FILTER`                  |
| "failed to bind with service account"         | `LDAP_BIND_DN` and `LDAP_BIND_PASSWORD`                |
| "invalid credentials" for a valid user        | the user's own password; check for a locked account    |
| "not a member of any authorized group"        | `LDAP_GROUP_FILTER`, and the group DNs                 |
| Group check fails only with a service account | the service account's read access to the group subtree |

Set `LOG_LEVEL=DEBUG` to see the base DN, the expanded filter and the result of
each operation. For LDAP result codes, see
[LDAP Error Codes](ldap-error-codes.md).
