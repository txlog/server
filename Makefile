.PHONY: all clean build rpm

all:
	@echo "Usage: make [OPTION]"
	@echo ""
	@echo "Options:"
	@echo "  clean      Remove all artifacts"
	@echo "  build      Compile a binary"
	@echo "  rpm        Create the RPM package"

clean:
	@rm -rf bin/

build:
	@CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build -ldflags="-s -w" -trimpath -o bin/txlog-server

rpm:
	@nfpm pkg --packager rpm --target ./bin/
