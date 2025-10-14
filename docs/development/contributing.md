# Contributing Guidelines

This document outlines how to contribute to Waffles, including development workflow, code standards, and review process.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Standards](#code-standards)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Issue Guidelines](#issue-guidelines)
- [Code Review](#code-review)
- [Release Process](#release-process)

## Getting Started

### Prerequisites

Before contributing, ensure you have:

1. **Development environment set up** - See [Development Setup](setup.md)
2. **Understanding of the architecture** - Read [Architecture Overview](architecture.md)
3. **Familiarity with code organization** - Review [Code Organization](code-organization.md)

### First Contribution

1. **Fork the repository**
   ```bash
   # Fork on GitHub UI, then clone your fork
   git clone https://github.com/YOUR_USERNAME/waffles.git
   cd waffles
   git remote add upstream https://github.com/toozej/waffles.git
   ```

2. **Find an issue to work on**
   - Look for issues labeled `good first issue` or `help wanted`
   - Check the project board for prioritized work
   - Ask in discussions if you need guidance

3. **Set up development environment**
   ```bash
   make deps
   make build
   make test
   ```

## Development Workflow

### Branch Naming

Use descriptive branch names with prefixes:

- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation changes
- `refactor/description` - Code refactoring
- `test/description` - Test improvements

Examples:
```bash
feature/add-prometheus-metrics
fix/database-connection-leak
docs/improve-setup-guide
refactor/simplify-config-loading
```

### Commit Message Format

We use [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation only changes
- `style` - Formatting, missing semicolons, etc.
- `refactor` - Code change that neither fixes a bug nor adds a feature
- `test` - Adding missing tests or correcting existing tests
- `chore` - Changes to build process or auxiliary tools

**Examples:**
```bash
feat(pipeline): add timeout configuration for external tools
fix(deps): resolve goroutine leak in installer
docs(api): add examples to configuration documentation
refactor(repo): simplify language detection logic
test(logging): add integration tests for database operations
```

### Development Process

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Follow existing code patterns
   - Add tests for new functionality
   - Update documentation if needed

3. **Test your changes**
   ```bash
   # Run all tests
   make test-all
   
   # Check code quality
   make check
   
   # Test the build
   make build
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

5. **Keep your branch updated**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

6. **Push and create PR**
   ```bash
   git push origin feature/your-feature-name
   # Create PR via GitHub UI
   ```

## Code Standards

### Go Style Guidelines

Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and these additional standards:

#### Formatting
- Use `gofmt` or `goimports` for consistent formatting
- Maximum line length: 100 characters (soft limit)
- Use tabs for indentation

#### Naming
```go
// Good: Clear, descriptive names
func LoadConfiguration() (*Config, error)
type DatabaseConnection struct{}
var ErrInvalidInput = errors.New("invalid input")

// Avoid: Abbreviations and unclear names
func LoadCfg() (*Cfg, error)
type DBConn struct{}
var ErrInp = errors.New("invalid input")
```

#### Error Handling
```go
// Good: Wrap errors with context
if err := processFile(filename); err != nil {
    return fmt.Errorf("failed to process file %s: %w", filename, err)
}

// Good: Custom error types for specific cases
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}
```

#### Function Design
```go
// Good: Single responsibility, clear parameters
func ValidateConfig(cfg *Config) error {
    if cfg.DefaultModel == "" {
        return ErrMissingModel
    }
    return nil
}

// Good: Interface parameters, concrete returns
func ProcessQuery(logger Logger, query string) (*Result, error) {
    // implementation
}
```

### Package Structure

#### Interface Definition
- Define interfaces at point of use, not implementation
- Keep interfaces small and focused
- Use descriptive interface names

```go
// Good: Small, focused interface
type ConfigValidator interface {
    Validate(cfg *Config) error
}

// Good: Define where used, not where implemented
type QueryProcessor struct {
    validator ConfigValidator  // Use interface
    logger    *Logger         // Concrete type for return values
}
```

#### Dependency Management
```go
// Good: Dependency injection
func NewQueryProcessor(validator ConfigValidator, logger *Logger) *QueryProcessor {
    return &QueryProcessor{
        validator: validator,
        logger:    logger,
    }
}

// Avoid: Global state
var globalLogger *Logger
```

### Documentation Standards

#### Package Documentation
```go
// Package pipeline provides orchestration functionality for executing
// the Waffles toolchain. It coordinates the execution of wheresmyprompt,
// files2prompt, and llm in sequence while handling errors, timeouts,
// and logging.
//
// The main entry point is the Pipeline type, which can be configured
// with different options and then executed with a query string.
//
// Example usage:
//
//     cfg := config.LoadConfig()
//     pipeline := pipeline.NewPipeline(cfg)
//     result, err := pipeline.Execute("analyze this code")
//
package pipeline
```

#### Function Documentation
```go
// Execute runs the complete Waffles pipeline with the given query.
// It coordinates the execution of wheresmyprompt, files2prompt, and llm
// in sequence, handling timeouts and error recovery.
//
// The function returns an ExecutionContext containing detailed information
// about each step of the execution, including timing, outputs, and any
// errors encountered.
//
// Example:
//
//     ctx, err := pipeline.Execute("review this code")
//     if err != nil {
//         log.Printf("Pipeline failed: %v", err)
//         return
//     }
//     fmt.Println("Result:", ctx.FinalOutput)
//
func (p *Pipeline) Execute(query string) (*ExecutionContext, error) {
    // implementation
}
```

### Configuration Management

#### Environment Variables
- All configuration should be environment-variable based
- Use `WAFFLES_` prefix for all variables
- Provide sensible defaults
- Document all configuration options

```go
const (
    DefaultModel    = "claude-3-sonnet"
    DefaultProvider = "anthropic"
    DefaultTimeout  = 300 // seconds
)

type Config struct {
    DefaultModel    string `env:"WAFFLES_DEFAULT_MODEL"`
    DefaultProvider string `env:"WAFFLES_DEFAULT_PROVIDER"`
    Timeout         int    `env:"WAFFLES_TIMEOUT"`
}
```

## Testing Requirements

### Test Coverage

- **Minimum coverage**: 80% overall, 90% for critical paths
- All new features must include tests
- Bug fixes must include regression tests
- Public APIs must have example tests

### Test Organization

#### Unit Tests
```go
func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  *Config
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid config",
            config:  &Config{DefaultModel: "gpt-4"},
            wantErr: false,
        },
        {
            name:    "missing model",
            config:  &Config{},
            wantErr: true,
            errMsg:  "default model not specified",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if tt.wantErr {
                if err == nil {
                    t.Errorf("expected error but got none")
                }
                if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
                    t.Errorf("expected error message %q, got %q", tt.errMsg, err.Error())
                }
            } else if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}
```

#### Integration Tests
```go
func TestPipelineIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Set up test environment
    tempDir := t.TempDir()
    cfg := &Config{
        DefaultModel: "test-model",
        LogDBPath:    filepath.Join(tempDir, "test.db"),
    }

    pipeline := NewPipeline(cfg)
    ctx, err := pipeline.Execute("test query")
    
    if err != nil {
        t.Fatalf("Pipeline execution failed: %v", err)
    }

    if ctx.Success != true {
        t.Errorf("Expected successful execution, got failure")
    }
}
```

#### Test Helpers
```go
// testhelpers.go
func setupTestConfig(t *testing.T) *Config {
    t.Helper()
    return &Config{
        DefaultModel:    "test-model",
        DefaultProvider: "test-provider",
        LogDBPath:       ":memory:",
    }
}

