.PHONY: all
all:
	@echo "Usage: make [OPTION]"
	@echo ""
	@echo "Options:"
	@echo "  clean      Remove all artifacts"
	@echo "  fmt        Recursively format all packages"
	@echo "  vet        Recursively check all packages"
	@echo "  doc        Write the swagger documentation based on method comments"
	@echo "  build      Compile a binary"
	@echo "  run        Run the server code"

.PHONY: clean
clean:
	@rm -rf bin/

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: vet
vet:
	@go vet ./...

.PHONY: build
build:
		@CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build -ldflags="-s -w -extldflags=-static" -trimpath -o bin/txlog-server

.PHONY: run
run:
	@air

.PHONY: doc
doc:
	@~/go/bin/swag init --outputTypes go
	@~/go/bin/swag fmt
