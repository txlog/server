# How to Configure LDAP Authentication

This guide explains how to connect Txlog Server to an LDAP directory (like Active Directory or OpenLDAP) for user authentication.

## Prerequisites

- Access to an LDAP server.
- A service account (Bind DN and Password) if your server requires authentication to search.
- Distinguished Names (DNs) for the groups that will map to Admin and Viewer roles.

## Configuration Steps

1. **Open your `.env` file**.

2. **Set the LDAP connection variables**:

    ```bash
    # Connection details
    LDAP_HOST=ldap.example.com
    LDAP_PORT=389 # or 636 for LDAPS
    LDAP_USE_TLS=false # Set to true for LDAPS

    # Service Account (optional for some servers)
    LDAP_BIND_DN=cn=admin,dc=example,dc=com
    LDAP_BIND_PASSWORD=secret
    ```

3. **Set the Search and Group variables**:

    ```bash
    # Where to search for users
    LDAP_BASE_DN=ou=users,dc=example,dc=com

    # Filter to find a user by their username (input from login form)
    LDAP_USER_FILTER=(uid=%s)
    # For Active Directory, often: (sAMAccountName=%s)

    # Group Mapping (At least one is required)
    LDAP_ADMIN_GROUP=cn=txlog-admins,ou=groups,dc=example,dc=com
    LDAP_VIEWER_GROUP=cn=txlog-viewers,ou=groups,dc=example,dc=com

    # Filter to check group membership
    LDAP_GROUP_FILTER=(member=%s)
    # For Active Directory with nested groups, you might need a custom filter.
    ```

4. **Restart the Server**.

## Advanced Configuration

For detailed information on filters, error codes, and specific setups (like Active Directory), refer to the detailed
guides in the `docs/` folder:

- [LDAP Authentication Deep Dive](../LDAP_AUTHENTICATION.md)
- [LDAP Filter Guide](../LDAP_FILTERS_QUICK.md)
- [Service Account FAQ](../LDAP_SERVICE_ACCOUNT_FAQ.md)

## Troubleshooting

- **"Invalid Credentials"**: Check your `LDAP_BIND_DN` and `LDAP_BIND_PASSWORD`.
- **User found but not authorized**: The user might not be in the `LDAP_ADMIN_GROUP` or `LDAP_VIEWER_GROUP`. Check the
  `LDAP_GROUP_FILTER`.
