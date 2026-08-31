# LDAP Authentication

This document explains how Txlog Server authenticates against an LDAP directory
and why the design makes the choices it does. For the values to put in your
`.env`, see the [LDAP configuration reference](../reference/ldap-configuration.md);
for the steps to set it up, see
[How to Configure LDAP Authentication](../how-to/configure-ldap.md).

LDAP is one of two optional authentication backends, alongside OIDC. Both can be
enabled at the same time, and both produce the same `models.User`. When neither
is configured, the server runs with no authentication at all.

## The authentication flow

A login does four things in order: find the user, verify the password, decide
the role, and create a session.

1. The user submits a username and password on `/login`.
2. The server opens a connection to the LDAP host, over TLS when
   `LDAP_USE_TLS=true`.
3. It searches `LDAP_BASE_DN` for the user, applying `LDAP_USER_FILTER` with
   the submitted username substituted for `%s`. This search runs as the service
   account when one is configured, and anonymously otherwise.
4. It **binds as the user's own DN with the submitted password**. This is the
   authentication step: LDAP itself decides whether the password is correct, and
   the server never sees a stored hash. Passwords are never written to the
   database.
5. It checks membership in `LDAP_ADMIN_GROUP` and `LDAP_VIEWER_GROUP` using
   `LDAP_GROUP_FILTER`. With a service account, the server re-binds as the
   service account first; without one, the check runs on the session it just
   authenticated as the user.
6. It creates or updates the user row and issues a session cookie.

A user who matches neither group is refused, even with a correct password.
Membership is what grants access; a valid password alone is not enough.

## Why the bind, and not a password comparison

Authenticating by binding means the directory remains the only authority on
credentials. Password policies, lockouts and expiry all keep working, because
the server is not reimplementing any of them — it asks LDAP to open a session
and treats a refusal as a failed login. It also means the server needs no read
access to password attributes, which most directories will not grant anyway.

## Attribute mapping

The server reads three things from the directory entry:

| Field        | Source attributes     | Fallback         |
| ------------ | --------------------- | ---------------- |
| Username     | the login form input  | —                |
| Email        | `mail`                | `username@local` |
| Display name | `cn` or `displayName` | the username     |

LDAP users are stored with a `sub` of `ldap:<username>`, while OIDC users keep
the `sub` their provider issued. The prefix is what keeps the two backends from
colliding on a shared username, and it is why enabling LDAP needs no database
migration: it reuses the existing `users` and `user_sessions` tables.

## Roles

The model is deliberately two-tier, matching what the dashboard actually
distinguishes:

- **Admin group** — everything a viewer can do, plus the admin panel, API key
  management, user management, and asset and package management.
- **Viewer group** — read-only access to transactions, assets, packages and
  statistics.

A user may be in both. Admin wins.

## The service account is optional

`LDAP_BIND_DN` and `LDAP_BIND_PASSWORD` are not required. Without them the user
search runs over an anonymous bind, and the group check runs as the
just-authenticated user. This works whenever the directory allows anonymous
reads on user entries and lets a user read their own group memberships — the
default for stock OpenLDAP and 389 Directory Server.

A service account becomes necessary when the directory refuses those reads.
Active Directory is the common case: it blocks anonymous bind by default, so a
search fails before authentication is ever attempted. A hardened OpenLDAP or a
FreeIPA deployment usually needs one too.

| Directory                | Service account | Why                                 |
| ------------------------ | --------------- | ----------------------------------- |
| OpenLDAP (default ACLs)  | Not needed      | Anonymous bind enabled by default   |
| OpenLDAP (hardened ACLs) | Needed          | Anonymous bind disabled             |
| 389 Directory Server     | Usually not     | Typically allows anonymous reads    |
| Active Directory         | Needed          | Blocks anonymous bind by policy     |
| FreeIPA                  | Needed          | Restrictive default access policies |

Neither choice is simply more secure than the other. A service account gives you
an auditable identity behind every search and lets you disable anonymous bind
entirely, at the cost of one more credential to store and rotate. Anonymous bind
has no credential to leak, but every search is unattributable. If you have
compliance requirements that name who read what, you want the service account;
if you are running a homelab against a permissive OpenLDAP, you do not.

Whichever you pick is a server-side setting with no effect on users, so it is
safe to start without one, try a login, and add the two variables only if the
search fails. When you do use one, give it read access to the user and group
subtrees and nothing else.

## TLS

`LDAP_USE_TLS=true` connects over TLS and **always verifies the certificate**.
There is no skip-verify option: an unverified TLS connection to a directory
server is a credential-forwarding hazard, since the password travels over it in
the bind. If your directory uses a private CA, install that CA on the Txlog
Server host rather than looking for a way to disable the check.

Failed authentication attempts are logged. The server does no rate limiting of
its own — put that in front of it, in the reverse proxy.
