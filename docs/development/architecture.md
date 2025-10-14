# Architecture Overview

This document describes the overall architecture of Waffles, including system design, component relationships, and key design decisions.

## Table of Contents

- [High-Level Architecture](#high-level-architecture)
- [Component Overview](#component-overview)
- [Data Flow](#data-flow)
- [Key Design Decisions](#key-design-decisions)
- [Dependency Management](#dependency-management)
- [Extension Points](#extension-points)

## High-Level Architecture

Waffles follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                     CLI Layer                               │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐ │
│  │   query     │ │    setup    │ │    deps    export       │ │
│  └─────────────┘ └─────────────┘ └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  Internal Services                          │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────┐ │
│  │    setup     │ │    query     │ │       export         │ │
│  │   wizard     │ │   engine     │ │      formats         │ │
│  └──────────────┘ └──────────────┘ └──────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 Core Libraries                              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌──────────┐│
│  │   config    │ │  pipeline   │ │    repo     │ │ logging  ││
│  └─────────────┘ └─────────────┘ └─────────────┘ └──────────┘│
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌──────────┐│
│  │    deps     │ │  version    │ │    man      │ │   ...    ││
│  └─────────────┘ └─────────────┘ └─────────────┘ └──────────┘│
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              External Dependencies                          │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────┐ │
│  │wheresmyprompt│ │files2prompt  │ │      llm CLI         │ │
│  └──────────────┘ └──────────────┘ └──────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Component Overview

### CLI Layer (`cmd/waffles/`)

The command-line interface layer handles user interaction and delegates to internal services.

**Key Files:**
- [`cmd/waffles/main.go`](../../cmd/waffles/main.go) - Application entry point
- [`cmd/waffles/root.go`](../../cmd/waffles/root.go) - Root command and global flags
- [`cmd/waffles/query.go`](../../cmd/waffles/query.go) - Query command implementation
- [`cmd/waffles/setup.go`](../../cmd/waffles/setup.go) - Setup wizard command
- [`cmd/waffles/deps.go`](../../cmd/waffles/deps.go) - Dependency management commands
- [`cmd/waffles/export.go`](../../cmd/waffles/export.go) - Export functionality

**Responsibilities:**
- Command-line argument parsing (using Cobra)
- User input validation
- Output formatting and display
- Error handling and user feedback
- Delegation to internal services

### Internal Services (`internal/`)

Business logic and orchestration services that implement core functionality.

#### Setup Service (`internal/setup/`)
- **Purpose**: Interactive setup wizard and configuration management
- **Key Components**:
  - `wizard.go` - Interactive setup flow
  - `prompts.go` - User interaction prompts
  - `validation.go` - Configuration validation

#### Query Engine (`internal/query/`)
- **Purpose**: Query processing and pipeline orchestration
- **Key Components**:
  - `engine.go` - Main query processing logic
  - `context.go` - Execution context management
  - `processors.go` - Query processing pipeline

#### Export System (`internal/export/`)
- **Purpose**: Data export and formatting
- **Key Components**:
  - `exporters.go` - Export orchestration
  - `formats.go` - Output format implementations
  - `filters.go` - Data filtering logic

### Core Libraries (`pkg/`)

Reusable components that implement core functionality.

#### Configuration Management (`pkg/config/`)
- **Purpose**: Environment-based configuration management
- **Architecture**: Hierarchical configuration loading
- **Key Features**:
  - Environment variable support
  - `.env` file loading
  - Default value management
  - Validation and type conversion

#### Pipeline Orchestration (`pkg/pipeline/`)
- **Purpose**: Execution pipeline management
- **Architecture**: Pipeline pattern with configurable steps
- **Components**:
  - Pipeline definition and validation
  - Execution context management
  - Error handling and recovery
  - Performance monitoring

#### Repository Analysis (`pkg/repo/`)
- **Purpose**: Project analysis and file detection
- **Architecture**: Strategy pattern for language detection
- **Components**:
  - Language detection algorithms
  - File pattern matching
  - Gitignore integration
  - Metadata extraction

#### Dependency Management (`pkg/deps/`)
- **Purpose**: External dependency management
- **Architecture**: Plugin-like architecture for different tools
- **Components**:
  - Dependency detection and validation
  - Installation automation
  - Version management
  - Platform-specific handling

#### Logging and Database (`pkg/logging/`)
- **Purpose**: Persistent storage and analytics
- **Architecture**: Repository pattern with SQLite backend
- **Components**:
  - Execution history tracking
  - Performance analytics
  - Query interface
  - Schema management

## Data Flow

### Query Execution Flow

```
User Input
    │
    ▼
┌───────────────────┐
│   CLI Command     │ ─── Parse arguments and flags
└───────────────────┘
    │
    ▼
┌───────────────────┐
│  Configuration    │ ─── Load config from env/files
│     Loading       │
└───────────────────┘
    │
    ▼
┌───────────────────┐
│   Dependency      │ ─── Verify required tools
│     Check         │
└───────────────────┘
    │
    ▼
┌───────────────────┐
│   Repository      │ ─── Analyze project structure
│    Analysis       │
└───────────────────┘
    │
    ▼
┌───────────────────┐
│   Pipeline        │ ─── wheresmyprompt → files2prompt → llm
│   Execution       │
└───────────────────┘
    │
    ▼
┌───────────────────┐
│    Database       │ ─── Log execution results
│    Logging        │
└───────────────────┘
    │
    ▼
┌───────────────────┐
│  Output           │ ─── Format and display results
│ Formatting        │
└───────────────────┘
```

### Configuration Flow

```
Default Values
    │
    ▼
Environment Variables ──────┐
    │                       │
    ▼                       │
Global .env File           │
    │                       │
    ▼                       │
Project .env File          │
    │                       │
    ▼                       │
Command Line Flags ────────┤
    │                       │
    ▼                       ▼
Final Configuration ───► Application
```

## Key Design Decisions

### 1. Environment-Based Configuration

**Decision**: Use environment variables and `.env` files instead of traditional config files.

**Rationale**:
- Follows 12-factor app principles
- Easy integration with CI/CD systems
- Supports per-project configuration
- Simple hierarchy and override mechanism

**Implementation**: [`pkg/config/config.go`](../../pkg/config/config.go)

### 2. Pipeline Architecture

**Decision**: Implement pipeline as configurable, orchestrated steps.

**Rationale**:
- Clear separation of concerns
- Easy to test individual components
- Configurable timeout and error handling
- Extensible for future tools

**Implementation**: [`pkg/pipeline/executor.go`](../../pkg/pipeline/executor.go)

### 3. SQLite for Storage

**Decision**: Use SQLite for persistent storage instead of flat files.

**Rationale**:
- Rich querying capabilities for analytics
- Atomic transactions for data integrity
- No external database dependency
- Excellent Go library support

**Implementation**: [`pkg/logging/database.go`](../../pkg/logging/database.go)

### 4. Cobra for CLI

**Decision**: Use Cobra framework for command-line interface.

**Rationale**:
- Industry standard for Go CLI applications
- Rich flag and subcommand support
- Built-in help generation
- Consistent with other tools (kubectl, helm, etc.)

**Implementation**: [`cmd/waffles/root.go`](../../cmd/waffles/root.go)

### 5. Clean Architecture Principles

**Decision**: Organize code using clean architecture patterns.

**Rationale**:
- Clear dependency direction (inward)
- Testable business logic
- Framework independence
- Maintainable codebase

**Structure**:
- `cmd/` - Interface adapters (CLI)
- `internal/` - Application business rules
- `pkg/` - Enterprise business rules (reusable)

### 6. Dependency Injection

**Decision**: Use dependency injection for major components.

**Rationale**:
- Testability with mock objects
- Flexible component configuration
- Clear component dependencies
- Runtime behavior modification

**Example**: Pipeline accepts configurable Logger interface

## Dependency Management

### External Dependencies

Waffles orchestrates three external tools:

1. **wheresmyprompt** - System prompt retrieval
2. **files2prompt** - File context extraction  
3. **llm** - LLM interaction

**Integration Strategy**:
- Command-line execution via `os/exec`
- Timeout management
- Error handling and recovery
- Version compatibility checking

### Go Dependencies

**Core Dependencies**:
- `github.com/spf13/cobra` - CLI framework
- `github.com/mattn/go-sqlite3` - SQLite driver
- `github.com/google/uuid` - UUID generation
- `github.com/Masterminds/semver/v3` - Semantic versioning

**Testing Dependencies**:
- Standard library `testing` package
- No external testing frameworks for simplicity

## Extension Points

### 1. New Output Formats

**Interface**: `internal/export/formats.go`

```go
type Formatter interface {
    Format(data interface{}) ([]byte, error)
    ContentType() string
}
```

**Extension Process**:
1. Implement the `Formatter` interface
2. Register in format registry
3. Add command-line flag support

### 2. New Pipeline Steps

**Interface**: `pkg/pipeline/types.go`

```go
type Step interface {
    Execute(ctx ExecutionContext) (*StepResult, error)
    Validate() error
}
```

**Extension Process**:
1. Implement the `Step` interface
2. Add to pipeline configuration
3. Handle errors and timeouts

### 3. New Repository Languages

**Interface**: `pkg/repo/detector.go`

Language detection is pattern-based and configurable:

```go
type LanguageDetector interface {
    DetectLanguage(path string) (Language, error)
    GetPatterns(lang Language) (include, exclude []string)
}
```

### 4. New Export Targets

**Interface**: `internal/export/exporters.go`

```go
type Exporter interface {
    Export(data []WafflesExecution, opts ExportOptions) error
    ValidateOptions(opts ExportOptions) error
}
```

## Performance Considerations

### 1. File System Operations
- Concurrent file scanning where safe
- Configurable file size and count limits
- Efficient pattern matching

### 2. Database Operations
- Connection pooling for concurrent access
- Indexed queries for common operations
- Batch operations for bulk data

### 3. External Process Execution
- Configurable timeouts per tool
- Process cleanup on cancellation
- Resource usage monitoring

### 4. Memory Management
- Streaming for large file processing
- Configurable memory limits
- Garbage collection optimization

## Security Considerations

### 1. Input Validation
- Command injection prevention
- File path traversal protection
- Configuration value validation

### 2. Process Execution
- Controlled environment variables
- Limited process privileges
- Timeout enforcement

### 3. Database Security
- Parameterized queries only
- File permission management
- No sensitive data logging

### 4. External Tool Integration
- Version verification
- Checksum validation (future)
- Sandbox execution (consideration)

## Future Architecture Considerations

### 1. Plugin System
- Dynamic loading of extensions
- Plugin versioning and compatibility
- Security model for plugins

### 2. Distributed Execution
- Remote LLM service support
- Caching and result sharing
- Load balancing considerations

### 3. Advanced Analytics
- Metrics collection
- Performance profiling
- Usage pattern analysis

### 4. Configuration Management
- Schema validation
- Migration support
- Environment-specific overrides

This architecture provides a solid foundation for current functionality while remaining extensible for future enhancements.