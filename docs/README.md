# Txlog Server Documentation

Welcome to the Txlog Server documentation. This documentation is divided into four types of guides:

## 📚 Documentation Index

### 1. Tutorials (Learning-oriented)

_Start here if you are new to the project._

- **[Getting Started](tutorials/getting-started.md)**: Set up the server locally with Docker.
- **[Your First API Request](tutorials/first-api-request.md)**: Learn how to interact with the API.

### 2. How-to Guides (Task-oriented)

_Step-by-step guides to achieve specific goals._

#### Authentication & Security

- **[Configure OIDC Authentication](how-to/configure-oidc.md)**: Connect with Google, Keycloak, etc.
- **[Configure LDAP Authentication](how-to/configure-ldap.md)**: Connect with Active Directory or OpenLDAP, with or
  without a service account.
- **[Discover LDAP Filters](how-to/discover-ldap-filters.md)**: How to find the right query filters for your directory.
- **[Manage API Keys](how-to/manage-api-keys.md)**: Create and revoke keys for agents.

#### Operations

- **[Configure Data Retention](how-to/configure-data-retention.md)**: Manage database cleanup policies.
- **[Manage OSV Vulnerabilities](how-to/manage-osv-vulnerabilities.md)**: Update, fetch, and rebuild OSV threat data.
- **[Search and Filter Assets](how-to/search-and-filter-assets.md)**: How to use the dashboard search and status
  filters.
- **[Run Database Migrations](how-to/run-migrations.md)**: Apply schema changes safely.
- **[Deploy to Kubernetes](how-to/deploy-kubernetes.md)**: Production deployment manifest.

#### Reports

- **[Detect Transaction Anomalies](how-to/detect-anomalies.md)**: Detect and manage unusual transactions.

#### Development

- **[Add a New Endpoint](how-to/add-endpoint.md)**: Workflow for contributors.
- **[Run Tests](how-to/run-tests.md)**: Execute the test suite.

### 3. Reference (Information-oriented)

_Technical descriptions and specifications._

#### System

- **[API Reference](reference/api-reference.md)**: High-level API overview.
- **[Database Schema](reference/database-schema.md)**: Tables, columns, and relationships.
- **[Environment Variables](reference/environment-variables.md)**: Complete configuration reference.

#### LDAP Specifics

- **[LDAP Configuration](reference/ldap-configuration.md)**: Every LDAP variable, plus the filters each directory
  server expects.
- **[LDAP Error Codes](reference/ldap-error-codes.md)**: Troubleshooting common error codes (32, 49, 50).

### 4. Explanation (Understanding-oriented)

_Background knowledge and design decisions._

#### Architecture

- **[System Architecture](explanation/architecture.md)**: High-level design, stack, and distributed scheduler.
- **[OSV Integration Details](explanation/osv-integration.md)**: How vulnerability fetching, payload batching, and
  scoring works.
- **[Data Model](explanation/data-model.md)**: Entities and relationships explanation.

#### Deep Dives

- **[LDAP Authentication](explanation/ldap-authentication.md)**: How the login flow works, and when a service account
  is actually required.
- **[Testing Strategy](explanation/testing-strategy.md)**: Overview of the test suite and coverage goals.

---

## �️ Tools & Scripts

The `scripts/` directory contains useful scripts for administrators:

- **[ldap-discovery.sh](../scripts/ldap-discovery.sh)**: An interactive script to help you discover your LDAP server's
  structure and test filters.

## 🔌 API Documentation (Swagger)

When the server is running, interactive API documentation is available at: `http://localhost:8080/swagger/index.html`

- **Source**: `docs/docs.go` (Generated from code comments)
- **Update**: Run `make doc` to regenerate.

---

## 🤝 Contributing

When adding new features, please update the relevant documentation sections above. Ensure all Markdown files pass
`markdownlint`.
