# Waffles Application Implementation Tasks

This document provides a comprehensive list of implementation tasks for building the Waffles application. Each task is designed to be executed by an AI assistant and includes specific deliverables, acceptance criteria, and implementation details derived from [`DESIGN.md`](DESIGN.md) and [`REQUIREMENTS.md`](REQUIREMENTS.md).

## Task Progress Tracking

Use this checklist to track progress and resume work from specific points:

### Phase 1: Foundation and Cleanup (High Priority)
- [ ] **Task 1.1**: Remove Golang-Starter Artifacts
- [ ] **Task 1.2**: Replace Viper with Environment-based Configuration
- [ ] **Task 1.3**: Update CLI Structure with Cobra

### Phase 2: Core System Components (High Priority)
- [ ] **Task 2.1**: Implement Dependency Management System
- [ ] **Task 2.2**: Implement Repository Analysis System
- [ ] **Task 2.3**: Implement Pipeline Orchestration System
- [ ] **Task 2.4**: Implement Logging and Database System

### Phase 3: Advanced Features (Medium Priority)
- [ ] **Task 3.1**: Implement Auto-Installation System
- [ ] **Task 3.2**: Implement Interactive Setup Wizard
- [ ] **Task 3.3**: Implement Query and Export System

### Phase 4: Testing, Documentation, and Polish (Medium-Low Priority)
- [ ] **Task 4.1**: Implement Comprehensive Testing Suite
- [ ] **Task 4.2**: Create User Documentation
- [ ] **Task 4.3**: Create Developer Documentation

## How to Use This Document

### For AI Implementation:
1. **Starting Fresh**: Begin with Task 1.1 and work sequentially through Phase 1
2. **Resuming Work**: Find the last completed task (marked with [x]) and continue with the next [ ] task
3. **After Completion**: Mark tasks as complete [x] and update progress
4. **On Failure**: Review the task's acceptance criteria and deliverables to understand what needs to be fixed

### Task Status Legend:
- [ ] **Not Started**: Task has not been attempted
- [x] **Completed**: Task finished and acceptance criteria met
- [-] **In Progress**: Task currently being worked on
- [!] **Blocked**: Task cannot proceed due to dependencies or issues

### Phase Completion Validation:
- **Phase 1 Complete**: Basic CLI works, configuration loads, no starter artifacts remain
- **Phase 2 Complete**: End-to-end pipeline executes successfully with all tools
- **Phase 3 Complete**: All advanced features work and are properly tested
- **Phase 4 Complete**: Comprehensive testing and documentation complete

## Quick Start Guide for AI Implementation

### Initial Setup Commands:
```bash
# Verify current project state
ls -la internal/
grep -r "starter" . || echo "No starter references found"
go mod tidy
make local-build

# Test basic functionality
./out/waffles --help
./out/waffles version
```

### Task Validation Commands:
Each task includes specific validation steps, but here are general verification commands:

```bash
# Check build status
make local-build
make local-test

# Check code quality
make local-vet

# Check dependencies
go mod verify
go mod tidy
```

### Common Restart Scenarios:

**Starting from scratch**: Begin with Task 1.1
**After partial completion**: Check the task's acceptance criteria and re-run validation
**After build failures**: Run `go mod tidy`, check for missing imports, verify syntax
**After test failures**: Review test output, fix failing tests, ensure mocks are correct
**After integration issues**: Verify external tools are available, check command execution

---

## Phase 1: Foundation and Cleanup (High Priority)

### Task 1.1: Remove Golang-Starter Artifacts
**Objective**: Clean up the codebase from golang-starter template artifacts
**Files to Modify**: Various files throughout the project
**Deliverables**:
- Delete [`internal/starter/`](internal/starter/) directory entirely
- Remove all references to "starter" in code, comments, and documentation
- Update any import paths referencing starter components
- Clean up starter-specific configuration code

**Acceptance Criteria**:
- No files or directories contain "starter" in their name or path
- No code references "starter" in imports, comments, or variable names
- Project structure follows clean Go conventions with [`cmd/`](cmd/), [`pkg/`](pkg/), and [`internal/`](internal/)
- All existing functionality remains intact after cleanup

### Task 1.2: Replace Viper with Environment-based Configuration
**Objective**: Replace Viper configuration system with godotenv and env libraries
**Files to Create/Modify**: 
- [`pkg/config/config.go`](pkg/config/config.go)
- [`go.mod`](go.mod)
- Remove existing Viper dependencies

**Deliverables**:
1. Add dependencies to [`go.mod`](go.mod):
   ```go
   github.com/joho/godotenv
   github.com/caarlos0/env/v8
   ```

