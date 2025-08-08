# Configuration Guide

This guide covers all configuration options for Waffles, including environment variables, configuration files, and project-specific settings.

## Table of Contents

- [Configuration Hierarchy](#configuration-hierarchy)
- [Environment Variables](#environment-variables)
- [Configuration Files](#configuration-files)
- [Project-Specific Configuration](#project-specific-configuration)
- [LLM Configuration](#llm-configuration)
- [Advanced Settings](#advanced-settings)
- [Configuration Examples](#configuration-examples)

## Configuration Hierarchy

Waffles uses the following configuration priority (highest to lowest):

1. **Command-line flags** - Override everything else
2. **Environment variables** - System-wide settings
3. **Project .env file** - Project-specific overrides
4. **Global .env file** - User-specific defaults
5. **Built-in defaults** - Fallback values

## Environment Variables

### Core Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `WAFFLES_DEFAULT_MODEL` | Default LLM model | `claude-3-sonnet` | `gpt-4` |
| `WAFFLES_DEFAULT_PROVIDER` | Default LLM provider | `anthropic` | `openai` |
| `WAFFLES_DB_PATH` | Database file location | `~/.config/waffles/waffles.db` | `/custom/path/db.sqlite` |
| `WAFFLES_AUTO_INSTALL` | Auto-install dependencies | `false` | `true` |
| `WAFFLES_VERBOSE` | Enable verbose output | `false` | `true` |

### Repository Analysis

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `WAFFLES_LANGUAGE_OVERRIDE` | Force language detection | _(auto-detect)_ | `go` |
| `WAFFLES_INCLUDE_PATTERNS` | Default include patterns | _(language-specific)_ | `*.go,*.md` |
| `WAFFLES_EXCLUDE_PATTERNS` | Default exclude patterns | _(language-specific)_ | `vendor/*,*.test.go` |
| `WAFFLES_IGNORE_GITIGNORE` | Ignore .gitignore rules | `false` | `true` |
| `WAFFLES_MAX_FILES` | Maximum files to process | `1000` | `500` |
| `WAFFLES_MAX_FILE_SIZE` | Maximum file size (bytes) | `1048576` | `2097152` |

### Tool Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `WAFFLES_WHERESMYPROMPT_ARGS` | Custom wheresmyprompt args | _(empty)_ | `--source custom` |
| `WAFFLES_FILES2PROMPT_ARGS` | Custom files2prompt args | _(empty)_ | `--max-tokens 4000` |
| `WAFFLES_LLM_ARGS` | Custom LLM args | _(empty)_ | `--temperature 0.1` |

### Timeouts and Limits

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `WAFFLES_TIMEOUT_WHERESMYPROMPT` | wheresmyprompt timeout (seconds) | `30` | `60` |
| `WAFFLES_TIMEOUT_FILES2PROMPT` | files2prompt timeout (seconds) | `60` | `120` |
| `WAFFLES_TIMEOUT_LLM` | LLM timeout (seconds) | `180` | `300` |

## Configuration Files

### Global Configuration

Create `~/.config/waffles/.env` for user-wide defaults:

```env
# ~/.config/waffles/.env

# Core settings
WAFFLES_DEFAULT_MODEL=claude-3-sonnet
WAFFLES_DEFAULT_PROVIDER=anthropic
WAFFLES_VERBOSE=false
WAFFLES_AUTO_INSTALL=true

# Database
WAFFLES_DB_PATH=/Users/username/.config/waffles/waffles.db

# Repository analysis
WAFFLES_MAX_FILES=1000
WAFFLES_MAX_FILE_SIZE=1048576

# Tool timeouts
WAFFLES_TIMEOUT_WHERESMYPROMPT=30
WAFFLES_TIMEOUT_FILES2PROMPT=60
WAFFLES_TIMEOUT_LLM=180
```

### Project Configuration

Create `.env` in your project root for project-specific settings:

```env
# .env (in project root)

# Override language detection for mixed projects
WAFFLES_LANGUAGE_OVERRIDE=go

# Project-specific file patterns
WAFFLES_INCLUDE_PATTERNS=*.go,*.yaml,*.md,Dockerfile
WAFFLES_EXCLUDE_PATTERNS=vendor/*,*_test.go,tmp/*,.git/*

# Use specific model for this project
WAFFLES_DEFAULT_MODEL=gpt-4

# Custom tool arguments
WAFFLES_FILES2PROMPT_ARGS=--max-tokens 6000
WAFFLES_LLM_ARGS=--temperature 0.2
```

## Project-Specific Configuration

### Language-Specific Defaults

Waffles automatically applies sensible defaults based on detected language:

#### Go Projects
```env
# Automatic settings for Go projects
WAFFLES_INCLUDE_PATTERNS=*.go,go.mod,go.sum
WAFFLES_EXCLUDE_PATTERNS=*_test.go,vendor/*,.git/*,pkg/version/*
```

#### Python Projects
```env
# Automatic settings for Python projects  
WAFFLES_INCLUDE_PATTERNS=*.py,requirements.txt,pyproject.toml
WAFFLES_EXCLUDE_PATTERNS=*test*.py,__pycache__/*,venv/*,.git/*
```

#### JavaScript/Node.js Projects
```env
# Automatic settings for JS projects
WAFFLES_INCLUDE_PATTERNS=*.js,*.ts,package.json,*.json
WAFFLES_EXCLUDE_PATTERNS=node_modules/*,dist/*,*.test.js,.git/*
```

### Override Examples

#### Focus on Specific Directories
```env
# Only analyze source code directories
WAFFLES_INCLUDE_PATTERNS=src/**/*,lib/**/*,*.md
WAFFLES_EXCLUDE_PATTERNS=test/*,spec/*,docs/*
```

#### Include Documentation
```env
# Include docs in analysis
WAFFLES_INCLUDE_PATTERNS=*.go,*.md,docs/**/*.md
WAFFLES_EXCLUDE_PATTERNS=vendor/*,*_test.go
```

#### Exclude Large Files
```env
# Skip large generated files
WAFFLES_EXCLUDE_PATTERNS=*.pb.go,*_generated.go,vendor/*,*.min.js
WAFFLES_MAX_FILE_SIZE=500000
```

## LLM Configuration

### Model Configuration

Waffles integrates with the LLM CLI. Configure your LLM models separately:

```bash
# Add API keys
llm keys set openai
llm keys set anthropic

# List available models
llm models list

# Set default model globally
llm models default gpt-4

# Configure model aliases
llm aliases set review gpt-4
llm aliases set quick gpt-3.5-turbo
```

### Provider-Specific Settings

#### OpenAI Configuration
```bash
# Set API key
export OPENAI_API_KEY="your-key-here"

# Or using llm CLI
llm keys set openai
```

#### Anthropic Configuration
```bash
# Set API key
export ANTHROPIC_API_KEY="your-key-here"

# Or using llm CLI  
llm keys set anthropic
```

#### Local Models (Ollama)
```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Install models
ollama pull llama2
ollama pull codellama

# Configure in waffles
export WAFFLES_DEFAULT_PROVIDER=ollama
export WAFFLES_DEFAULT_MODEL=llama2
```

## Advanced Settings

### Database Configuration

#### Custom Database Location
```env
# Use custom database location
WAFFLES_DB_PATH=/path/to/custom/waffles.db

# Use temporary database (testing)
WAFFLES_DB_PATH=:memory:
```

#### Database Connection Options
```env
# SQLite-specific options via connection string
WAFFLES_DB_PATH=/path/to/db.sqlite?cache=shared&mode=rwc
```

### Performance Tuning

#### For Large Codebases
```env
# Increase limits for large projects
WAFFLES_MAX_FILES=5000
WAFFLES_MAX_FILE_SIZE=2097152
WAFFLES_TIMEOUT_FILES2PROMPT=300
```

#### For Fast Queries
```env
# Reduce limits for speed
WAFFLES_MAX_FILES=100
WAFFLES_MAX_FILE_SIZE=524288
WAFFLES_TIMEOUT_WHERESMYPROMPT=10
WAFFLES_TIMEOUT_FILES2PROMPT=30
```

### Security Settings

#### Sensitive Projects
```env
# Disable automatic installation for security
WAFFLES_AUTO_INSTALL=false

# Use specific, vetted models
WAFFLES_DEFAULT_MODEL=gpt-4
WAFFLES_DEFAULT_PROVIDER=openai

# Exclude sensitive files
WAFFLES_EXCLUDE_PATTERNS=*.key,*.pem,secrets/*,config/*.yaml
```

#### Air-Gapped Environments
```env
# Use local models only
WAFFLES_DEFAULT_PROVIDER=ollama
WAFFLES_DEFAULT_MODEL=codellama
WAFFLES_AUTO_INSTALL=false
```

## Configuration Examples

### Development Setup

```env
# ~/.config/waffles/.env - Development machine
WAFFLES_DEFAULT_MODEL=gpt-4
WAFFLES_DEFAULT_PROVIDER=openai
WAFFLES_VERBOSE=true
WAFFLES_AUTO_INSTALL=true
WAFFLES_MAX_FILES=2000
WAFFLES_TIMEOUT_LLM=300
```

### CI/CD Environment

```env
# CI environment configuration
WAFFLES_DEFAULT_MODEL=gpt-3.5-turbo
WAFFLES_DEFAULT_PROVIDER=openai
WAFFLES_VERBOSE=false
WAFFLES_AUTO_INSTALL=false
WAFFLES_MAX_FILES=500
WAFFLES_TIMEOUT_WHERESMYPROMPT=60
WAFFLES_TIMEOUT_FILES2PROMPT=120
WAFFLES_TIMEOUT_LLM=180
```

### Enterprise Setup

```env
# Enterprise configuration with local models
WAFFLES_DEFAULT_PROVIDER=ollama
WAFFLES_DEFAULT_MODEL=codellama:34b
WAFFLES_AUTO_INSTALL=false
WAFFLES_DB_PATH=/shared/waffles/database.db
WAFFLES_MAX_FILES=1000
WAFFLES_EXCLUDE_PATTERNS=*.key,*.pem,secrets/*,*.env
```

### Multi-Project Workspace

#### Workspace-Level Config (`~/.config/waffles/.env`)
```env
# Shared settings for all projects
WAFFLES_DEFAULT_PROVIDER=anthropic
WAFFLES_VERBOSE=false
WAFFLES_AUTO_INSTALL=true
```

#### Project-Specific Configs

**Go Project (`.env`)**
```env
WAFFLES_DEFAULT_MODEL=claude-3-sonnet
WAFFLES_LANGUAGE_OVERRIDE=go
WAFFLES_INCLUDE_PATTERNS=*.go,go.mod,go.sum,*.md
WAFFLES_EXCLUDE_PATTERNS=vendor/*,*_test.go
```

**Python Project (`.env`)**
```env
WAFFLES_DEFAULT_MODEL=gpt-4
WAFFLES_LANGUAGE_OVERRIDE=python
WAFFLES_INCLUDE_PATTERNS=*.py,requirements.txt,pyproject.toml
WAFFLES_EXCLUDE_PATTERNS=__pycache__/*,venv/*,*.pyc
```

**Frontend Project (`.env`)**
```env
WAFFLES_DEFAULT_MODEL=gpt-3.5-turbo
WAFFLES_LANGUAGE_OVERRIDE=javascript
WAFFLES_INCLUDE_PATTERNS=src/**/*.js,src/**/*.ts,*.json
WAFFLES_EXCLUDE_PATTERNS=node_modules/*,dist/*,build/*
```

## Configuration Validation

### Check Current Configuration
```bash
# Show all current settings
waffles config show

# Show specific setting
waffles config show WAFFLES_DEFAULT_MODEL

# Show effective configuration (after all overrides)
waffles config effective
```

### Validate Configuration
```bash
# Validate current configuration
waffles config validate

# Test configuration with dry run
waffles query --dry-run "test configuration"
```

### Reset Configuration
```bash
# Reset to defaults
waffles config reset

# Reset specific setting
waffles config reset WAFFLES_DEFAULT_MODEL
```

## Troubleshooting Configuration

### Common Issues

#### Environment Variables Not Loading
```bash
# Check if variables are set
env | grep WAFFLES_

# Source your shell profile
source ~/.bashrc
# or
source ~/.zshrc
```

#### Configuration File Not Found
```bash
# Check file location
ls -la ~/.config/waffles/.env

# Create directory if missing
mkdir -p ~/.config/waffles
```

#### Permission Issues
```bash
# Fix database directory permissions
chmod 755 ~/.config/waffles
chmod 644 ~/.config/waffles/.env
```

### Debug Configuration Loading
```bash
# Enable verbose output to see configuration loading
waffles --verbose config show

# Show configuration precedence
waffles config debug
```

## Next Steps

- **[Commands Reference](commands.md)**: Complete command documentation
- **[Usage Guide](usage.md)**: How to use Waffles effectively
- **[Examples](examples.md)**: Real-world configuration examples
- **[Troubleshooting](troubleshooting.md)**: Fix common configuration issues