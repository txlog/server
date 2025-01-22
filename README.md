# Transaction Log Server

<!-- markdownlint-disable MD033 -->
<p align="center">
  <p align="center"><img width="100" height="100" src="https://raw.githubusercontent.com/txlog/.github/refs/heads/main/profile/logbook.png" alt="The Logo"></p>
  <p align="center"><strong>Server to receive data from TxLog Agent</strong></p>
  <p align="center">
    <a href="https://semver.org"><img src="https://img.shields.io/badge/SemVer-2.0.0-22bfda.svg" alt="SemVer Format"></a>
    <a href="https://github.com/txlog/.github/blob/main/profile/CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg" alt="Contributor Covenant"></a>
  </p>
</p>

This repository contains the code for the TxLog Server.

## Installation

```bash
sudo dnf localinstall -y https://rpm.rda.run/rpm-rda-run-1.0-1.noarch.rpm
sudo dnf install -y txlog-server
```

## Development

To make changes on this project, you need:

### Golang

```bash
sudo dnf install -y go
```

### nFPM

```bash
echo '[goreleaser]
name=GoReleaser
baseurl=https://repo.goreleaser.com/yum/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/goreleaser.repo
sudo yum install -y nfpm
```

### A `.env` file

```bash
HOST=postgresql.host
PORT=5432
USER=txlog
DB_NAME=txlog
PASSWORD=txlog-password
```

### Development commands

The `Makefile` contains all the necessary commands for development. You can run
`make` to view all options.

To create the binary and distribute

* `make clean`: remove compiled binaries and packages
* `make build`: build a production-ready binary on `./bin` directory
* `make rpm`: create new `.rpm` package

## Contributing

1. Fork it (<https://github.com/txlog/server/fork>)
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request
