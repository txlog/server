# Transaction Log Server

<!-- markdownlint-disable MD033 MD013 -->
<p align="center">
  <p align="center"><img width="100" height="100" src="https://raw.githubusercontent.com/txlog/.github/refs/heads/main/profile/logbook.png" alt="The Logo"></p>
  <p align="center"><strong>Server to receive data from Txlog Agent</strong></p>
  <p align="center">
    <a href="https://semver.org"><img src="https://img.shields.io/badge/SemVer-2.0.0-22bfda.svg" alt="SemVer Format"></a>
    <a href="./CHANGELOG.md"><img src="https://img.shields.io/badge/changelog-Keep_a_Changelog_v1.1.0-E05735" alt="Keep a Changelog"></a>
    <a href="https://github.com/txlog/.github/blob/main/profile/CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg" alt="Contributor Covenant"></a>
    <a href="https://newreleases.io/github/txlog/server"><img src="https://newreleases.io/badge.svg" alt="NewReleases"></a>
    <a href="https://deepwiki.com/txlog/server"><img src="https://deepwiki.com/badge.svg" alt="Ask DeepWiki"></a>
  </p>
</p>
<!-- markdownlint-enable MD033 MD013 -->

This repository contains the code for the Txlog Server.

## Installation

Use Docker to run this server.

```bash
docker pull ghcr.io/txlog/server:main
```

Run the server.

```bash
docker run -d -p 8080:8080 \
  -e INSTANCE=Datacenter 001 \
  -e LOG_LEVEL=INFO \
  -e PGSQL_HOST=postgres.example.com \
  -e PGSQL_PORT=5432 \
  -e PGSQL_USER=txlog \
  -e PGSQL_DB=txlog \
  -e PGSQL_PASSWORD=your_db_password \
  -e PGSQL_SSLMODE=require \
  -e CRON_RETENTION_DAYS=7 \
  -e CRON_RETENTION_EXPRESSION=0 2 * * * \
  -e CRON_STATS_EXPRESSION=0 * * * * \
  ghcr.io/txlog/server:main
```

Or use it on your Kubernetes cluster

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: txlog-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: txlog-server
  template:
    metadata:
      labels:
        app: txlog-server
    spec:
      containers:
      - name: txlog-server
        image: ghcr.io/txlog/server:main
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        env:
        - name: INSTANCE
          value: "Datacenter 001"
        - name: LOG_LEVEL
          value: "INFO"
        - name: PGSQL_HOST
          value: "postgres.example.com"
        - name: PGSQL_PORT
          value: "5432"
        - name: PGSQL_USER
          value: "txlog"
        - name: PGSQL_DB
          value: "txlog"
        - name: PGSQL_PASSWORD
          valueFrom:
            secretKeyRef:
              name: txlog-secrets
              key: db-password
        - name: PGSQL_SSLMODE
          value: "require"
        - name: CRON_RETENTION_DAYS
          value: 7
        - name: CRON_RETENTION_EXPRESSION
          value: 0 2 * * *
        - name: CRON_STATS_EXPRESSION
          value: 0 * * * *
```

If you want to use a production (stable) version, replace `main` by the version
number (e.g. `v1.0`) in the Docker commands and Kubernetes configuration.

## 🪴 Project Activity

![Alt](https://repobeats.axiom.co/api/embed/e7072dd27ed7e95ffffdca0b6b8b1b9b8a9687ed.svg "Repobeats analytics image")

## Development

To make changes on this project, you need:

### Golang

```bash
wget https://go.dev/dl/go1.25.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.5.linux-amd64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc
source ~/.bashrc
rm go1.25.5.linux-amd64.tar.gz
```

### Swaggo

```bash
curl -fsSL https://install.rda.run/swaggo/swag@latest! | bash
```

### Air

```bash
curl -fsSL https://install.rda.run/air-verse/air@latest! | bash
```

### A `.env` file

```bash
INSTANCE=Development environment
LOG_LEVEL=DEBUG
GIN_MODE=debug
PGSQL_HOST=127.0.0.1
PGSQL_PORT=5432
PGSQL_USER=postgres
PGSQL_DB=txlog
PGSQL_PASSWORD=your_db_password
PGSQL_SSLMODE=require
CRON_RETENTION_DAYS=1
CRON_RETENTION_EXPRESSION=0 2 * * *
CRON_STATS_EXPRESSION=0 * * * *