func setupTestDatabase(t *testing.T) *Database {
    t.Helper()
    db, err := NewDatabase(":memory:")
    if err != nil {
        t.Fatalf("Failed to create test database: %v", err)
    }
    t.Cleanup(func() { db.Close() })
    return db
}
```

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Integration tests only  
make test-integration

# Specific package
go test -v ./pkg/config

# Race condition detection
go test -race ./...

# Benchmarks
go test -bench=. ./pkg/repo
```

## Pull Request Process

### Before Opening a PR

1. **Ensure tests pass**
   ```bash
   make test-all
   make check
   ```

2. **Update documentation**
   - Add/update function documentation
   - Update user documentation if needed
   - Add changelog entry for significant changes

3. **Rebase on main**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

### PR Template

When opening a PR, include:

```markdown
## Description
Brief description of changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)  
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated and passing
- [ ] No new warnings or errors introduced
```

### PR Guidelines

- **Keep PRs focused** - One logical change per PR
- **Write clear descriptions** - Explain what and why, not just what
- **Include tests** - All functionality should be tested
- **Update documentation** - Keep docs in sync with code
- **Respond to feedback** - Address review comments promptly

## Issue Guidelines

### Reporting Bugs

Use the bug report template:

```markdown
**Describe the bug**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Run command '...'
2. With configuration '...'
3. See error

**Expected behavior**
What you expected to happen.

**Environment:**
- OS: [e.g. macOS 12.0]
- Waffles version: [e.g. v1.0.0]
- Go version: [e.g. 1.21]

**Additional context**
Add any other context about the problem.
```

