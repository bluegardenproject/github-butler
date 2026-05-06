.PHONY: setup build build-all release clean test fmt vet tidy

BINARY_NAME := github-butler
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BUILD_TIME  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS     := -s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)

# `setup` wires the in-repo git hooks. It is idempotent and is a prerequisite
# of `build` and `test`, so any common dev command re-asserts the hook config.
setup:
	@git config core.hooksPath .githooks
	@chmod +x .githooks/* scripts/*.sh 2>/dev/null || true
	@echo "hooks active: core.hooksPath = $$(git config core.hooksPath)"

build: setup
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@go build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) .
	@echo "Built ./$(BINARY_NAME)"

# Cross-compile to dist/ for the 5 release targets. CGO is off so the
# resulting binaries are fully static and don't depend on libc on the
# host they're installed to.
build-all:
	@echo "Building $(BINARY_NAME) $(VERSION) for all platforms..."
	@mkdir -p dist
	@CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-amd64 .
	@CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-arm64 .
	@CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-amd64 .
	@CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-arm64 .
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Built all platform binaries in dist/"

release: clean build-all
	@echo "Release artifacts ready in dist/:"
	@ls -la dist/

clean:
	@rm -rf dist/ $(BINARY_NAME)
	@echo "Cleaned build artifacts"

test: setup
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy
