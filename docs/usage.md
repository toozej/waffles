# Usage Guide

This guide covers how to use Waffles effectively for AI-assisted development workflows.

## Table of Contents

- [Basic Usage](#basic-usage)
- [Core Commands](#core-commands)
- [Query Operations](#query-operations)
- [Configuration Management](#configuration-management)
- [Data Export](#data-export)
- [Advanced Features](#advanced-features)
- [Best Practices](#best-practices)

## Basic Usage

### First-Time Setup

```bash
# Run the interactive setup wizard
waffles setup

# Setup with automatic dependency installation
waffles setup --auto-install

# Non-interactive setup
waffles setup --model claude-3-sonnet --provider anthropic --auto-install
```

### Simple Queries

```bash
# Basic query - uses project context automatically
waffles query "What does this codebase do?"

# Query with specific model
waffles query --model gpt-4 "Review this code for security issues"

# Query with custom provider
waffles query --provider openai "Suggest performance improvements"
```

### Checking System Status

```bash
# Check all dependencies
waffles deps check

# Check specific dependency
waffles deps check wheresmyprompt

# Show system information
waffles deps info
```

## Core Commands

### `waffles query`

Execute AI queries with automatic context gathering.

```bash
# Basic syntax
waffles query [flags] "your prompt here"

# Common flags
--model string         # LLM model to use
--provider string      # LLM provider  
--include string       # File patterns to include
--exclude string       # File patterns to exclude
--language string      # Override detected language
--dry-run             # Show what would be executed
--verbose             # Show detailed output
```

**Examples:**

```bash
# Architecture analysis
waffles query "Analyze the architecture of this codebase"

# Code review with specific focus
waffles query --include "*.go,*.yaml" "Review the Go code and configuration"

# Security audit
waffles query --model gpt-4 "Perform a security audit of this application"

# Documentation request
waffles query "Generate comprehensive documentation for this project"
```

### `waffles setup`

Configure Waffles for your environment.

```bash
# Interactive setup
waffles setup

# Automated setup
waffles setup --auto-install --model claude-3-sonnet

# Reconfigure existing setup
waffles setup --reset
```

### `waffles deps`

Manage dependencies and system requirements.

```bash
# Check all dependencies
waffles deps check

# Install missing dependencies
waffles deps install

# Get installation instructions
waffles deps install --dry-run

# Show detailed dependency information
waffles deps info
```

### `waffles export`

Export execution history and analytics.

```bash
# Export recent executions
waffles export --format json --days 7

# Export with filters
waffles export --format csv --language go --success-only

# Export to file
waffles export --format markdown --output report.md
```

## Query Operations

### Context Control

Waffles automatically analyzes your repository and includes relevant files. You can control this behavior:

```bash
# Include specific patterns
waffles query --include "*.go,*.md,Dockerfile" "Analyze the deployment setup"

# Exclude unwanted files
waffles query --exclude "vendor/*,*.test.go" "Review the main code"

# Override language detection
waffles query --language python "Analyze this Python project"

# Ignore gitignore rules
waffles query --no-gitignore "Include all files including ignored ones"
```

### Model Selection

```bash
# Use specific model
waffles query --model gpt-4 "Complex architectural analysis"

# Use different provider
waffles query --provider openai --model gpt-3.5-turbo "Quick code review"

# List available models (if supported by LLM CLI)
llm models list
```

### Advanced Query Options

```bash
# Dry run to see what would happen
waffles query --dry-run "Test prompt"

# Verbose output for debugging
waffles query --verbose "Debug this execution"

# Custom timeout
waffles query --timeout 300 "Complex analysis that might take longer"

# Skip dependency checks (faster)
waffles query --skip-deps "Quick query on known good system"
```

## Configuration Management

### Environment Variables

Set persistent configuration:

```bash
# In ~/.bashrc or ~/.zshrc
export WAFFLES_DEFAULT_MODEL="claude-3-sonnet"
export WAFFLES_DEFAULT_PROVIDER="anthropic"
export WAFFLES_DB_PATH="$HOME/.config/waffles/waffles.db"
export WAFFLES_VERBOSE="false"
export WAFFLES_AUTO_INSTALL="true"
```

### Project-Specific Configuration

Create a `.env` file in your project:

```env
# .env in project root
WAFFLES_LANGUAGE_OVERRIDE=go
WAFFLES_INCLUDE_PATTERNS=*.go,*.md,go.mod
WAFFLES_EXCLUDE_PATTERNS=vendor/*,*_test.go
WAFFLES_DEFAULT_MODEL=gpt-4
```

### Configuration Validation

```bash
# Show current configuration
waffles config show

# Validate configuration
waffles config validate

# Reset to defaults
waffles config reset
```

## Data Export

### Export Formats

Waffles supports multiple export formats for analysis:

#### JSON Export
```bash
# Detailed structured data
waffles export --format json --output analysis.json
```

#### CSV Export  
```bash
# Spreadsheet-friendly format
waffles export --format csv --output usage-stats.csv
```

#### Markdown Export
```bash
# Human-readable reports
waffles export --format markdown --output report.md
```

#### SQL Export
```bash
# Database import format
waffles export --format sql --output data.sql
```

### Filtering and Selection

```bash
# Filter by language
waffles export --language go --format csv

# Filter by date range
waffles export --days 30 --format json

# Filter by success status
waffles export --success-only --format markdown

# Filter by model used
waffles export --model gpt-4 --format csv

# Combined filters
waffles export --language python --days 7 --success-only --format json
```

### Analytics Queries

```bash
# Usage statistics
waffles export --format json --days 30 | jq '.[] | .model_used' | sort | uniq -c

# Success rates by language
waffles export --format csv --days 7 | cut -d, -f4,6 | sort | uniq -c

# Average execution times
waffles export --format json | jq '.[] | .execution_time_ms' | awk '{sum+=$1; count++} END {print sum/count}'
```

## Advanced Features

### Pipeline Customization

```bash
# Custom arguments for tools
waffles query --wheresmyprompt-args "--source custom" "Custom prompt source"

# Custom files2prompt arguments
waffles query --files2prompt-args "--max-tokens 4000" "Large context analysis"

# Custom LLM arguments
waffles query --llm-args "--temperature 0.1" "Precise code analysis"
```

### Integration with Development Workflow

#### Git Hooks Integration

```bash
# Pre-commit hook example
#!/bin/bash
# .git/hooks/pre-commit
waffles query "Review these changes for potential issues" --include "$(git diff --cached --name-only | tr '\n' ',')"
```

#### CI/CD Integration

```bash
# In your CI pipeline
- name: AI Code Review
  run: |
    waffles query "Analyze this pull request for security and performance issues" \
      --format json --output code-review.json
```

#### IDE Integration

```bash
# VSCode task example (tasks.json)
{
    "label": "Waffles Code Review",
    "type": "shell", 
    "command": "waffles",
    "args": ["query", "Review the current file for improvements", "--include", "${file}"]
}
```

### Batch Operations

```bash
# Multiple queries in sequence
queries=(
    "What is the architecture of this project?"
    "Are there any security vulnerabilities?"
    "What are the performance bottlenecks?"
    "How can we improve code quality?"
)

for query in "${queries[@]}"; do
    echo "Executing: $query"
    waffles query "$query" --format json >> batch-results.jsonl
done
```

### Custom Workflows

#### Documentation Generation
```bash
#!/bin/bash
# generate-docs.sh
waffles query "Generate API documentation" --include "*.go" --exclude "*_test.go" \
    --output api-docs.md --format markdown

waffles query "Create user guide" --include "README.md,docs/*" \
    --output user-guide.md --format markdown
```

#### Code Quality Assessment
```bash
#!/bin/bash
# quality-check.sh
waffles query "Assess code quality and suggest improvements" \
    --language go --exclude "vendor/*" \
    --model gpt-4 --output quality-report.json --format json
```

## Best Practices

### Query Design

1. **Be Specific**: Clear, specific prompts yield better results
   ```bash
   # Good
   waffles query "Review the authentication middleware for security vulnerabilities"
   
   # Less effective
   waffles query "check security"
   ```

2. **Use Context Control**: Include/exclude relevant files
   ```bash
   # Focus on specific components
   waffles query --include "auth/*,middleware/*" "Review authentication system"
   ```

3. **Choose Appropriate Models**: Match model to task complexity
   ```bash
   # Complex architectural analysis
   waffles query --model gpt-4 "Design system architecture review"
   
   # Simple code explanation
   waffles query --model gpt-3.5-turbo "Explain this function"
   ```

### Performance Optimization

1. **Use Appropriate Timeouts**: Adjust for query complexity
   ```bash
   waffles query --timeout 60 "Quick analysis"
   waffles query --timeout 300 "Comprehensive review"
   ```

2. **Filter Files Effectively**: Reduce context size
   ```bash
   # Exclude unnecessary files
   waffles query --exclude "*.log,tmp/*,vendor/*" "Code review"
   ```

3. **Skip Dependencies When Safe**: Faster execution
   ```bash
   waffles query --skip-deps "Quick query on verified system"
   ```

### Workflow Integration

1. **Standardize Project Configs**: Use `.env` files for consistency
2. **Create Aliases**: For common operations
   ```bash
   alias wr="waffles query --model gpt-4"  # "waffles review"
   alias wa="waffles query --verbose"       # "waffles analyze"
   ```

3. **Regular Maintenance**: Clean up old data
   ```bash
   # Archive old executions
   waffles export --days 90 --format json --output archive-$(date +%Y%m%d).json
   ```

### Security Considerations

1. **Sensitive Data**: Be careful with confidential code
2. **API Keys**: Store securely, never commit
3. **Output Review**: Always review AI suggestions before applying

### Monitoring and Analytics

1. **Track Usage**: Regular exports for analysis
   ```bash
   # Weekly usage report
   waffles export --days 7 --format csv --output weekly-usage.csv
   ```

2. **Monitor Success Rates**: Identify optimization opportunities
   ```bash
   # Success rate analysis
   waffles export --format json | jq '[.[] | .success] | add/length'
   ```

3. **Performance Tracking**: Monitor execution times
   ```bash
   # Average execution time
   waffles export --format json | jq '[.[] | .execution_time_ms] | add/length'
   ```

## Next Steps

- **[Configuration](configuration.md)**: Detailed configuration options
- **[Commands Reference](commands.md)**: Complete command documentation  
- **[Examples](examples.md)**: Real-world usage examples
- **[Troubleshooting](troubleshooting.md)**: Common issues and solutions