# Transaction Log Server

<!-- markdownlint-disable MD033 -->
<p align="center">
  <p align="center"><img width="100" height="100" src="https://raw.githubusercontent.com/txlog/.github/refs/heads/main/profile/logbook.png" alt="The Logo"></p>
  <p align="center"><strong>Server to receive data from Txlog Agent</strong></p>
  <p align="center">
    <a href="https://semver.org"><img src="https://img.shields.io/badge/SemVer-2.0.0-22bfda.svg" alt="SemVer Format"></a>
    <a href="https://github.com/txlog/.github/blob/main/profile/CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg" alt="Contributor Covenant"></a>
    <a href="https://newreleases.io/github/txlog/server"><img src="https://newreleases.io/badge.svg" alt="NewReleases"></a>
  </p>
</p>

This repository contains the code for the Txlog Server.

> [!WARNING]
> This repository is under active development and may introduce breaking changes at any time. Please note:
>
> - The codebase is evolving rapidly
> - Breaking changes may occur between commits
> - API stability is not guaranteed
> - Regular updates are recommended to stay current
> - Check the changelog before updating

## Installation

Use Docker to run this server.

```bash
docker pull cr.rda.run/txlog/server:main
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
  -e IGNORE_EMPTY_EXECUTION=true \
  cr.rda.run/txlog/server:main
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
        image: cr.rda.run/txlog/server:main
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
        - name: IGNORE_EMPTY_EXECUTION
          value: true
```

If you want to use a production (stable) version, replace `main` by the version
number (e.g. `v1.0`) in the Docker commands and Kubernetes configuration.

## 🪴 Project Activity

![Alt](https://repobeats.axiom.co/api/embed/e7072dd27ed7e95ffffdca0b6b8b1b9b8a9687ed.svg "Repobeats analytics image")

## Development

To make changes on this project, you need:

### Golang

```bash
wget https://go.dev/dl/go1.24.2.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc
source ~/.bashrc
rm go1.24.2.linux-amd64.tar.gz

go install github.com/swaggo/swag/cmd/swag@latest
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
IGNORE_EMPTY_EXECUTION=true
```

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
