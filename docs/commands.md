# Commands Reference

Complete reference for all Waffles commands, flags, and options.

## Table of Contents

- [Global Flags](#global-flags)
- [waffles query](#waffles-query)
- [waffles setup](#waffles-setup)  
- [waffles deps](#waffles-deps)
- [waffles export](#waffles-export)
- [waffles config](#waffles-config)
- [waffles version](#waffles-version)

## Global Flags

These flags are available for all commands:

| Flag | Description | Default |
|------|-------------|---------|
| `--help, -h` | Show help information | |
| `--verbose, -v` | Enable verbose output | `false` |
| `--config string` | Config file path | `~/.config/waffles/.env` |
| `--no-color` | Disable colored output | `false` |

## waffles query

Execute AI queries with automatic context gathering.

### Syntax
```bash
waffles query [flags] "prompt text"
```

### Flags

#### Model and Provider
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--model, -m string` | LLM model to use | `claude-3-sonnet` | `--model gpt-4` |
| `--provider, -p string` | LLM provider | `anthropic` | `--provider openai` |

#### File Selection
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--include string` | File patterns to include | _(auto)_ | `--include "*.go,*.md"` |
| `--exclude string` | File patterns to exclude | _(auto)_ | `--exclude "vendor/*"` |
| `--language string` | Override language detection | _(auto)_ | `--language python` |
| `--no-gitignore` | Ignore .gitignore rules | `false` | `--no-gitignore` |

#### Execution Control
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--dry-run` | Show what would be executed | `false` | `--dry-run` |
| `--timeout int` | Total timeout in seconds | `300` | `--timeout 600` |
| `--skip-deps` | Skip dependency checks | `false` | `--skip-deps` |

#### Tool Arguments
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--wheresmyprompt-args string` | Custom wheresmyprompt arguments | | `--wheresmyprompt-args "--source custom"` |
| `--files2prompt-args string` | Custom files2prompt arguments | | `--files2prompt-args "--max-tokens 4000"` |
| `--llm-args string` | Custom LLM arguments | | `--llm-args "--temperature 0.1"` |

#### Output Control
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--output, -o string` | Output file path | _(stdout)_ | `--output result.txt` |
| `--format string` | Output format | `text` | `--format json` |
| `--quiet, -q` | Suppress non-essential output | `false` | `--quiet` |

### Examples

```bash
# Basic query
waffles query "What does this codebase do?"

# Specific model and provider
waffles query --model gpt-4 --provider openai "Review this code"

# Custom file selection
waffles query --include "*.go,*.yaml" --exclude "vendor/*" "Analyze configuration"

# Override language detection
waffles query --language python "Review this Python project"

# Dry run to see execution plan
waffles query --dry-run "Test prompt"

# Custom output format and file
waffles query --format json --output analysis.json "Generate analysis report"

# Custom tool arguments
waffles query --llm-args "--temperature 0.1 --max-tokens 2000" "Precise code review"
```

## waffles setup

Interactive setup wizard for initial configuration.

### Syntax
```bash
waffles setup [flags]
```

### Flags

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--auto-install` | Automatically install dependencies | `false` | `--auto-install` |
| `--model string` | Default model to configure | | `--model claude-3-sonnet` |
| `--provider string` | Default provider to configure | | `--provider anthropic` |
| `--reset` | Reset existing configuration | `false` | `--reset` |
| `--non-interactive` | Run without prompts | `false` | `--non-interactive` |
| `--db-path string` | Database file path | | `--db-path ~/custom/waffles.db` |

### Examples

```bash
# Interactive setup
waffles setup

# Automated setup with specific model
waffles setup --auto-install --model gpt-4 --provider openai

# Reset and reconfigure
waffles setup --reset

# Non-interactive setup (for CI/CD)
waffles setup --non-interactive --model claude-3-sonnet --provider anthropic
```

## waffles deps

Manage dependencies and system requirements.

### Subcommands

#### waffles deps check
Check status of all dependencies.

```bash
waffles deps check [flags] [dependency-name]
```

**Flags:**
| Flag | Description | Default |
|------|-------------|---------|
| `--json` | Output in JSON format | `false` |
| `--verbose` | Show detailed information | `false` |

**Examples:**
```bash
# Check all dependencies
waffles deps check

# Check specific dependency
waffles deps check wheresmyprompt

# JSON output
waffles deps check --json
```

#### waffles deps install
Install missing dependencies.

```bash
waffles deps install [flags] [dependency-name]
```

**Flags:**
| Flag | Description | Default |
|------|-------------|---------|
| `--dry-run` | Show installation instructions only | `false` |
| `--force` | Force reinstallation | `false` |
| `--skip-verification` | Skip post-install verification | `false` |

**Examples:**
```bash
# Install all missing dependencies
waffles deps install

# Install specific dependency
waffles deps install llm

# Show installation instructions only
waffles deps install --dry-run

# Force reinstall
waffles deps install --force wheresmyprompt
```

#### waffles deps info
Show detailed dependency information.

```bash
waffles deps info [flags]
```

**Examples:**
```bash
# Show system information
waffles deps info

# Verbose system details
waffles deps info --verbose
```

## waffles export

Export execution history and analytics.

### Syntax
```bash
waffles export [flags]
```

### Flags

#### Output Control
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--format, -f string` | Export format | `json` | `--format csv` |
| `--output, -o string` | Output file path | _(stdout)_ | `--output report.json` |
| `--pretty` | Pretty-print output | `false` | `--pretty` |

#### Filtering
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--days int` | Filter by days back | | `--days 7` |
| `--since string` | Filter since date | | `--since "2024-01-01"` |
| `--until string` | Filter until date | | `--until "2024-12-31"` |
| `--language string` | Filter by language | | `--language go` |
| `--model string` | Filter by model used | | `--model gpt-4` |
| `--provider string` | Filter by provider | | `--provider openai` |
| `--success-only` | Include only successful executions | `false` | `--success-only` |
| `--failures-only` | Include only failed executions | `false` | `--failures-only` |

#### Data Selection
| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--include-files` | Include file details | `false` | `--include-files` |
| `--include-steps` | Include step details | `false` | `--include-steps` |
| `--limit int` | Limit number of results | | `--limit 100` |

### Export Formats

#### JSON Format
```bash
# Standard JSON export
waffles export --format json

# Pretty-printed JSON
waffles export --format json --pretty --output report.json
```

#### CSV Format  
```bash
# CSV for spreadsheet analysis
waffles export --format csv --output usage.csv

# CSV with file details
waffles export --format csv --include-files --output detailed.csv
```

#### Markdown Format
```bash
# Human-readable report
waffles export --format markdown --output report.md

# Report with step details
waffles export --format markdown --include-steps --output detailed-report.md
```

#### SQL Format
```bash
# SQL INSERT statements
waffles export --format sql --output data.sql
```

### Examples

```bash
# Export last 7 days as JSON
waffles export --days 7 --format json --output weekly.json

# Export successful Go queries only
waffles export --language go --success-only --format csv

# Export detailed report with files and steps
waffles export --include-files --include-steps --format markdown --output full-report.md

# Export recent failures for debugging
waffles export --days 1 --failures-only --format json --pretty
```

## waffles config

Manage configuration settings.

### Subcommands

#### waffles config show
Display current configuration.

```bash
waffles config show [flags] [key]
```

**Flags:**
| Flag | Description | Default |
|------|-------------|---------|
| `--effective` | Show effective config after overrides | `false` |
| `--json` | Output in JSON format | `false` |

**Examples:**
```bash
# Show all configuration
waffles config show

# Show specific setting
waffles config show WAFFLES_DEFAULT_MODEL

# Show effective configuration
waffles config show --effective

# JSON output
waffles config show --json
```

#### waffles config set
Set configuration values.

```bash
waffles config set [flags] <key> <value>
```

**Examples:**
```bash
# Set default model
waffles config set WAFFLES_DEFAULT_MODEL gpt-4

# Set database path
waffles config set WAFFLES_DB_PATH /custom/path/waffles.db
```

#### waffles config validate
Validate current configuration.

```bash
waffles config validate [flags]
```

**Examples:**
```bash
# Validate configuration
waffles config validate

# Validate with verbose output
waffles config validate --verbose
```

#### waffles config reset
Reset configuration to defaults.

```bash
waffles config reset [flags] [key]
```

**Flags:**
| Flag | Description | Default |
|------|-------------|---------|
| `--all` | Reset all settings | `false` |
| `--confirm` | Skip confirmation prompt | `false` |

**Examples:**
```bash
# Reset specific setting
waffles config reset WAFFLES_DEFAULT_MODEL

# Reset all settings
waffles config reset --all --confirm
```

## waffles version

Display version information.

### Syntax
```bash
waffles version [flags]
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--short` | Show version number only | `false` |
| `--json` | Output in JSON format | `false` |

### Examples

```bash
# Show version information
waffles version

# Short version only
waffles version --short

# JSON format with build details
waffles version --json
```

## Command Combinations

### Common Workflows

#### Setup and First Use
```bash
# Complete setup workflow
waffles setup --auto-install
waffles deps check
waffles query "What does this project do?"
```

#### Daily Development
```bash
# Quick code review
waffles query --model gpt-4 "Review recent changes"

# Focused analysis
waffles query --include "src/*" --exclude "*.test.*" "Analyze core logic"
```

#### Debugging and Troubleshooting
```bash
# Verbose dry run
waffles query --dry-run --verbose "Test query"

# Check system status
waffles deps check --verbose
waffles config validate
```

#### Analytics and Reporting
```bash
# Weekly usage report
waffles export --days 7 --format csv --output weekly-usage.csv

# Success rate analysis  
waffles export --format json | jq '[.[] | .success] | add/length'
```

## Exit Codes

Waffles uses standard exit codes:

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Invalid command line arguments |
| `3` | Configuration error |
| `4` | Dependency error |
| `5` | Pipeline execution error |
| `6` | Database error |

## Environment Variable Integration

Most command flags have corresponding environment variables:

```bash
# Command flag equivalents
--model gpt-4          → WAFFLES_DEFAULT_MODEL=gpt-4
--provider openai      → WAFFLES_DEFAULT_PROVIDER=openai  
--include "*.go"       → WAFFLES_INCLUDE_PATTERNS="*.go"
--exclude "vendor/*"   → WAFFLES_EXCLUDE_PATTERNS="vendor/*"
--verbose              → WAFFLES_VERBOSE=true
```

This allows for flexible configuration via environment variables, config files, or command-line flags.

## Next Steps

- **[Usage Guide](usage.md)**: Learn effective usage patterns
- **[Configuration](configuration.md)**: Detailed configuration options
- **[Examples](examples.md)**: Real-world usage examples
- **[Troubleshooting](troubleshooting.md)**: Common issues and solutions