2. Create [`pkg/config/config.go`](pkg/config/config.go) with:
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

3. Implement configuration loading functions:
   - `LoadConfig() (*Config, error)` - Load configuration with precedence
   - `LoadFromEnv(cfg *Config) error` - Load from environment variables
   - `LoadFromFile(cfg *Config, filepath string) error` - Load from .env file
   - `MergeWithDefaults(cfg *Config)` - Apply default values

**Acceptance Criteria**:
- Configuration loads from multiple sources in correct precedence order
- All environment variables parse correctly into struct fields
- .env file loading works for both local and global config files
- Command-line flags can override environment/file configuration
- No dependency on Viper library remains

### Task 1.3: Update CLI Structure with Cobra
**Objective**: Set up Cobra CLI framework with all required commands and flags
**Files to Create/Modify**:
- [`cmd/waffles/root.go`](cmd/waffles/root.go)
- [`cmd/waffles/deps.go`](cmd/waffles/deps.go)
- [`cmd/waffles/setup.go`](cmd/waffles/setup.go)
- [`cmd/waffles/query.go`](cmd/waffles/query.go)
- [`cmd/waffles/export.go`](cmd/waffles/export.go)
- [`cmd/waffles/version.go`](cmd/waffles/version.go)

**Deliverables**:
1. Update [`cmd/waffles/root.go`](cmd/waffles/root.go) with:
   - Primary command: `waffles [options] [prompt-search-terms...]`
   - All required flags as specified in requirements
   - Integration with new config system
   - Proper help text and usage examples

2. Create subcommands:
   - `deps`: Dependency management functionality
   - `setup`: Interactive setup wizard
   - `query`: Query logged conversations
   - `export`: Export logged data
   - `version`: Version information

3. Implement flag binding to configuration system
   - All CLI flags should override corresponding config values
   - Support for tool-specific argument passthrough
   - Proper validation and error handling

**Acceptance Criteria**:
- All required CLI flags are implemented and functional
- Subcommands provide appropriate help and usage information
- Configuration integration works correctly with flag precedence
- All commands can be invoked without errors (even if not fully implemented)

## Phase 2: Core System Components (High Priority)

### Task 2.1: Implement Dependency Management System
**Objective**: Create the dependency detection and management system
**Files to Create**:
- [`pkg/deps/detector.go`](pkg/deps/detector.go)
- [`pkg/deps/installer.go`](pkg/deps/installer.go)
- [`pkg/deps/types.go`](pkg/deps/types.go)

**Deliverables**:
1. Create [`pkg/deps/types.go`](pkg/deps/types.go):
   ```go
   type Dependency struct {
       Name         string
       Command      string
       MinVersion   string
       CheckCommand string
       InstallURL   string
       Plugins      []string
   }
   
   type DependencyStatus struct {
       Name      string
       Installed bool
       Version   string
       Valid     bool
       Plugins   []PluginStatus
   }
   
   type PluginStatus struct {
       Name      string
       Installed bool
   }
   ```

2. Create [`pkg/deps/detector.go`](pkg/deps/detector.go):
   - `CheckDependency(dep Dependency) (*DependencyStatus, error)`
   - `CheckAllDependencies() ([]DependencyStatus, error)`
   - `CheckVersion(command, minVersion string) (bool, string, error)`
   - `CheckLLMPlugins() ([]PluginStatus, error)`

3. Create [`pkg/deps/installer.go`](pkg/deps/installer.go):
   - `InstallDependency(dep Dependency) error`
   - `InstallLLMPlugin(plugin string) error`
   - `GetInstallInstructions(dep Dependency) string`
   - `AutoInstallAll() error`

**Acceptance Criteria**:
- Can detect presence of wheresmyprompt, files2prompt, and llm in PATH
- Version checking works correctly for all tools
- LLM plugin detection identifies required plugins
- Auto-installation works for supported platforms
- Clear error messages and installation instructions when dependencies missing

### Task 2.2: Implement Repository Analysis System
**Objective**: Create the repository language detection and file analysis system
**Files to Create**:
- [`pkg/repo/detector.go`](pkg/repo/detector.go)
- [`pkg/repo/patterns.go`](pkg/repo/patterns.go)
- [`pkg/repo/types.go`](pkg/repo/types.go)

