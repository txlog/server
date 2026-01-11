.PHONY: all help clean fmt vet build run doc

all: help

VERSION := $(shell cat .version)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

## help: You know what this target does
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/ /'

## clean: Remove all artifacts
clean:
	@rm -rf bin/

## fmt: Recursively format all packages
fmt:
	@go fmt ./...

## vet: Recursively check all packages
vet:
	@go vet ./...

## build: Compile a binary
build:
	@CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build -ldflags="-s -w -extldflags=-static -X 'github.com/txlog/server/version.SemVer=$(VERSION)'" -trimpath -o bin/txlog-server

## run: Run the server code
run:
	@air

## doc: Write the swagger documentation based on method comments
doc:
	@~/go/bin/swag init --outputTypes go
	@~/go/bin/swag fmt
