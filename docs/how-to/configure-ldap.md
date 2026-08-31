# How to Configure LDAP Authentication

Connect Txlog Server to an LDAP directory — Active Directory, OpenLDAP, FreeIPA
— so users log in with their directory credentials.

## Prerequisites

- Network access from the Txlog Server host to the directory.
- The base DN under which user entries live.
- The DN of at least one group whose members should have access.
- A service account (bind DN and password) **only if** your directory refuses
  anonymous searches. Step 1 tells you whether it does.

## Step 1: Find out whether you need a service account

Try an anonymous search for a real user:

```bash
ldapsearch -x -H ldap://ldap.example.com:389 \
  -b "ou=users,dc=example,dc=com" "(uid=youruser)"
```

If it returns the entry, skip the service account. If it fails with an access
error, you need one — Active Directory almost always does.

## Step 2: Set the connection variables

In your `.env`:

```bash
LDAP_HOST=ldap.example.com
LDAP_PORT=389            # 636 for LDAPS
LDAP_USE_TLS=false       # true for LDAPS
```

> [!IMPORTANT]
> With `LDAP_USE_TLS=true` the server always verifies the directory's
> certificate, and there is no way to turn that off. The user's password
> travels over this connection in the bind. If your directory uses a private
> CA, install that CA on the Txlog Server host.

Add the service account only if step 1 said you need one:

```bash
LDAP_BIND_DN=cn=readonly,dc=example,dc=com
LDAP_BIND_PASSWORD=secret
```

Give that account read access to the user and group subtrees and nothing more.

## Step 3: Set the search and group variables

```bash
# Where user entries live
LDAP_BASE_DN=ou=users,dc=example,dc=com

# How to find a user from the name typed on the login form
LDAP_USER_FILTER=(uid=%s)              # (sAMAccountName=%s) for Active Directory

# Who gets in, and as what. At least one of the two is required.
LDAP_ADMIN_GROUP=cn=txlog-admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=txlog-viewers,ou=groups,dc=example,dc=com

# How membership is recorded in those groups
LDAP_GROUP_FILTER=(member=%s)          # (memberUid=%s) for posixGroup
```

A user in neither group cannot log in, even with the right password.

If you are unsure which filters your directory wants, the
[LDAP configuration reference](../reference/ldap-configuration.md) has a table
per server type, and [Discover LDAP Filters](discover-ldap-filters.md) walks
you through finding them.

## Step 4: Restart and test

Restart the server and log in at `/login`. A user in the admin group should see
the admin panel; a user in the viewer group should not.

## Troubleshooting

Set `LOG_LEVEL=DEBUG` first — the logs name the base DN, the expanded filter and
the failing operation, which is usually enough to identify the wrong variable.

- **"user not found"** — `LDAP_BASE_DN` does not contain the user, or
  `LDAP_USER_FILTER` names the wrong attribute. Repeat the step 1 search with
  the exact filter you configured.
- **"failed to bind with service account"** — wrong `LDAP_BIND_DN` or
  `LDAP_BIND_PASSWORD`.
- **"not a member of any authorized group"** — the user authenticated, so the
  password is fine. Either the group DN is wrong or `LDAP_GROUP_FILTER` does
  not match how the group stores members: `posixGroup` uses `memberUid` with a
  bare uid, everything else uses `member` with a full DN.
- **Connection refused** — `LDAP_HOST`, `LDAP_PORT` or a firewall.

For LDAP result codes such as 32, 49 and 50, see
[LDAP Error Codes](../reference/ldap-error-codes.md).
