# Waffles Application Requirements Document

## Overview

This document outlines the specific requirements for implementing the Waffles application, a command-line tool that orchestrates an LLM toolchain for local development workflows. The application automates the process of gathering project context, retrieving prompts, and executing LLM queries while maintaining comprehensive logging for transparency.

## 1. Code Base Cleanup Requirements

### 1.1 Remove Golang-Starter Artifacts
- **Requirement**: Remove all references and artifacts from the golang-starter template
- **Specific Actions**:
  - Delete the [`internal/starter/`](../internal/starter/) directory entirely
  - Remove any references to "starter" in code comments, documentation, or variable names
  - Update import paths that reference starter components
  - Remove any starter-specific configuration or setup code

### 1.2 Project Structure Reorganization
- **Requirement**: Maintain clean Go project structure following standard conventions
- **Specific Actions**:
  - Keep [`cmd/waffles/`](../cmd/waffles/) for CLI entry point
  - Utilize [`pkg/`](../pkg/) for public library code
  - Utilize [`internal/`](../internal/) for private application code
  - Remove starter-specific packages, keep only those relevant to Waffles

## 2. Configuration Management Requirements

### 2.1 Replace Viper with Environment-based Configuration
- **Requirement**: Replace Viper library with [`godotenv`](https://github.com/joho/godotenv) and [`env`](https://github.com/caarlos0/env)
- **Rationale**: Simplify configuration management using environment variables and `.env` files
- **Implementation Requirements**:
  - Use `godotenv` to load environment variables from `.env` file
  - Use `env` to parse environment variables into Go structs
  - Support for both `.env` file and system environment variables
  - Environment variables take precedence over `.env` file values

### 2.2 Configuration Structure
```go
type Config struct {
    // Core Application Settings
    DefaultModel    string `env:"WAFFLES_DEFAULT_MODEL" envDefault:"claude-3-sonnet"`
    DefaultProvider string `env:"WAFFLES_DEFAULT_PROVIDER" envDefault:"anthropic"`
    LogDBPath      string `env:"WAFFLES_LOG_DB_PATH" envDefault:"./llm-logs.sqlite"`
    ConfigPath     string `env:"WAFFLES_CONFIG_PATH" envDefault:".waffles.env"`
    
    // Tool Configurations
    WheresmypromptArgs string `env:"WHERESMYPROMPT_ARGS" envDefault:""`
    Files2promptArgs   string `env:"FILES2PROMPT_ARGS" envDefault:""`
    LLMArgs           string `env:"LLM_ARGS" envDefault:""`
    
    // Behavior Settings
    Verbose         bool   `env:"WAFFLES_VERBOSE" envDefault:"false"`
    Quiet          bool   `env:"WAFFLES_QUIET" envDefault:"false"`
    AutoInstall    bool   `env:"WAFFLES_AUTO_INSTALL" envDefault:"false"`
    IgnoreGitignore bool  `env:"WAFFLES_IGNORE_GITIGNORE" envDefault:"false"`
    
    // Language-specific Settings
    LanguageOverride string `env:"WAFFLES_LANGUAGE" envDefault:""`
    IncludePatterns  string `env:"WAFFLES_INCLUDE_PATTERNS" envDefault:""`
    ExcludePatterns  string `env:"WAFFLES_EXCLUDE_PATTERNS" envDefault:""`
}
```

### 2.3 Configuration Loading Requirements
- Load configuration in the following priority order:
  1. Command-line flags (highest priority)
  2. System environment variables
  3. `.env` file in current directory
  4. Global `.env` file in `~/.config/waffles/.env`
  5. Built-in defaults (lowest priority)

## 3. Core Functional Requirements

### 3.1 External Tool Integration
- **Requirement**: Integrate with three external tools
- **Tools Required**:
  - [`wheresmyprompt`](https://github.com/toozej/wheresmyprompt): Go-based prompt retrieval
  - [`files2prompt`](https://github.com/toozej/files2prompt): Go-based context extraction  
  - [`llm`](https://github.com/simonw/llm): Python-based LLM CLI with SQLite logging

### 3.2 Dependency Management System
- **Component**: [`pkg/deps/`](../pkg/deps/)
- **Requirements**:
  - Detect presence of required tools in system PATH
  - Verify minimum compatible versions
  - Validate LLM plugins are installed (`llm-anthropic`, `llm-ollama`, `llm-gemini`, etc.)
  - Provide auto-installation capabilities with `--install` flag
  - Generate clear installation instructions when auto-install unavailable
  - Support for both local and system-wide tool installations

### 3.3 Repository Analysis System  
- **Component**: [`pkg/repo/`](../pkg/repo/)
- **Requirements**:
  - Automatically detect project programming language
  - Generate appropriate file inclusion/exclusion patterns
  - Support for multiple languages:
    - **Go**: Include `*.go`, exclude `*_test.go`, `pkg/version/*`, `pkg/man/*`
    - **Python**: Include `*.py`, exclude `*test*.py`, `__init__.py`, `__pycache__/`
    - **Extensible**: Plugin architecture for additional languages
  - Respect `.gitignore` rules (with override option)
  - Manual override capabilities via CLI flags

### 3.4 Pipeline Orchestration System
- **Component**: [`pkg/pipeline/`](../pkg/pipeline/)
- **Requirements**:
  - Sequential execution: `wheresmyprompt` → `files2prompt` → `llm`
  - Robust error handling with graceful degradation
  - Output coordination between tools (pipe outputs as inputs)
  - Support for concurrent operations where safe
  - Execution timing and performance metrics

### 3.5 Logging and Database System
- **Component**: [`pkg/logging/`](../pkg/logging/)
- **Requirements**:
  - Per-repository SQLite databases (`../llm-logs.sqlite`)
  - Extend llm CLI's native logging with Waffles-specific data
  - Track complete execution context (files processed, arguments used, etc.)
  - Support for querying and exporting logged data
  - Configurable log retention policies

## 4. CLI Interface Requirements

### 4.1 Primary Command Structure
```bash
waffles [options] [prompt-search-terms...]
```

### 4.2 Required CLI Flags
```bash
# Model and Provider Options
--model, -m          LLM model to use
--provider          LLM provider override

# Tool Configuration  
--wheresmyprompt-args    Pass arguments to wheresmyprompt
--files2prompt-args      Pass arguments to files2prompt
--llm-args              Pass arguments to llm CLI

# Repository Options
--language          Override auto-detected language
--include           Additional file patterns to include
--exclude           File patterns to exclude
--ignore-gitignore  Don't respect .gitignore rules

# Installation and Setup
--install           Auto-install missing dependencies
--check-deps        Only check dependencies
--setup             Interactive setup wizard

# Output and Logging
--output, -o        Output file (default: stdout)
--log-db           SQLite database path
--quiet, -q         Suppress progress output
--verbose, -v       Detailed execution logging

# Configuration
--config            Configuration file path (.env)
--env-file          Alternative .env file path
```

### 4.3 Required Subcommands
```bash
waffles deps        # Dependency management
waffles setup       # Interactive setup
waffles query       # Query logged conversations  
waffles export      # Export logged data
waffles version     # Version information
waffles help        # Help and usage information
```

## 5. Database Schema Requirements

### 5.1 SQLite Database Structure
- **Location**: `../llm-logs.sqlite` in current working directory
- **Integration**: Leverage existing llm CLI logging tables
- **Extensions**: Add Waffles-specific tracking tables

### 5.2 Required Schema Extensions
```sql
-- Waffles execution tracking
CREATE TABLE waffles_executions (
    id TEXT PRIMARY KEY,
    conversation_id TEXT,
    command_args TEXT,
    wheresmyprompt_query TEXT,
    files2prompt_args TEXT,
    detected_language TEXT,
    file_count INTEGER,
    execution_time_ms INTEGER,
    created INTEGER,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);

-- File processing tracking
CREATE TABLE waffles_files (
    id TEXT PRIMARY KEY,
    execution_id TEXT,
    file_path TEXT,
    file_size INTEGER,
    included BOOLEAN,
    exclusion_reason TEXT,
    FOREIGN KEY (execution_id) REFERENCES waffles_executions(id)
);
```

## 6. Error Handling Requirements

### 6.1 Dependency Error Handling
- Provide clear error messages for missing tools
- Offer specific installation commands for each missing dependency
- Support for partial functionality when some tools are unavailable
- Graceful fallback options where possible

### 6.2 Repository Analysis Error Handling
- Handle unrecognized project types with manual configuration options
- Manage mixed-language projects with prioritization algorithms
- Provide informative messages for empty or unusual repositories
- Support manual file specification when auto-detection fails

### 6.3 Tool Execution Error Handling
- Propagate errors from external tools with context
- Provide suggested fixes for common tool failures
- Support for manual input when tool chains break
- Comprehensive logging of all error conditions

## 7. Security Requirements

### 7.1 API Key Management
- Secure handling of LLM provider API keys via environment variables
- No API keys stored in configuration files or logs
- Support for multiple provider configurations

### 7.2 Input Validation
- Sanitize all user inputs passed to external tools
- Prevent command injection in tool execution
- Validate file paths to prevent directory traversal
- Set reasonable resource limits for file processing

### 7.3 Data Privacy
- All data processing occurs locally
- No data transmitted to external services except chosen LLM providers
- Configurable filtering of sensitive information
- Optional SQLite database encryption

## 8. Performance Requirements

### 8.1 Scalability
- Handle repositories with thousands of files efficiently
- Support for incremental processing of changed files
- Configurable limits on memory and disk usage
- Progress indication for long-running operations

### 8.2 Optimization
- Concurrent processing where thread-safe
- Intelligent caching of repository analysis results
- Minimal overhead for small repositories
- Configurable timeout values for external tool execution

## 9. Testing Requirements

### 9.1 Unit Testing
- Comprehensive test coverage for all packages
- Mock external tool dependencies for testing
- Test configuration loading and validation
- Test error handling paths

### 9.2 Integration Testing
- End-to-end testing with real external tools
- Database integration testing
- CLI interface testing
- Cross-platform compatibility testing

## 10. Documentation Requirements

### 10.1 User Documentation
- Complete CLI usage documentation
- Installation and setup guides
- Configuration examples and best practices
- Troubleshooting guides

### 10.2 Developer Documentation
- Code documentation with godoc
- Architecture decision records
- Contributing guidelines
- API documentation for public packages
- GitHub releases with binary attachments
- Package manager support (Homebrew, APT, etc.)
- Docker container support
- Installation scripts for automated setup

## 11. Migration Requirements

### 11.1 From Golang-Starter Template
- Remove all starter-specific code and references
- Migrate existing configuration system to environment-based approach
- Update documentation to remove starter references
- Preserve useful components (CLI structure, build system)

### 11.2 Configuration Migration
- Provide migration tools for existing Viper-based configurations
- Clear migration documentation
- Backwards compatibility considerations
- Validation of migrated configurations

## 12. Implementation Priority

### 12.1 Phase 1: Foundation (High Priority)
- Code cleanup and starter removal
- Environment-based configuration system
- Basic CLI structure with cobra
- Dependency detection system

### 12.2 Phase 2: Core Features (High Priority)  
- Repository analysis and language detection
- External tool integration
- Pipeline orchestration
- SQLite logging system

### 12.3 Phase 3: Advanced Features (Medium Priority)
- Auto-installation system
- Query and export capabilities
- Enhanced error handling
- Performance optimization

### 12.4 Phase 4: Polish (Low Priority)
- Comprehensive testing
- Documentation completion
- Distribution packages
- Optional features and extensions

---

This requirements document serves as the comprehensive specification for implementing the Waffles application. Each requirement should be verifiable and testable upon completion.