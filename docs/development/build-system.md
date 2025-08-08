# Build System Guide

This document covers the build system, release processes, and distribution strategies for Waffles.

## Table of Contents

- [Build System Overview](#build-system-overview)
- [Build Configuration](#build-configuration)
- [Build Targets](#build-targets)
- [Cross-Platform Compilation](#cross-platform-compilation)
- [Release Process](#release-process)
- [Packaging and Distribution](#packaging-and-distribution)
- [Continuous Integration](#continuous-integration)
- [Docker Support](#docker-support)
- [Troubleshooting](#troubleshooting)

## Build System Overview

Waffles uses a combination of Go's native build tools, Make, and GitHub Actions to provide a comprehensive build system that supports:

- **Local development builds** for testing and debugging
- **Cross-platform compilation** for multiple OS and architecture combinations
- **Automated releases** with proper versioning and artifact generation
- **Package distribution** through multiple channels
- **Docker images** for containerized deployments

### Build Tools

- **Go toolchain** - Primary build system
- **Make** - Build automation and task runner
- **GitHub Actions** - CI/CD pipeline
- **GoReleaser** - Release automation and artifact generation
- **Docker** - Container image building

## Build Configuration

### Makefile Structure

The project uses a comprehensive Makefile for build automation:

```makefile
# Build configuration
BINARY_NAME=waffles
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(shell git rev-parse --short HEAD)

# Go configuration
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
CGO_ENABLED?=1

# Directories
BUILD_DIR=dist
COVERAGE_DIR=coverage

# Build flags
LDFLAGS=-ldflags="-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commit=$(COMMIT)"
BUILD_FLAGS=-trimpath $(LDFLAGS)

# Default target
.PHONY: all
all: clean deps build test

# Development build
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/waffles

# Install for local development
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(BUILD_FLAGS) ./cmd/waffles

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(COVERAGE_DIR)
	go clean -cache
	go clean -testcache

# Dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Verify dependencies
.PHONY: deps-verify
deps-verify:
	@echo "Verifying dependencies..."
	go mod verify

# Update dependencies
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy
```

### Version Management

Version information is embedded at build time:

```go
// cmd/waffles/version.go
package main

var (
    version   = "dev"     // Set by ldflags
    buildTime = "unknown" // Set by ldflags
    commit    = "unknown" // Set by ldflags
)

func printVersion() {
    fmt.Printf("Waffles %s\n", version)
    fmt.Printf("Built: %s\n", buildTime)
    fmt.Printf("Commit: %s\n", commit)
}
```

### Build Environment Variables

```bash
# Core build settings
export CGO_ENABLED=1              # Required for SQLite support
export GO111MODULE=on             # Enable Go modules
export GOPROXY=https://proxy.golang.org

# Cross-compilation settings
export GOOS=linux                 # Target OS
export GOARCH=amd64              # Target architecture

# Build optimization
export GOAMD64=v1                # AMD64 compatibility level
export GOEXPERIMENT=             # Experimental features
```

## Build Targets

### Development Builds

```makefile
# Quick development build
.PHONY: dev
dev:
	go build -o $(BUILD_DIR)/$(BINARY_NAME)-dev ./cmd/waffles

# Build with race detection
.PHONY: dev-race
dev-race:
	go build -race -o $(BUILD_DIR)/$(BINARY_NAME)-race ./cmd/waffles

# Debug build with symbols
.PHONY: debug
debug:
	go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME)-debug ./cmd/waffles
```

### Production Builds

```makefile
# Production build with optimizations
.PHONY: release
release: clean deps
	CGO_ENABLED=1 go build $(BUILD_FLAGS) -a -installsuffix cgo \
		-o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/waffles

# Static binary (where possible)
.PHONY: static
static: clean deps
	CGO_ENABLED=1 go build $(BUILD_FLAGS) \
		-ldflags="-linkmode external -extldflags -static" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-static ./cmd/waffles
```

### Testing Builds

```makefile
# Build test binaries
.PHONY: test-build
test-build:
	go test -c ./cmd/waffles
	go test -c ./pkg/...

# Build with test coverage
.PHONY: build-coverage
build-coverage:
	go build -cover -o $(BUILD_DIR)/$(BINARY_NAME)-coverage ./cmd/waffles
```

## Cross-Platform Compilation

### Supported Platforms

Waffles supports the following platform combinations:

| OS      | Architecture | CGO | Notes                    |
|---------|-------------|-----|--------------------------|
| linux   | amd64       | Yes | Primary development      |
| linux   | arm64       | Yes | ARM64 servers           |
| darwin  | amd64       | Yes | Intel Macs              |
| darwin  | arm64       | Yes | Apple Silicon Macs      |
| windows | amd64       | Yes | Windows 10/11           |
| freebsd | amd64       | Yes | FreeBSD support         |

### Cross-Compilation Setup

```bash
# Install cross-compilation toolchains
# For Linux targets
apt-get install gcc-multilib gcc-mingw-w64

# For macOS from Linux (using osxcross)
git clone https://github.com/tpoechtrager/osxcross
# Follow osxcross setup instructions
```

### Build Matrix

```makefile
# All supported platforms
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 freebsd/amd64

.PHONY: build-all
build-all: clean deps
	@echo "Building for all platforms..."
	@for platform in $(PLATFORMS); do \
		OS=$${platform%/*}; \
		ARCH=$${platform#*/}; \
		echo "Building for $$OS/$$ARCH..."; \
		GOOS=$$OS GOARCH=$$ARCH CGO_ENABLED=1 \
			go build $(BUILD_FLAGS) \
			-o $(BUILD_DIR)/$(BINARY_NAME)-$$OS-$$ARCH \
			./cmd/waffles; \
	done

# Windows-specific build
.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
		go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).exe ./cmd/waffles
```

### CGO Cross-Compilation

Since Waffles uses SQLite (requires CGO), cross-compilation needs specific setup:

```makefile
# Linux ARM64 with CGO
.PHONY: build-linux-arm64
build-linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc \
		go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/waffles

# macOS with CGO (from Linux)
.PHONY: build-darwin
build-darwin:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 CC=o64-clang \
		go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/waffles
```

## Release Process

### Semantic Versioning

Waffles follows [Semantic Versioning](https://semver.org/):

- **MAJOR** (v1.0.0) - Incompatible API changes
- **MINOR** (v0.1.0) - Backward-compatible functionality additions
- **PATCH** (v0.0.1) - Backward-compatible bug fixes

### Release Workflow

```bash
# 1. Create release branch
git checkout -b release/v1.2.0

# 2. Update version files
echo "v1.2.0" > VERSION
git add VERSION
git commit -m "chore: bump version to v1.2.0"

# 3. Update changelog
vim CHANGELOG.md
git add CHANGELOG.md
git commit -m "docs: update changelog for v1.2.0"

# 4. Create tag
git tag -a v1.2.0 -m "Release v1.2.0"

# 5. Push tag (triggers release automation)
git push origin v1.2.0
```

### GoReleaser Configuration

```yaml
# .goreleaser.yml
version: 1

project_name: waffles

before:
  hooks:
    - go mod tidy
    - go mod download

builds:
  - id: waffles
    env:
      - CGO_ENABLED=1
    main: ./cmd/waffles
    binary: waffles
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.buildTime={{.Date}}
      - -X main.commit={{.Commit}}
    goos:
      - linux
      - darwin
      - windows
      - freebsd
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64

archives:
  - id: waffles
    builds:
      - waffles
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^style:'

release:
  github:
    owner: toozej
    name: waffles
  draft: false
  prerelease: auto

brews:
  - tap:
      owner: toozej
      name: homebrew-tap
    homepage: https://github.com/toozej/waffles
    description: "LLM toolchain orchestration CLI"
    license: MIT
    install: |
      bin.install "waffles"
    test: |
      system "#{bin}/waffles version"

nfpms:
  - id: packages
    package_name: waffles
    vendor: toozej
    homepage: https://github.com/toozej/waffles
    maintainer: James Toozej <james@toozej.com>
    description: "LLM toolchain orchestration CLI"
    license: MIT
    formats:
      - deb
      - rpm
      - apk
```

### Manual Release Process

```makefile
.PHONY: release-local
release-local: clean deps test
	@echo "Building release artifacts..."
	goreleaser release --snapshot --rm-dist
	@echo "Release artifacts built in dist/"

.PHONY: release-test
release-test: clean deps test
	@echo "Testing release process..."
	goreleaser release --skip-publish --rm-dist

.PHONY: release-publish
release-publish: clean deps test
	@echo "Publishing release..."
	goreleaser release --rm-dist
```

## Packaging and Distribution

### Package Formats

#### Homebrew

```ruby
# homebrew/waffles.rb
class Waffles < Formula
  desc "LLM toolchain orchestration CLI"
  homepage "https://github.com/toozej/waffles"
  url "https://github.com/toozej/waffles/archive/v1.0.0.tar.gz"
  sha256 "..."
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/waffles"
  end

  test do
    system "#{bin}/waffles", "version"
  end
end
```

#### Debian Package

```makefile
.PHONY: deb
deb: build-linux-amd64
	@echo "Building Debian package..."
	mkdir -p packaging/deb/usr/bin
	cp $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 packaging/deb/usr/bin/$(BINARY_NAME)
	dpkg-deb --build packaging/deb $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_amd64.deb
```

#### RPM Package

```spec
# packaging/rpm/waffles.spec
Name:           waffles
Version:        %{_version}
Release:        1%{?dist}
Summary:        LLM toolchain orchestration CLI

License:        MIT
URL:            https://github.com/toozej/waffles
Source0:        %{name}-%{version}.tar.gz

%description
Waffles is a command-line tool that orchestrates an LLM toolchain
for local development workflows.

%prep
%setup -q

%build
make build

%install
mkdir -p %{buildroot}%{_bindir}
cp dist/%{name} %{buildroot}%{_bindir}/

%files
%{_bindir}/%{name}

%changelog
* Wed Jan 15 2024 James Toozej <james@toozej.com> - 1.0.0-1
- Initial package
```

### Distribution Channels

#### GitHub Releases

Automated through GitHub Actions:

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0
    
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - uses: goreleaser/goreleaser-action@v4
      with:
        distribution: goreleaser
        version: latest
        args: release --rm-dist
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

#### Package Managers

```makefile
# Homebrew tap update
.PHONY: homebrew-update
homebrew-update:
	@echo "Updating Homebrew formula..."
	cd homebrew-tap && \
	git pull && \
	brew bump-formula-pr --url=$(ARCHIVE_URL) --sha256=$(SHA256) waffles

# Arch AUR package
.PHONY: aur-update
aur-update:
	@echo "Updating AUR package..."
	cd aur-waffles && \
	makepkg --printsrcinfo > .SRCINFO && \
	git add . && \
	git commit -m "Update to $(VERSION)" && \
	git push
```

## Continuous Integration

### GitHub Actions Workflows

#### Build and Test

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go-version: [1.21, 1.22]
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Cache dependencies
      uses: actions/cache@v3
      with:
        path: |
          ~/.cache/go-build
          ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    
    - name: Install dependencies
      run: make deps
    
    - name: Build
      run: make build
    
    - name: Test
      run: make test-all
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
```

#### Security Scanning

```yaml
# .github/workflows/security.yml
name: Security

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Run Gosec
      uses: securecodewarrior/github-action-gosec@master
      with:
        args: './...'
    
    - name: Run Nancy (dependency check)
      run: |
        go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth
```

### Build Optimization

```makefile
# Parallel builds
.PHONY: build-parallel
build-parallel:
	@echo "Building in parallel..."
	$(MAKE) -j$(shell nproc) build-all

# Build cache optimization
.PHONY: build-cached
build-cached:
	@echo "Building with cache optimization..."
	GOCACHE=$(PWD)/.cache/go-build go build $(BUILD_FLAGS) ./cmd/waffles
```

## Docker Support

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o waffles ./cmd/waffles

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite

# Create non-root user
RUN addgroup -g 1001 waffles && \
    adduser -D -s /bin/sh -u 1001 -G waffles waffles

WORKDIR /home/waffles

COPY --from=builder /app/waffles /usr/local/bin/waffles

USER waffles

ENTRYPOINT ["waffles"]
CMD ["--help"]
```

### Multi-stage Build

```dockerfile
# Multi-arch build
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-s -w" -o waffles ./cmd/waffles

FROM alpine:latest
# ... runtime setup
```

### Docker Compose Development

```yaml
# docker-compose.yml
version: '3.8'

services:
  waffles-dev:
    build:
      context: .
      target: builder
    volumes:
      - .:/app
      - go-mod-cache:/go/pkg/mod
      - go-build-cache:/root/.cache/go-build
    working_dir: /app
    command: go run ./cmd/waffles
    environment:
      - WAFFLES_LOG_LEVEL=debug

volumes:
  go-mod-cache:
  go-build-cache:
```

## Troubleshooting

### Common Build Issues

#### CGO Compilation Errors

```bash
# Error: CGO not enabled
export CGO_ENABLED=1

# Error: Missing C compiler
# On Ubuntu/Debian
sudo apt-get install gcc libc6-dev

# On macOS
xcode-select --install

# On Windows (using MinGW)
choco install mingw
```

#### SQLite Build Issues

```bash
# Error: sqlite3.h not found
# On Ubuntu/Debian
sudo apt-get install libsqlite3-dev

# On CentOS/RHEL
sudo yum install sqlite-devel

# On Alpine
apk add sqlite-dev
```

#### Cross-compilation Issues

```bash
# Error: Missing cross-compiler
# Install cross-compilation tools
sudo apt-get install gcc-multilib gcc-mingw-w64

# Error: Missing CGO cross-compiler
# Set appropriate CC variable
export CC=x86_64-w64-mingw32-gcc  # For Windows
export CC=aarch64-linux-gnu-gcc   # For Linux ARM64
```

### Debug Build Issues

```makefile
.PHONY: debug-build
debug-build:
	@echo "Go version: $(shell go version)"
	@echo "CGO enabled: $(CGO_ENABLED)"
	@echo "GOOS: $(GOOS)"
	@echo "GOARCH: $(GOARCH)"
	@echo "Build flags: $(BUILD_FLAGS)"
	go build -v $(BUILD_FLAGS) ./cmd/waffles
```

### Performance Optimization

```makefile
# Profile-guided optimization (PGO)
.PHONY: build-pgo
build-pgo:
	# Build with profiling
	go build -pgo=auto $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-pgo ./cmd/waffles

# Link-time optimization
.PHONY: build-lto
build-lto:
	go build -ldflags="-s -w -linkmode external -extldflags '-flto'" \
		$(BUILD_FLAGS) ./cmd/waffles
```

### Build Verification

```makefile
.PHONY: verify-build
verify-build: build
	@echo "Verifying build..."
	@file $(BUILD_DIR)/$(BINARY_NAME)
	@$(BUILD_DIR)/$(BINARY_NAME) version
	@$(BUILD_DIR)/$(BINARY_NAME) --help > /dev/null
	@echo "Build verification passed"
```

## Conclusion

The Waffles build system provides comprehensive support for development, testing, and distribution across multiple platforms. Key features include:

- **Cross-platform compilation** with CGO support
- **Automated releases** with GoReleaser
- **Multiple distribution channels** (GitHub, Homebrew, package managers)
- **Docker support** with multi-stage builds
- **Comprehensive CI/CD** with GitHub Actions

For build-related issues or questions, refer to the [troubleshooting section](#troubleshooting) or consult the [contributing guide](contributing.md).