**Deliverables**:
1. Create [`pkg/repo/types.go`](pkg/repo/types.go):
   ```go
   type Language string
   
   const (
       LanguageGo     Language = "go"
       LanguagePython Language = "python"
       LanguageUnknown Language = "unknown"
   )
   
   type RepositoryInfo struct {
       Language        Language
       RootPath        string
       IncludePatterns []string
       ExcludePatterns []string
       DetectedFiles   []FileInfo
   }
   
   type FileInfo struct {
       Path     string
       Size     int64
       Included bool
       Reason   string
   }
   ```

2. Create [`pkg/repo/detector.go`](pkg/repo/detector.go):
   - `DetectLanguage(path string) (Language, error)`
   - `AnalyzeRepository(path string, overrides *RepositoryOverrides) (*RepositoryInfo, error)`
   - `ScanFiles(path string, patterns []string) ([]FileInfo, error)`
   - `ApplyGitignore(files []FileInfo, path string) []FileInfo`

3. Create [`pkg/repo/patterns.go`](pkg/repo/patterns.go):
   - Language-specific inclusion/exclusion patterns
   - `GetPatternsForLanguage(lang Language) (include, exclude []string)`
   - `ApplyPatterns(files []string, include, exclude []string) []string`

**Acceptance Criteria**:
- Correctly detects Go projects (go.mod, *.go files)
- Correctly detects Python projects (requirements.txt, *.py files)
- Generates appropriate file patterns for each language
- Respects .gitignore rules unless overridden
- Handles mixed-language projects with prioritization
- Supports manual overrides via CLI flags

### Task 2.3: Implement Pipeline Orchestration System
**Objective**: Create the tool execution pipeline
**Files to Create**:
- [`pkg/pipeline/executor.go`](pkg/pipeline/executor.go)
- [`pkg/pipeline/types.go`](pkg/pipeline/types.go)
- [`pkg/pipeline/tools.go`](pkg/pipeline/tools.go)

**Deliverables**:
1. Create [`pkg/pipeline/types.go`](pkg/pipeline/types.go):
   ```go
   type Pipeline struct {
       Config    *config.Config
       RepoInfo  *repo.RepositoryInfo
       LogDB     *logging.Database
   }
   
   type ExecutionContext struct {
       ID              string
       StartTime       time.Time
       PromptQuery     string
       Files           []string
       ExecutionSteps  []StepResult
   }
   
   type StepResult struct {
       Tool      string
       Command   []string
       Output    string
       Error     error
       Duration  time.Duration
   }
   ```

2. Create [`pkg/pipeline/executor.go`](pkg/pipeline/executor.go):
   - `NewPipeline(config *config.Config) *Pipeline`
   - `Execute(promptQuery string, args []string) (*ExecutionContext, error)`
   - `executeWheresmyprompt(query string, args []string) (*StepResult, error)`
   - `executeFiles2prompt(repoInfo *repo.RepositoryInfo, args []string) (*StepResult, error)`
   - `executeLLM(prompt, context, model string, args []string) (*StepResult, error)`

3. Create [`pkg/pipeline/tools.go`](pkg/pipeline/tools.go):
   - `buildWheresmypromptCommand(query string, args []string) []string`
   - `buildFiles2promptCommand(repoInfo *repo.RepositoryInfo, args []string) []string`
   - `buildLLMCommand(prompt, context, model string, args []string) []string`
   - `sanitizeArgs(args []string) []string`

**Acceptance Criteria**:
- Sequential execution works: wheresmyprompt → files2prompt → llm
- Output from each tool correctly pipes to next tool
- Command injection prevention through proper argument sanitization
- Error handling with graceful degradation
- Execution timing and performance metrics
- Support for tool-specific argument passthrough

### Task 2.4: Implement Logging and Database System
**Objective**: Create the SQLite logging system with Waffles extensions
**Files to Create**:
- [`pkg/logging/database.go`](pkg/logging/database.go)
- [`pkg/logging/schema.go`](pkg/logging/schema.go)
- [`pkg/logging/types.go`](pkg/logging/types.go)
- [`pkg/logging/query.go`](pkg/logging/query.go)

**Deliverables**:
1. Create [`pkg/logging/types.go`](pkg/logging/types.go):
   ```go
   type Database struct {
       path string
       db   *sql.DB
   }
   
   type WafflesExecution struct {
       ID                  string
       ConversationID      string
       CommandArgs         string
       WheresmypromptQuery string
       Files2promptArgs    string
       DetectedLanguage    string
       FileCount           int
       ExecutionTimeMS     int64
       Created             int64
   }
   
   type WafflesFile struct {
       ID              string
       ExecutionID     string
       FilePath        string
       FileSize        int64
       Included        bool
       ExclusionReason string
   }
   ```

