# Txlog Server Documentation

Welcome to the Txlog Server documentation. This documentation is divided into four types of guides:

## 📚 Documentation Index

### 1. Tutorials (Learning-oriented)

*Start here if you are new to the project.*

- **[Getting Started](tutorials/getting-started.md)**: Set up the server locally with Docker.
- **[Your First API Request](tutorials/first-api-request.md)**: Learn how to interact with the API.

### 2. How-to Guides (Task-oriented)

*Step-by-step guides to achieve specific goals.*

#### Authentication & Security

- **[Configure OIDC Authentication](how-to/configure-oidc.md)**: Connect with Google, Keycloak, etc.
- **[Configure LDAP Authentication](how-to/configure-ldap.md)**: Connect with Active Directory or OpenLDAP.
- **[Configure Anonymous LDAP](how-to/configure-ldap-anonymous.md)**: For servers without service accounts.
- **[Discover LDAP Filters](how-to/discover-ldap-filters.md)**: How to find the right query filters for your directory.
- **[Manage API Keys](how-to/manage-api-keys.md)**: Create and revoke keys for agents.

#### Operations

- **[Configure Data Retention](how-to/configure-data-retention.md)**: Manage database cleanup policies.
- **[Run Database Migrations](how-to/run-migrations.md)**: Apply schema changes safely.
- **[Deploy to Kubernetes](how-to/deploy-kubernetes.md)**: Production deployment manifest.

#### Development

- **[Add a New Endpoint](how-to/add-endpoint.md)**: Workflow for contributors.
- **[Run Tests](how-to/run-tests.md)**: Execute the test suite.

### 3. Reference (Information-oriented)

*Technical descriptions and specifications.*

#### System

- **[API Reference](reference/api-reference.md)**: High-level API overview.
- **[Database Schema](reference/database-schema.md)**: Tables, columns, and relationships.
- **[Environment Variables](reference/environment-variables.md)**: Complete configuration reference.

#### LDAP Specifics

- **[LDAP Cheatsheet](reference/ldap-cheatsheet.md)**: Quick reference for variables and common setups.
- **[LDAP Error Codes](reference/ldap-error-codes.md)**: Troubleshooting common error codes (32, 49, 50).
- **[LDAP Filters Reference](reference/ldap-filters.md)**: Common filter patterns for AD, OpenLDAP, etc.

### 4. Explanation (Understanding-oriented)

*Background knowledge and design decisions.*

#### Architecture

- **[System Architecture](explanation/architecture.md)**: High-level design, stack, and distributed scheduler.
- **[Data Model](explanation/data-model.md)**: Entities and relationships explanation.

#### Deep Dives

- **[LDAP Authentication Deep Dive](explanation/ldap-deep-dive.md)**: Comprehensive guide to how LDAP auth works.
- **[LDAP Implementation Details](explanation/ldap-implementation-details.md)**: Internal code structure of the LDAP module.
- **[LDAP Service Accounts FAQ](explanation/ldap-service-account-faq.md)**: Best practices for bind accounts.
- **[Testing Strategy](explanation/testing-strategy.md)**: Overview of the test suite and coverage goals.

---

## �️ Tools & Scripts

The `docs/` directory also contains useful scripts for administrators:

- **[ldap-discovery.sh](ldap-discovery.sh)**: An interactive script to help you discover your LDAP server's structure
  and test filters.

## 🔌 API Documentation (Swagger)

When the server is running, interactive API documentation is available at:
`http://localhost:8080/swagger/index.html`

- **Source**: `docs/docs.go` (Generated from code comments)
- **Update**: Run `make doc` to regenerate.

---

## 🤝 Contributing

When adding new features, please update the relevant documentation sections above. Ensure all Markdown files pass `markdownlint`.
