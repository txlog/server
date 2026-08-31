# How to Discover LDAP Filters for Your Directory

`LDAP_USER_FILTER` and `LDAP_GROUP_FILTER` have to match your directory's
schema, and the defaults (`(uid=%s)` and `(member=%s)`) are right for stock
OpenLDAP but wrong for Active Directory and for posixGroup setups. This guide
finds the correct values for yours.

If your directory is one of the common ones, the table in the
[LDAP configuration reference](../reference/ldap-configuration.md) probably
already has your answer — start there and come back here if it does not work.

## The fast path

[`scripts/ldap-discovery.sh`](../../scripts/ldap-discovery.sh) asks for your
connection details and walks the directory interactively, printing the filters
to use:

```bash
./scripts/ldap-discovery.sh
```

The rest of this guide is what that script does, by hand.

## Prerequisites

Install the OpenLDAP client tools:

```bash
sudo dnf install openldap-clients     # Red Hat, AlmaLinux, Rocky, Fedora
sudo apt install ldap-utils           # Debian, Ubuntu
brew install openldap                 # macOS
```

Then confirm you can reach the directory at all. Every command below repeats
these connection flags, so set them once:

```bash
LDAPURI=ldap://ldap.example.com:389
BASE=dc=example,dc=com
BIND="-D cn=readonly,dc=example,dc=com -w password"   # or just -x for anonymous
```

```bash
ldapsearch -H "$LDAPURI" $BIND -b "$BASE" -s base
```

An access error here means you need a service account before going further; see
[How to Configure LDAP Authentication](configure-ldap.md).

## Step 1: Find where users live

List the top-level structure to see which organizational units exist:

```bash
ldapsearch -H "$LDAPURI" $BIND -b "$BASE" -s one dn
```

Then find a real user anywhere under the base:

```bash
ldapsearch -H "$LDAPURI" $BIND -b "$BASE" "(objectClass=person)" dn | head
```

The DN that comes back tells you what `LDAP_BASE_DN` should be: the smallest
subtree that contains **every** user who needs access. If users are spread
across several OUs, use their common ancestor rather than listing one.

## Step 2: Find the login attribute

Look at one user's entry and see which attribute holds the name they would type
on a login form:

```bash
ldapsearch -H "$LDAPURI" $BIND -b "$BASE" "(cn=John Doe)" \
  uid cn sAMAccountName userPrincipalName mail
```

Whichever attribute contains the login name is the one your filter uses:

| Attribute present   | `LDAP_USER_FILTER`       |
| ------------------- | ------------------------ |
| `uid`               | `(uid=%s)`               |
| `sAMAccountName`    | `(sAMAccountName=%s)`    |
| `userPrincipalName` | `(userPrincipalName=%s)` |
| `mail`              | `(mail=%s)`              |

Verify it before moving on — substitute a real username for `%s`:

```bash
ldapsearch -H "$LDAPURI" $BIND -b "$BASE" "(uid=john)" dn
```

Exactly one entry means the filter is right. Zero means the wrong attribute;
more than one means the attribute is not unique and you should pick another.

## Step 3: Find the group membership attribute

Look at one of the groups you plan to use:

```bash
ldapsearch -H "$LDAPURI" $BIND -b "cn=admins,ou=groups,$BASE" \
  objectClass member uniqueMember memberUid
```

The `objectClass` and the member attribute together decide the filter:

| objectClass          | Member attribute | `LDAP_GROUP_FILTER` | `%s` expands to |
| -------------------- | ---------------- | ------------------- | --------------- |
| `groupOfNames`       | `member`         | `(member=%s)`       | the full DN     |
| `groupOfUniqueNames` | `uniqueMember`   | `(uniqueMember=%s)` | the full DN     |
| `group` (AD)         | `member`         | `(member=%s)`       | the full DN     |
| `posixGroup`         | `memberUid`      | `(memberUid=%s)`    | the uid alone   |

The distinction that catches people out: `posixGroup` records members as bare
uids (`john`), everything else as full DNs
(`uid=john,ou=users,dc=example,dc=com`). Using the wrong one matches nothing and
locks every user out with "not a member of any authorized group".

Verify it:

```bash
# For member/uniqueMember, pass the user's full DN
ldapsearch -H "$LDAPURI" $BIND -b "cn=admins,ou=groups,$BASE" \
  "(member=uid=john,ou=users,$BASE)" dn

# For memberUid, pass the bare uid
ldapsearch -H "$LDAPURI" $BIND -b "cn=admins,ou=groups,$BASE" \
  "(memberUid=john)" dn
```

One result means the user is in the group and the filter works.

## Step 4: Write it down

```bash
LDAP_BASE_DN=ou=users,dc=example,dc=com
LDAP_USER_FILTER=(uid=%s)
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
LDAP_GROUP_FILTER=(member=%s)
```

Restart the server and log in. If it still fails, set `LOG_LEVEL=DEBUG` — the
logs print the expanded filter the server actually sent, which you can paste
straight into `ldapsearch` to compare.
