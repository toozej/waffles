# Code Organization

This document explains the organization and structure of the Waffles codebase, including directory layout, naming conventions, and architectural patterns.

## Table of Contents

- [Directory Structure](#directory-structure)
- [Package Organization](#package-organization)
- [Naming Conventions](#naming-conventions)
- [File Organization](#file-organization)
- [Import Structure](#import-structure)
- [Interface Design](#interface-design)

## Directory Structure

```
waffles/
├── cmd/                    # Command-line applications
│   └── waffles/           # Main CLI application
├── internal/              # Private application code
│   ├── export/            # Export functionality
│   ├── query/             # Query processing
│   └── setup/             # Setup wizard
├── pkg/                   # Public library code
│   ├── config/            # Configuration management
│   ├── deps/              # Dependency management
│   ├── logging/           # Database and logging
│   ├── man/               # Manual page generation
│   ├── pipeline/          # Pipeline orchestration
│   ├── repo/              # Repository analysis
│   └── version/           # Version information
├── docs/                  # Documentation
│   ├── development/       # Developer documentation
│   └── *.md              # User documentation
├── .github/               # GitHub specific files
│   └── workflows/         # CI/CD workflows
├── testdata/              # Test data and fixtures
├── scripts/               # Build and utility scripts
└── ...                    # Config files (go.mod, Makefile, etc.)
```

## Package Organization

### `cmd/waffles/` - CLI Application

Entry point and command-line interface for the application.

**Purpose**: 
- Command-line argument parsing
- User interface and interaction
- Delegation to internal services

**Key Files**:
- [`main.go`](../../cmd/waffles/main.go) - Application entry point
- [`root.go`](../../cmd/waffles/root.go) - Root command and global configuration
- [`query.go`](../../cmd/waffles/query.go) - Query command implementation
- [`setup.go`](../../cmd/waffles/setup.go) - Setup wizard command
- [`deps.go`](../../cmd/waffles/deps.go) - Dependency management commands
- [`export.go`](../../cmd/waffles/export.go) - Export command implementation

**Dependencies**: 
- `internal/` packages for business logic
- `pkg/` packages for utilities
- External: `github.com/spf13/cobra`

### `internal/` - Private Application Logic

Business logic that is specific to Waffles and not intended for reuse.

#### `internal/setup/` - Setup Wizard

Interactive configuration and first-time setup.

**Files**:
- [`wizard.go`](../../internal/setup/wizard.go) - Main setup wizard logic
- [`prompts.go`](../../internal/setup/prompts.go) - User interaction prompts
- [`validation.go`](../../internal/setup/validation.go) - Input validation

**Responsibilities**:
- Interactive user prompts
- Configuration validation
- Dependency installation coordination
- First-time setup workflow

#### `internal/query/` - Query Processing

Core query processing and orchestration logic.

**Files**:
- [`engine.go`](../../internal/query/engine.go) - Main query processing engine
- [`context.go`](../../internal/query/context.go) - Execution context management
- [`processors.go`](../../internal/query/processors.go) - Query processing pipeline

**Responsibilities**:
- Query validation and preprocessing
- Pipeline orchestration
- Result processing and formatting
- Error handling and recovery

#### `internal/export/` - Export System

Data export functionality and format handling.

**Files**:
- [`exporters.go`](../../internal/export/exporters.go) - Export coordination
- [`formats.go`](../../internal/export/formats.go) - Output format implementations
- [`filters.go`](../../internal/export/filters.go) - Data filtering logic

**Responsibilities**:
- Data query and retrieval
- Format conversion
- Output generation
- Filter application

### `pkg/` - Reusable Library Code

Public packages that could potentially be used by other applications.

#### `pkg/config/` - Configuration Management

Environment-based configuration loading and management.

**Files**:
- [`config.go`](../../pkg/config/config.go) - Main configuration struct and loading
- [`env.go`](../../pkg/config/env.go) - Environment variable handling
- [`validation.go`](../../pkg/config/validation.go) - Configuration validation

**Key Types**:
```go
type Config struct {
    DefaultModel    string
    DefaultProvider string
    LogDBPath       string
    // ... other fields
}
```

#### `pkg/deps/` - Dependency Management

External tool detection, validation, and installation.

**Files**:
- [`types.go`](../../pkg/deps/types.go) - Core types and interfaces
- [`detector.go`](../../pkg/deps/detector.go) - Dependency detection logic
- [`installer.go`](../../pkg/deps/installer.go) - Installation automation
- [`dependencies.go`](../../pkg/deps/dependencies.go) - Dependency definitions

**Key Types**:
```go
type Dependency struct {
    Name        string
    Command     string
    CheckCommand string
    MinVersion  string
    Plugins     []string
}

type DependencyStatus struct {
    Name      string
    Installed bool
    Valid     bool
    Version   string
    Error     string
    Plugins   []PluginStatus
}
```

#### `pkg/pipeline/` - Pipeline Orchestration

Execution pipeline for coordinating external tools.

**Files**:
- [`types.go`](../../pkg/pipeline/types.go) - Pipeline types and interfaces
- [`executor.go`](../../pkg/pipeline/executor.go) - Pipeline execution logic
- [`commands.go`](../../pkg/pipeline/commands.go) - Command building utilities
- [`utils.go`](../../pkg/pipeline/utils.go) - Pipeline utilities

**Key Types**:
```go
type Pipeline struct {
    Config   *config.Config
    RepoInfo *repo.RepositoryInfo
    Logger   Logger
}

type ExecutionContext struct {
    ID             string
    StartTime      time.Time
    PromptQuery    string
    ExecutionSteps []StepResult
    Success        bool
    FinalOutput    string
}
```

#### `pkg/repo/` - Repository Analysis

Project analysis, language detection, and file pattern matching.

**Files**:
- [`types.go`](../../pkg/repo/types.go) - Repository types and enums
- [`detector.go`](../../pkg/repo/detector.go) - Language detection logic
- [`patterns.go`](../../pkg/repo/patterns.go) - File pattern matching
- [`gitignore.go`](../../pkg/repo/gitignore.go) - Gitignore integration

**Key Types**:
```go
type Language string

type RepositoryInfo struct {
    Language        Language
    RootPath        string
    IncludePatterns []string
    ExcludePatterns []string
    DetectedFiles   []FileInfo
    GitIgnoreRules  []string
}
```

#### `pkg/logging/` - Database and Logging

Persistent storage for execution history and analytics.

**Files**:
- [`types.go`](../../pkg/logging/types.go) - Database types and schemas
- [`database.go`](../../pkg/logging/database.go) - Database operations
- [`query.go`](../../pkg/logging/query.go) - Query interface
- [`schema.go`](../../pkg/logging/schema.go) - Database schema and migrations

**Key Types**:
```go
type Database struct {
    path string
    db   *sql.DB
}

type WafflesExecution struct {
    ID                  string
    CommandArgs         string
    WheresmypromptQuery string
    DetectedLanguage    string
    FileCount           int
    Success             bool
    Created             time.Time
}
```

## Naming Conventions

### Package Names
- **Short and descriptive**: `config`, `deps`, `repo`
- **Lowercase**: No camelCase or underscores
- **Singular**: `config` not `configs`
- **Avoid abbreviations**: `repository` shortened to `repo` is acceptable

### File Names
- **Lowercase with underscores**: `detector_test.go`
- **Descriptive**: `config.go`, `installer.go`
- **Test files**: `*_test.go` suffix
- **Main files**: Name after primary type/function

### Type Names
- **Exported types**: PascalCase (`Config`, `Pipeline`)
- **Unexported types**: camelCase (`executionContext`)
- **Interfaces**: Often end with `-er` (`Logger`, `Exporter`)
- **Descriptive**: `DependencyStatus` not `DepStat`

### Function Names
- **Exported functions**: PascalCase (`LoadConfig`)
- **Unexported functions**: camelCase (`parseArgs`)
- **Verbs first**: `GetConfig()`, `ValidateInput()`
- **Boolean returns**: `IsValid()`, `HasFeature()`

### Variable Names
- **Short in small scopes**: `cfg` for config in function
- **Descriptive in larger scopes**: `configuration` for struct field
- **No Hungarian notation**: `userID` not `strUserID`

### Constants
- **All caps with underscores**: `DEFAULT_TIMEOUT`
- **Grouped in blocks**:
  ```go
  const (
      DefaultModel = "claude-3-sonnet"
      DefaultProvider = "anthropic"
  )
  ```

## File Organization

### Within Packages

Each package follows a consistent file organization pattern:

1. **`types.go`** - Core types, structs, interfaces, and constants
2. **Main implementation files** - Named after primary functionality
3. **`*_test.go`** - Test files alongside the code they test
4. **`example_test.go`** - Example usage (where applicable)

### File Structure Template

```go
// Package comment explaining purpose
package packagename

// Imports organized in groups:
// 1. Standard library
// 2. External dependencies  
// 3. Internal project imports

import (
    // Standard library
    "context"
    "fmt"
    "os"
    
    // External dependencies
    "github.com/spf13/cobra"
    
    // Internal imports
    "github.com/toozej/waffles/pkg/config"
)

// Constants and variables
const (
    DefaultValue = "default"
)

// Types (interfaces first, then structs)
type SomeInterface interface {
    Method() error
}

type SomeStruct struct {
    field string
}

// Functions (exported first, then unexported)
func PublicFunction() error {
    return nil
}

func privateFunction() {
    // implementation
}
```

## Import Structure

### Import Groups

Imports are organized into three groups, separated by blank lines:

1. **Standard library packages**
2. **External dependencies**
3. **Internal project packages**

### Import Paths

```go
import (
    // Standard library
    "context"
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"
    
    // External dependencies
    "github.com/google/uuid"
    "github.com/mattn/go-sqlite3"
    "github.com/spf13/cobra"
    
    // Internal packages
    "github.com/toozej/waffles/internal/query"
    "github.com/toozej/waffles/pkg/config"
    "github.com/toozej/waffles/pkg/pipeline"
)
```

### Import Aliases

- Avoid aliases unless necessary for disambiguation
- Use descriptive aliases when needed:
  ```go
  import (
      "database/sql"
      sqlite "github.com/mattn/go-sqlite3"
  )
  ```

## Interface Design

### Interface Principles

1. **Small interfaces**: Prefer single-method interfaces
2. **Focused responsibility**: Each interface has one clear purpose  
3. **Accept interfaces, return structs**: Function parameters use interfaces
4. **Define at usage point**: Interfaces defined where they're used

### Common Interface Patterns

#### Logger Interface
```go
type Logger interface {
    LogExecution(exec *WafflesExecution) error
    LogFiles(executionID string, files []WafflesFile) error
    LogSteps(executionID string, steps []WafflesStep) error
}
```

#### Exporter Interface
```go
type Exporter interface {
    Export(data []WafflesExecution, opts ExportOptions) error
    ValidateOptions(opts ExportOptions) error
}
```

#### Formatter Interface
```go
type Formatter interface {
    Format(data interface{}) ([]byte, error)
    ContentType() string
}
```

## Error Handling

### Error Types

Custom error types for different categories:

```go
// Pipeline errors
type PipelineError struct {
    Phase   ExecutionPhase
    Tool    string
    Message string
    Err     error
}

func (pe *PipelineError) Error() string {
    return fmt.Sprintf("pipeline error in %s phase: %s", pe.Phase, pe.Message)
}

func (pe *PipelineError) Unwrap() error {
    return pe.Err
}
```

### Error Wrapping

Use `fmt.Errorf` with `%w` verb for error wrapping:

```go
if err := someOperation(); err != nil {
    return fmt.Errorf("failed to perform operation: %w", err)
}
```

## Testing Organization

### Test File Structure

- Test files named `*_test.go`
- Tests in same package as code under test
- Use table-driven tests where appropriate
- Separate unit tests from integration tests

### Test Helper Functions

Common test helpers in package-level files:

```go
// testhelpers.go
func setupTestDB(t *testing.T) *Database {
    db, err := NewDatabase(":memory:")
    if err != nil {
        t.Fatalf("Failed to create test database: %v", err)
    }
    return db
}
```

### Test Data

- Test data in `testdata/` directories
- JSON fixtures for complex test data
- Generated test data for property-based testing

## Documentation

### Code Documentation

#### Package Documentation
```go
// Package config provides environment-based configuration management
// for the Waffles application. It supports hierarchical configuration
// loading from environment variables, .env files, and default values.
package config
```

#### Function Documentation
```go
// LoadConfig loads configuration from environment variables and .env files.
// It follows a hierarchical loading strategy where command-line flags take
// precedence over environment variables, which take precedence over .env
// files, which take precedence over default values.
//
// The function returns an error if required configuration is missing or
// if there are validation errors.
func LoadConfig() (*Config, error) {
    // implementation
}
```

### Inline Comments

- Explain **why**, not **what**
- Focus on business logic and complex algorithms
- Avoid obvious comments

```go
// Good: Explains reasoning
// Use exponential backoff to handle rate limiting from LLM providers
time.Sleep(time.Duration(attempt*attempt) * time.Second)

// Bad: States the obvious  
// Sleep for some time
time.Sleep(duration)
```

## Code Quality Standards

### Formatting
- Use `gofmt` or `goimports` for consistent formatting
- Line length soft limit of 100 characters
- Group related functionality together

### Linting
- Use `golangci-lint` for comprehensive linting
- Address all linter warnings
- Configure project-specific linting rules in `.golangci.yml`

### Code Review Guidelines
- Functions should do one thing well
- Minimize cyclomatic complexity
- Handle all error cases appropriately
- Include tests for new functionality
- Update documentation for user-facing changes

This organization provides a solid foundation for maintainable, scalable Go code while following established conventions and best practices.