2. Create [`pkg/logging/database.go`](pkg/logging/database.go):
   - `NewDatabase(path string) (*Database, error)`
   - `InitializeSchema() error`
   - `LogExecution(exec *WafflesExecution) error`
   - `LogFiles(executionID string, files []WafflesFile) error`
   - `Close() error`

3. Create [`pkg/logging/schema.go`](pkg/logging/schema.go):
   - SQL schema definitions for Waffles-specific tables
   - Migration functions for schema updates
   - Index creation for performance

4. Create [`pkg/logging/query.go`](pkg/logging/query.go):
   - `QueryExecutions(filters map[string]interface{}) ([]WafflesExecution, error)`
   - `QueryFiles(executionID string) ([]WafflesFile, error)`
   - `ExportExecutions(format string, output io.Writer) error`

**Acceptance Criteria**:
- SQLite database creates correctly with Waffles tables
- Integrates with existing llm CLI database schema
- All execution context logged comprehensively
- Query interface works for filtering and searching
- Export functionality supports multiple formats
- Database handles concurrent access safely

## Phase 3: Advanced Features (Medium Priority)

### Task 3.1: Implement Auto-Installation System
**Objective**: Create automated dependency installation
**Files to Modify/Create**:
- [`pkg/deps/installer.go`](pkg/deps/installer.go) (extend)
- [`pkg/deps/platforms.go`](pkg/deps/platforms.go) (new)

**Deliverables**:
1. Extend [`pkg/deps/installer.go`](pkg/deps/installer.go):
   - Platform detection (Linux, macOS, Windows)
   - Package manager detection (Homebrew, apt, pip, etc.)
   - Automated download and installation scripts
   - Progress indication for long downloads

2. Create [`pkg/deps/platforms.go`](pkg/deps/platforms.go):
   - Platform-specific installation methods
   - Binary download and verification
   - Path manipulation and environment setup

**Acceptance Criteria**:
- Detects user's platform and package managers
- Downloads and installs wheresmyprompt and files2prompt binaries
- Installs Python pip packages including llm and plugins
- Provides progress feedback during installation
- Verifies installations post-completion
- Fails gracefully with manual instructions

### Task 3.2: Implement Interactive Setup Wizard
**Objective**: Create user-friendly setup experience
**Files to Create**:
- [`internal/setup/wizard.go`](internal/setup/wizard.go)
- [`internal/setup/prompts.go`](internal/setup/prompts.go)

**Deliverables**:
1. Create [`internal/setup/wizard.go`](internal/setup/wizard.go):
   - Interactive dependency checking and installation
   - Configuration file generation
   - API key setup guidance
   - LLM provider configuration

2. Create [`internal/setup/prompts.go`](internal/setup/prompts.go):
   - User input collection and validation
   - Configuration option explanations
   - Progress indication and status updates

**Acceptance Criteria**:
- Guides user through complete setup process
- Checks and installs dependencies interactively
- Creates proper configuration files
- Validates setup at completion
- Provides clear next steps and usage examples

### Task 3.3: Implement Query and Export System
**Objective**: Create comprehensive data query and export capabilities
**Files to Create**:
- [`internal/query/engine.go`](internal/query/engine.go)
- [`internal/export/formats.go`](internal/export/formats.go)
- [`internal/export/exporters.go`](internal/export/exporters.go)

**Deliverables**:
1. Create [`internal/query/engine.go`](internal/query/engine.go):
   - Complex query building and execution
   - Filter by date, model, language, file count, etc.
   - Full-text search in prompts and responses
   - Statistical queries (usage patterns, performance)

2. Create [`internal/export/formats.go`](internal/export/formats.go):
   - JSON export format
   - CSV export format
   - Markdown report format
   - Custom template support

3. Create [`internal/export/exporters.go`](internal/export/exporters.go):
   - Export execution with progress indication
   - Large dataset streaming for memory efficiency
   - Compression options for large exports

**Acceptance Criteria**:
- Complex filtering and searching works correctly
- Multiple export formats produce valid output
- Large datasets export without memory issues
- Progress indication for long-running exports
- Export files are properly formatted and readable

## Phase 4: Testing, Documentation, and Polish (Medium-Low Priority)

### Task 4.1: Implement Comprehensive Testing Suite
**Objective**: Create thorough test coverage for all components
**Files to Create**:
- [`pkg/config/config_test.go`](pkg/config/config_test.go)
- [`pkg/deps/detector_test.go`](pkg/deps/detector_test.go)
- [`pkg/repo/detector_test.go`](pkg/repo/detector_test.go)
- [`pkg/pipeline/executor_test.go`](pkg/pipeline/executor_test.go)
- [`pkg/logging/database_test.go`](pkg/logging/database_test.go)
- [`integration_test.go`](integration_test.go)