# OIDC Authentication (Optional)
OIDC_ISSUER_URL=https://id.example.com
OIDC_CLIENT_ID=your_oidc_client_id
OIDC_CLIENT_SECRET=your_oidc_client_secret
OIDC_REDIRECT_URL=https://txlog.example.com/auth/callback
OIDC_SKIP_TLS_VERIFY=false

# LDAP Authentication (Optional)
LDAP_HOST=ldap.example.com
LDAP_PORT=389
LDAP_USE_TLS=false
LDAP_SKIP_TLS_VERIFY=false
LDAP_BIND_DN=cn=admin,dc=example,dc=com
LDAP_BIND_PASSWORD=your_bind_password
LDAP_BASE_DN=ou=users,dc=example,dc=com
LDAP_USER_FILTER=(uid=%s)
LDAP_ADMIN_GROUP=cn=admins,ou=groups,dc=example,dc=com
LDAP_VIEWER_GROUP=cn=viewers,ou=groups,dc=example,dc=com
LDAP_GROUP_FILTER=(member=%s)
```

#### Authentication Configuration

Txlog Server supports three authentication modes:

1. **No Authentication** (Default): If neither OIDC nor LDAP is configured, the
   server runs without authentication
2. **OIDC Authentication**: Configure OIDC environment variables to enable
   OpenID Connect authentication
3. **LDAP Authentication**: Configure LDAP environment variables to enable LDAP
   authentication
4. **Both OIDC and LDAP**: Both can be enabled simultaneously, allowing users to
   choose their preferred authentication method

##### LDAP Configuration Details

- **LDAP_HOST** (Required): LDAP server hostname
- **LDAP_PORT**: LDAP server port (default: 389 for LDAP, 636 for LDAPS)
- **LDAP_USE_TLS**: Enable TLS connection (default: false)
- **LDAP_SKIP_TLS_VERIFY**: Skip TLS certificate verification for self-signed
  certificates (default: false)
- **LDAP_BIND_DN** (Optional): Distinguished Name for service account bind
  - **Not required** if your LDAP allows anonymous searches
  - **Recommended** for Active Directory and restricted LDAP servers
- **LDAP_BIND_PASSWORD** (Optional): Password for service account
- **LDAP_BASE_DN** (Required): Base DN for user searches
- **LDAP_USER_FILTER**: LDAP filter for finding users (default: `(uid=%s)`,
  where %s is replaced with username)
- **LDAP_ADMIN_GROUP**: DN of the admin group (users in this group have full
  admin access)
- **LDAP_VIEWER_GROUP**: DN of the viewer group (users in this group have
  read-only access)
- **LDAP_GROUP_FILTER**: LDAP filter for checking group membership (default:
  `(member=%s)`, where %s is replaced with user DN)

**Note**: At least one of `LDAP_ADMIN_GROUP` or `LDAP_VIEWER_GROUP` must be
configured. Users must be members of at least one of these groups to
authenticate successfully.

**Service Account**: `LDAP_BIND_DN` and `LDAP_BIND_PASSWORD` are **optional**.
If not provided, the server will:

- Use anonymous bind for user searches (works with OpenLDAP)
- Use the authenticated user's session for group membership checks
- Active Directory typically **requires** a service account

### Development commands

The `Makefile` contains all the necessary commands for development. You can run
`make` to view all options.

To create the binary and distribute

- `make clean`: remove compiled binaries and packages
- `make run`: execute the server code
- `make build`: build a production-ready binary on `./bin` directory
- `make doc`: write the swagger documentation based on method comments

The server will run at <http://localhost:8080> and the Swagger docs at
<http://localhost:8080/swagger/index.html>.

## Contributing

1. Fork it (<https://github.com/txlog/server/fork>)
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request
