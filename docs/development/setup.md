# Development Setup

This guide covers setting up a development environment for Waffles, including prerequisites, tooling, and development workflow.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Initial Setup](#initial-setup)
- [Development Tools](#development-tools)
- [IDE Configuration](#ide-configuration)
- [Development Workflow](#development-workflow)
- [Testing Setup](#testing-setup)
- [Debugging](#debugging)

## Prerequisites

### Required Software

#### Go
- **Version**: 1.21 or later
- **Installation**: 
  ```bash
  # macOS (Homebrew)
  brew install go
  
  # Linux (manual)
  curl -L https://go.dev/dl/go1.21.0.linux-amd64.tar.gz | sudo tar -xz -C /usr/local
  echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
  ```

#### Git
- **Version**: 2.20 or later
- **Configuration**:
  ```bash
  git config --global user.name "Your Name"
  git config --global user.email "your.email@example.com"
  ```

#### Make
- **Purpose**: Build automation
- **Installation**:
  ```bash
  # macOS
  xcode-select --install
  
  # Ubuntu/Debian
  sudo apt install build-essential
  
  # CentOS/RHEL
  sudo yum groupinstall "Development Tools"
  ```

#### Python 3.8+
- **Purpose**: Required for LLM CLI dependency
- **Installation**:
  ```bash
  # macOS
  brew install python
  
  # Linux
  sudo apt install python3 python3-pip
  ```

### Optional but Recommended

#### pipx
- **Purpose**: Isolated Python package installation
- **Installation**:
  ```bash
  pip install --user pipx
  pipx ensurepath
  ```

#### Docker
- **Purpose**: Containerized testing and building
- **Installation**: Follow [Docker documentation](https://docs.docker.com/get-docker/)

#### jq
- **Purpose**: JSON processing in scripts
- **Installation**:
  ```bash
  # macOS
  brew install jq
  
  # Linux
  sudo apt install jq
  ```

## Initial Setup

### 1. Clone Repository

```bash
# Clone the repository
git clone https://github.com/toozej/waffles.git
cd waffles

# Set up upstream remote if forking
git remote add upstream https://github.com/toozej/waffles.git
```

### 2. Install Dependencies

```bash
# Install Go module dependencies
go mod download

# Install development dependencies
make deps

# Install pre-commit hooks (optional)
make install-hooks
```

### 3. Build Project

```bash
# Build binary
make build

# Install to $GOPATH/bin
make install

# Verify installation
waffles version
```

### 4. Install Runtime Dependencies

```bash
# Auto-install all dependencies
waffles setup --auto-install

# Or install manually
make install-deps

# Verify dependencies
waffles deps check
```

### 5. Run Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests
make test-integration
```

## Development Tools

### Makefile Targets

The project includes a comprehensive Makefile with the following targets:

#### Building
```bash
make build          # Build binary for current platform
make build-all      # Build for all supported platforms
make install        # Install to $GOPATH/bin
make clean          # Clean build artifacts
```

#### Testing
```bash
make test           # Run unit tests
make test-all       # Run all tests including integration
make test-coverage  # Generate test coverage report
make test-race      # Run tests with race detection
make test-bench     # Run benchmark tests
```

#### Development
```bash
make dev            # Start development mode with auto-rebuild
make fmt            # Format code
make lint           # Run linter
make vet            # Run go vet
make check          # Run all checks (fmt, lint, vet, test)
```

#### Dependencies
```bash
make deps           # Install development dependencies
make deps-update    # Update dependencies
make install-deps   # Install runtime dependencies
make tidy           # Tidy go modules
```

#### Documentation
```bash
make docs           # Generate documentation
make docs-serve     # Serve documentation locally
```

### Go Tools Setup

#### Essential Tools
```bash
# Install Go tools
go install golang.org/x/tools/cmd/goimports@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/goreleaser/goreleaser@latest
```

#### Editor Integration
```bash
# VS Code Go extension
code --install-extension golang.go

# Vim-go (if using Vim)
# Follow vim-go installation instructions
```

## IDE Configuration

### Visual Studio Code

#### Recommended Extensions
- Go (golang.go)
- SQLite Viewer
- Git History
- GitLens
- Markdown All in One

#### Settings (`.vscode/settings.json`)
```json
{
  "go.toolsManagement.autoUpdate": true,
  "go.useLanguageServer": true,
  "go.lintOnSave": "package",
  "go.vetOnSave": "package",
  "go.testOnSave": false,
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "editor.formatOnSave": true,
  "files.associations": {
    "*.env": "properties"
  }
}
```

#### Tasks (`.vscode/tasks.json`)
```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Build",
      "type": "shell",
      "command": "make build",
      "group": "build",
      "problemMatcher": "$go"
    },
    {
      "label": "Test",
      "type": "shell", 
      "command": "make test",
      "group": "test"
    },
    {
      "label": "Test Current Package",
      "type": "shell",
      "command": "go test -v ./...",
      "options": {
        "cwd": "${fileDirname}"
      },
      "group": "test"
    }
  ]
}
```

### GoLand/IntelliJ IDEA

#### Configuration
1. Import project as Go module
2. Enable Go modules integration
3. Configure code style to use goimports
4. Set up run configurations for main commands

### Vim/Neovim

#### With vim-go
```vim
" Add to .vimrc
let g:go_fmt_command = "goimports"
let g:go_auto_type_info = 1
let g:go_def_mode='gopls'
let g:go_info_mode='gopls'
let g:go_metalinter_command='golangci-lint'

" Key mappings
au FileType go nmap <leader>r :GoRun<CR>
au FileType go nmap <leader>b :GoBuild<CR>
au FileType go nmap <leader>t :GoTest<CR>
```

## Development Workflow

### 1. Feature Development

```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Make changes
# ... edit files ...

# Run checks
make check

# Commit changes (use conventional commits)
git add .
git commit -m "feat: add new feature description"

# Push branch
git push origin feature/your-feature-name
```

### 2. Bug Fixes

```bash
# Create fix branch
git checkout -b fix/issue-description

# Make changes
# ... fix the bug ...

# Add test to prevent regression
# ... add test ...

# Run tests
make test

# Commit fix
git commit -m "fix: resolve issue description"
```

### 3. Testing Changes

```bash
# Run unit tests
make test

# Run integration tests
make test-integration  

# Test specific package
go test -v ./pkg/config

# Run with coverage
make test-coverage

# Race condition detection
make test-race
```

### 4. Building and Testing

```bash
# Build and test locally
make build
./bin/waffles version

# Test installation
make install
waffles --help

# Cross-platform build
make build-all
```

## Testing Setup

### Unit Testing

#### Test File Organization
- Test files named `*_test.go`
- Tests in same package as code under test
- Test helper functions in `testdata/` directories

#### Running Tests
```bash
# All tests
go test ./...

# Specific package
go test ./pkg/config

# Verbose output
go test -v ./pkg/config

# With coverage
go test -cover ./pkg/config
```

### Integration Testing

#### Database Tests
```bash
# Requires SQLite
sudo apt install sqlite3 libsqlite3-dev  # Linux
brew install sqlite                       # macOS

# Run database integration tests
go test -tags=integration ./pkg/logging
```

#### CLI Tests
```bash
# Test CLI commands
go test -v ./cmd/waffles

# Test with real dependencies (if installed)
go test -v -tags=integration ./...
```

### Benchmark Testing

```bash
# Run benchmarks
go test -bench=. ./pkg/repo

# With memory allocation stats
go test -bench=. -benchmem ./pkg/repo

# Profile CPU usage
go test -bench=. -cpuprofile=cpu.prof ./pkg/repo
```

## Debugging

### Local Debugging

#### Using Delve
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug main application
dlv debug ./cmd/waffles -- query "test prompt"

# Debug tests
dlv test ./pkg/config
```

#### Using VS Code Debugger

Create `.vscode/launch.json`:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Waffles",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/waffles",
      "args": ["query", "test prompt", "--verbose"]
    },
    {
      "name": "Debug Tests",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/pkg/config"
    }
  ]
}
```

### Logging and Tracing

#### Enable Debug Logging
```bash
export WAFFLES_VERBOSE=true
waffles query --verbose "debug test"
```

#### Trace Execution
```bash
# Go execution tracer
go run -trace=trace.out ./cmd/waffles query "test"
go tool trace trace.out
```

### Performance Profiling

#### CPU Profiling
```bash
# Profile a run
go run ./cmd/waffles query "test" -cpuprofile=cpu.prof

# Analyze profile
go tool pprof cpu.prof
```

#### Memory Profiling
```bash
# Profile memory usage
go run ./cmd/waffles query "test" -memprofile=mem.prof

# Analyze profile
go tool pprof mem.prof
```

## Environment Setup

### Development Environment Variables

Create a `.env.development` file:
```env
# Development settings
WAFFLES_VERBOSE=true
WAFFLES_DEFAULT_MODEL=gpt-3.5-turbo
WAFFLES_DB_PATH=./dev-waffles.db

# Test API keys (use test/mock keys)
OPENAI_API_KEY=sk-test-key-for-development
```

### Testing Environment
```bash
# Set up test environment
export WAFFLES_DB_PATH=":memory:"
export WAFFLES_VERBOSE=false
export WAFFLES_TIMEOUT_WHERESMYPROMPT=5
export WAFFLES_TIMEOUT_FILES2PROMPT=10
export WAFFLES_TIMEOUT_LLM=15
```

## Troubleshooting Development Issues

### Common Issues

#### Go Module Issues
```bash
# Clean module cache
go clean -modcache
go mod download
```

#### Build Issues
```bash
# Clean and rebuild
make clean
make build
```

#### Test Failures
```bash
# Run tests with more verbose output
go test -v -race ./...

# Check for race conditions
go test -race ./...
```

#### Database Issues
```bash
# Reset development database
rm -f dev-waffles.db
waffles setup
```

### Getting Help

1. Check existing [GitHub Issues](https://github.com/toozej/waffles/issues)
2. Look at [CI/CD logs](.github/workflows/) for examples
3. Ask in [GitHub Discussions](https://github.com/toozej/waffles/discussions)
4. Review existing code and tests for patterns

## Next Steps

Once your development environment is set up:

1. Read the [Code Organization](code-organization.md) guide
2. Review [Contributing Guidelines](contributing.md)
3. Check out the [Testing Guide](testing.md) for detailed testing practices
4. Explore the [API Documentation](api.md) for internal interfaces

Happy coding! 🚀