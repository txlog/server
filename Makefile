.PHONY: all clean build run doc

all:
	@echo "Usage: make [OPTION]"
	@echo ""
	@echo "Options:"
	@echo "  clean      Remove all artifacts"
	@echo "  run        Run the server code"
	@echo "  build      Compile a binary"
	@echo "  doc        Write the swagger documentation based on method comments"

clean:
	@rm -rf bin/

build:
	@CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build -ldflags="-s -w" -trimpath -o bin/txlog-server

run:
	@go run main.go

doc:
	@~/go/bin/swag init --outputTypes go
	@~/go/bin/swag fmt
