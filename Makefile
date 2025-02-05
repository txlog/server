.PHONY: all clean build run

all:
	@echo "Usage: make [OPTION]"
	@echo ""
	@echo "Options:"
	@echo "  clean      Remove all artifacts"
	@echo "  run        Run the server code"
	@echo "  build      Compile a binary"

clean:
	@rm -rf bin/

build:
	@CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build -ldflags="-s -w" -trimpath -o bin/txlog-server

run:
	@go run main.go