**Deliverables**:
1. Unit tests for all packages:
   - Mock external dependencies
   - Test error handling paths
   - Validate configuration loading
   - Test repository detection logic

2. Integration tests:
   - End-to-end workflow testing
   - Real external tool integration
   - Database integration testing
   - CLI interface testing

3. Test utilities and fixtures:
   - Mock repositories for testing
   - Test configuration files
   - Database test fixtures

**Acceptance Criteria**:
- Minimum 80% test coverage across all packages
- All error paths tested and covered
- Integration tests pass with real tools
- Tests run reliably in CI environment
- Performance benchmarks included

### Task 4.2: Create User Documentation
**Objective**: Comprehensive user-facing documentation
**Files to Create**:
- [`docs/installation.md`](docs/installation.md)
- [`docs/configuration.md`](docs/configuration.md)
- [`docs/usage.md`](docs/usage.md)
- [`docs/troubleshooting.md`](docs/troubleshooting.md)
- Update [`README.md`](README.md)

**Deliverables**:
1. Installation guide:
   - System requirements
   - Installation methods (binary, package managers, from source)
   - Dependency setup instructions
   - Verification steps

2. Configuration guide:
   - Environment variable reference
   - .env file examples
   - Profile configuration examples
   - Tool-specific configuration

3. Usage guide:
   - CLI command reference
   - Common workflows and examples
   - Advanced usage patterns
   - Integration with development workflows

4. Troubleshooting guide:
   - Common error messages and solutions
   - Dependency issues resolution
   - Performance optimization tips
   - Debug mode usage

**Acceptance Criteria**:
- Documentation covers all major features
- Examples are tested and working
- Clear installation instructions for all platforms
- Troubleshooting guide addresses common issues
- Documentation is well-organized and searchable

### Task 4.3: Create Developer Documentation
**Objective**: Documentation for contributors and developers
**Files to Create**:
- [`docs/architecture.md`](docs/architecture.md)
- [`docs/contributing.md`](docs/contributing.md)
- [`docs/api.md`](docs/api.md)
- Add comprehensive godoc comments

**Deliverables**:
1. Architecture documentation:
   - System overview and component interactions
   - Design decision explanations
   - Extension points for plugins
   - Performance considerations

2. Contributing guide:
   - Development setup instructions
   - Code style and standards
   - Testing requirements
   - Pull request process

3. API documentation:
   - Public package interfaces
   - Usage examples for each package
   - Integration patterns for external tools

**Acceptance Criteria**:
- All public APIs documented with godoc
- Architecture decisions clearly explained
- Contributing guide enables new contributors
- Code examples compile and work correctly

## Implementation Guidelines

### Code Quality Standards
- All code must be formatted with `gofmt`
- All public APIs must have godoc documentation
- Error handling must be comprehensive and informative
- Input validation must prevent security issues
- Concurrent code must be thread-safe

### Testing Requirements
- Unit tests for all public functions
- Integration tests for external tool interactions
- Error path testing for all error conditions
- Performance benchmarks for critical paths
- Cross-platform compatibility testing

### Security Considerations
- Input sanitization for external tool execution
- API key handling best practices
- File path validation to prevent traversal attacks
- Resource limits to prevent DoS conditions
- Secure defaults for all configuration options

### Performance Requirements
- Handle repositories with 1000+ files efficiently
- External tool execution timeouts
- Database operations optimized with indexes
- Memory usage bounded for large operations
- Concurrent processing where thread-safe

## Task Dependencies

### Critical Path
1. Task 1.1 → Task 1.2 → Task 1.3 (Foundation)
2. Task 2.1 → Task 2.2 → Task 2.3 → Task 2.4 (Core System)
3. Task 4.1 (Testing throughout development)

### Parallel Development
- Task 3.1, 3.2, 3.3 can be developed in parallel after Phase 2
- Documentation tasks (4.2, 4.3) can be developed alongside implementation
- Build system (5.1, 5.2) can be developed independently

### Validation Checkpoints
- After Phase 1: Basic CLI structure and configuration working
- After Phase 2: End-to-end pipeline execution working
- After Phase 3: All features implemented and tested
- After Phase 4: Production-ready with documentation
- After Phase 5: Distributed and packaged

---

This task list provides comprehensive guidance for implementing the complete Waffles application. Each task is designed to be independently implementable by an AI assistant while maintaining system coherence and quality standards.