### Feature Requests

Use the feature request template:

```markdown
**Is your feature request related to a problem?**
A clear description of what the problem is.

**Describe the solution you'd like**
A clear description of what you want to happen.

**Describe alternatives you've considered**
Alternative solutions or features you've considered.

**Use cases**
Specific use cases where this feature would be helpful.

**Additional context**
Add any other context or screenshots about the feature request.
```

## Code Review

### Review Process

1. **Automated checks** - CI must pass before human review
2. **Self-review** - Author reviews own PR first
3. **Peer review** - At least one maintainer review required
4. **Address feedback** - Make requested changes
5. **Final approval** - Maintainer approves and merges

### Review Checklist

#### Functionality
- [ ] Code does what it claims to do
- [ ] Edge cases are handled appropriately
- [ ] Error conditions are handled gracefully
- [ ] Performance considerations addressed

#### Code Quality
- [ ] Code is readable and well-structured
- [ ] Functions are focused and not too complex
- [ ] Naming is clear and consistent
- [ ] No unnecessary complexity

#### Testing
- [ ] Adequate test coverage
- [ ] Tests are meaningful and well-structured
- [ ] Integration tests where appropriate
- [ ] No flaky tests introduced

#### Documentation
- [ ] Public APIs are documented
- [ ] Complex logic is commented
- [ ] User-facing changes documented
- [ ] Breaking changes noted

### Providing Feedback

#### Giving Good Feedback
- Be specific and actionable
- Explain the reasoning behind suggestions
- Distinguish between must-fix and suggestions
- Be respectful and constructive

```markdown
# Good feedback
This function is doing too many things. Consider splitting the validation 
logic into a separate function to improve readability and testability.

# Better feedback with example
This function is doing too many things. Consider splitting like this:

```go
func ProcessConfig(cfg *Config) error {
    if err := validateConfig(cfg); err != nil {
        return err
    }
    return applyConfig(cfg)
}
```

This would improve readability and make the validation logic testable in isolation.
```

#### Receiving Feedback
- Assume positive intent
- Ask for clarification if feedback is unclear
- Address feedback promptly
- Thank reviewers for their time

## Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backward-compatible functionality additions
- **PATCH** version for backward-compatible bug fixes

### Release Checklist

1. **Update version** in relevant files
2. **Update CHANGELOG.md** with release notes
3. **Ensure tests pass** on all supported platforms
4. **Create release tag**
   ```bash
   git tag -a v1.2.0 -m "Release v1.2.0"
   git push upstream v1.2.0
   ```
5. **GitHub Actions** will automatically build and publish release

### Changelog Format

```markdown
## [1.2.0] - 2024-01-15

### Added
- New export format: SQL
- Support for custom LLM providers
- Interactive setup wizard

### Changed
- Improved error messages for configuration validation
- Updated dependency detection logic

### Deprecated
- Legacy configuration file format (will be removed in v2.0.0)

### Removed
- Deprecated `--legacy-mode` flag

### Fixed
- Race condition in database operations
- Memory leak in file processing

### Security
- Updated dependencies to address CVE-2024-12345
```

## Community Guidelines

### Code of Conduct

- Be welcoming and inclusive
- Be respectful of different viewpoints and experiences
- Give and gracefully accept constructive feedback
- Focus on what is best for the community
- Show empathy towards other community members

### Getting Help

- **GitHub Discussions** for questions and general discussion
- **GitHub Issues** for bug reports and feature requests
- **Code reviews** for technical guidance
- **Documentation** for implementation details

### Recognition

Contributors are recognized through:
- GitHub contributor statistics
- Mention in release notes for significant contributions
- Invitation to maintainer team for sustained contributions

Thank you for contributing to Waffles! 🧇