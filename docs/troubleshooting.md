# Troubleshooting Guide

Common issues and their solutions when using Waffles.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Dependency Problems](#dependency-problems)
- [Configuration Issues](#configuration-issues)
- [Execution Errors](#execution-errors)
- [Database Problems](#database-problems)
- [Performance Issues](#performance-issues)
- [Integration Issues](#integration-issues)
- [Getting Help](#getting-help)

## Installation Issues

### "waffles: command not found"

**Problem**: The `waffles` command is not found in your PATH.

**Solutions**:

1. **Check if waffles is installed**:
   ```bash
   which waffles
   ls -la $(go env GOPATH)/bin/waffles
   ```

2. **Add Go bin to PATH**:
   ```bash
   echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
   source ~/.bashrc
   ```

3. **For Homebrew installation**:
   ```bash
   brew doctor
   brew --prefix
   echo $PATH
   ```

### "Permission denied" during installation

**Problem**: Insufficient permissions to install waffles.

**Solutions**:

1. **Install to user directory** (Go):
   ```bash
   export GOPATH=$HOME/go
   go install github.com/toozej/waffles/cmd/waffles@latest
   ```

2. **Use sudo for system installation**:
   ```bash
   sudo brew install waffles
   ```

3. **Fix directory permissions**:
   ```bash
   sudo chown -R $(whoami) $(go env GOPATH)
   ```

### "Module not found" error

**Problem**: Go cannot find the module during installation.

**Solutions**:

1. **Update Go**:
   ```bash
   go version  # Should be 1.21+
   ```

2. **Clear module cache**:
   ```bash
   go clean -modcache
   go install github.com/toozej/waffles/cmd/waffles@latest
   ```

3. **Check Go proxy settings**:
   ```bash
   go env GOPROXY
   export GOPROXY=https://proxy.golang.org,direct
   ```

## Dependency Problems

### "wheresmyprompt not found"

**Problem**: The `wheresmyprompt` dependency is not installed or not in PATH.

**Solutions**:

1. **Auto-install dependencies**:
   ```bash
   waffles deps install
   ```

2. **Manual installation**:
   ```bash
   # Using Homebrew
   brew install toozej/tap/wheresmyprompt
   
   # Using Go
   go install github.com/toozej/wheresmyprompt@latest
   ```

3. **Verify installation**:
   ```bash
   wheresmyprompt --version
   which wheresmyprompt
   ```

### "files2prompt not found"

**Problem**: The `files2prompt` dependency is missing.

**Solutions**:

1. **Install using Homebrew**:
   ```bash
   brew install toozej/tap/files2prompt
   ```

2. **Install using Go**:
   ```bash
   go install github.com/toozej/files2prompt@latest
   ```

3. **Check PATH**:
   ```bash
   echo $PATH
   which files2prompt
   ```

### "llm not found"

**Problem**: The LLM CLI is not installed.

**Solutions**:

1. **Install using pipx** (recommended):
   ```bash
   pipx install llm
   ```

2. **Install using pip**:
   ```bash
   pip install llm
   ```

3. **Install using Homebrew**:
   ```bash
   brew install llm
   ```

4. **Verify installation**:
   ```bash
   llm --version
   llm models list
   ```

### "Failed to check dependencies"

**Problem**: Dependency check is failing.

**Solutions**:

1. **Run verbose check**:
   ```bash
   waffles deps check --verbose
   ```

2. **Check system info**:
   ```bash
   waffles deps info
   ```

3. **Reset and reinstall**:
   ```bash
   waffles setup --reset --auto-install
   ```

## Configuration Issues

### "Configuration file not found"

**Problem**: Waffles cannot find its configuration file.

**Solutions**:

1. **Create config directory**:
   ```bash
   mkdir -p ~/.config/waffles
   ```

2. **Run setup wizard**:
   ```bash
   waffles setup
   ```

3. **Create basic config**:
   ```bash
   cat > ~/.config/waffles/.env << EOF
   WAFFLES_DEFAULT_MODEL=claude-3-sonnet
   WAFFLES_DEFAULT_PROVIDER=anthropic
   EOF
   ```

### "Invalid configuration"

**Problem**: Configuration validation is failing.

**Solutions**:

1. **Validate configuration**:
   ```bash
   waffles config validate --verbose
   ```

2. **Show current config**:
   ```bash
   waffles config show
   ```

3. **Reset problematic settings**:
   ```bash
   waffles config reset WAFFLES_DEFAULT_MODEL
   waffles config set WAFFLES_DEFAULT_MODEL claude-3-sonnet
   ```

### "Environment variables not loading"

**Problem**: Environment variables are not being recognized.

**Solutions**:

1. **Check if variables are set**:
   ```bash
   env | grep WAFFLES_
   ```

2. **Source shell profile**:
   ```bash
   source ~/.bashrc  # or ~/.zshrc
   ```

3. **Check config file syntax**:
   ```bash
   cat ~/.config/waffles/.env
   # Ensure no extra spaces around =
   ```

## Execution Errors

### "Pipeline execution failed"

**Problem**: The pipeline fails during execution.

**Solutions**:

1. **Run with verbose output**:
   ```bash
   waffles query --verbose "your prompt"
   ```

2. **Test with dry run**:
   ```bash
   waffles query --dry-run "your prompt"
   ```

3. **Check each tool individually**:
   ```bash
   wheresmyprompt --help
   files2prompt --help
   llm --help
   ```

### "Authentication failed"

**Problem**: LLM provider authentication is failing.

**Solutions**:

1. **Check API keys**:
   ```bash
   llm keys list
   ```

2. **Set API keys**:
   ```bash
   # For OpenAI
   llm keys set openai
   export OPENAI_API_KEY="your-key-here"
   
   # For Anthropic
   llm keys set anthropic  
   export ANTHROPIC_API_KEY="your-key-here"
   ```

3. **Test LLM directly**:
   ```bash
   echo "test" | llm -m gpt-3.5-turbo
   ```

### "Timeout during execution"

**Problem**: Execution times out before completion.

**Solutions**:

1. **Increase timeout**:
   ```bash
   waffles query --timeout 600 "complex prompt"
   ```

2. **Reduce scope**:
   ```bash
   waffles query --exclude "vendor/*,node_modules/*" "focused prompt"
   ```

3. **Check individual tool timeouts**:
   ```bash
   export WAFFLES_TIMEOUT_LLM=300
   export WAFFLES_TIMEOUT_FILES2PROMPT=120
   ```

### "Too many files detected"

**Problem**: Repository has too many files, exceeding limits.

**Solutions**:

1. **Increase limits**:
   ```bash
   export WAFFLES_MAX_FILES=2000
   waffles query "your prompt"
   ```

2. **Use better filtering**:
   ```bash
   waffles query --include "src/**/*.go" --exclude "vendor/*,*_test.go" "focused analysis"
   ```

3. **Check detected files**:
   ```bash
   waffles query --dry-run --verbose "test" 2>&1 | grep "detected files"
   ```

## Database Problems

### "Database connection failed"

**Problem**: Cannot connect to or create the SQLite database.

**Solutions**:

1. **Check database directory permissions**:
   ```bash
   ls -la ~/.config/waffles/
   chmod 755 ~/.config/waffles/
   ```

2. **Create directory manually**:
   ```bash
   mkdir -p ~/.config/waffles
   touch ~/.config/waffles/waffles.db
   ```

3. **Use alternative database path**:
   ```bash
   export WAFFLES_DB_PATH="/tmp/waffles-test.db"
   waffles query "test"
   ```

### "Database schema error"

**Problem**: Database schema is corrupted or incompatible.

**Solutions**:

1. **Reset database**:
   ```bash
   rm ~/.config/waffles/waffles.db
   waffles setup
   ```

2. **Check schema version**:
   ```bash
   sqlite3 ~/.config/waffles/waffles.db "SELECT * FROM schema_version;"
   ```

3. **Export data before reset**:
   ```bash
   waffles export --format json --output backup.json
   rm ~/.config/waffles/waffles.db
   waffles setup
   ```

### "SQLite not found"

**Problem**: SQLite library is not installed.

**Solutions**:

1. **Install SQLite** (Linux):
   ```bash
   # Ubuntu/Debian
   sudo apt install libsqlite3-dev sqlite3
   
   # CentOS/RHEL
   sudo yum install sqlite-devel sqlite
   ```

2. **Install SQLite** (macOS):
   ```bash
   brew install sqlite
   ```

3. **Verify installation**:
   ```bash
   sqlite3 --version
   ```

## Performance Issues

### "Waffles is running slowly"

**Problem**: Execution takes too long.

**Solutions**:

1. **Profile execution**:
   ```bash
   time waffles query --verbose "your prompt"
   ```

2. **Reduce file scope**:
   ```bash
   waffles query --max-files 500 --exclude "vendor/*,node_modules/*" "prompt"
   ```

3. **Use faster model**:
   ```bash
   waffles query --model gpt-3.5-turbo "prompt"
   ```

### "High memory usage"

**Problem**: Waffles is using too much memory.

**Solutions**:

1. **Limit file size**:
   ```bash
   export WAFFLES_MAX_FILE_SIZE=524288  # 512KB
   waffles query "prompt"
   ```

2. **Reduce concurrent operations**:
   ```bash
   waffles query --no-parallel "prompt"
   ```

3. **Monitor memory usage**:
   ```bash
   top -p $(pgrep waffles)
   ```

### "Files2prompt taking too long"

**Problem**: File context extraction is slow.

**Solutions**:

1. **Increase timeout**:
   ```bash
   export WAFFLES_TIMEOUT_FILES2PROMPT=300
   ```

2. **Reduce file count**:
   ```bash
   waffles query --exclude "*.min.js,*.bundle.*,dist/*" "prompt"
   ```

3. **Check file sizes**:
   ```bash
   find . -name "*.go" -size +1M -ls
   ```

## Integration Issues

### "Git integration not working"

**Problem**: Git-related features are failing.

**Solutions**:

1. **Check if in git repository**:
   ```bash
   git status
   ```

2. **Verify git is installed**:
   ```bash
   git --version
   which git
   ```

3. **Check repository permissions**:
   ```bash
   ls -la .git/
   ```

### "1Password integration failing"

**Problem**: 1Password CLI integration is not working.

**Solutions**:

1. **Install 1Password CLI**:
   ```bash
   brew install 1password-cli
   ```

2. **Authenticate with 1Password**:
   ```bash
   op signin
   ```

3. **Check configuration**:
   ```bash
   wheresmyprompt --help | grep -i password
   ```

### "IDE integration not working"

**Problem**: Waffles doesn't work correctly from IDE.

**Solutions**:

1. **Check IDE PATH**:
   ```bash
   echo $PATH
   which waffles
   ```

2. **Use full paths in IDE config**:
   ```json
   {
     "command": "/usr/local/bin/waffles",
     "args": ["query", "analyze this file"]
   }
   ```

3. **Set environment in IDE**:
   ```bash
   WAFFLES_DEFAULT_MODEL=gpt-4
   WAFFLES_VERBOSE=true
   ```

## Platform-Specific Issues

### macOS Issues

#### "App cannot be opened because it is from an unidentified developer"

**Solution**:
```bash
sudo xattr -rd com.apple.quarantine /usr/local/bin/waffles
```

#### "Permission denied on /usr/local"

**Solution**:
```bash
sudo chown -R $(whoami) /usr/local
# Or use Homebrew in user directory
export HOMEBREW_PREFIX=$HOME/.homebrew
```

### Linux Issues

#### "libsqlite3.so not found"

**Solution**:
```bash
# Ubuntu/Debian
sudo apt install libsqlite3-dev

# CentOS/RHEL  
sudo yum install sqlite-devel
```

#### "CGO not enabled"

**Solution**:
```bash
export CGO_ENABLED=1
go install github.com/toozej/waffles/cmd/waffles@latest
```

### Windows Issues

#### "Windows is not directly supported"

**Solution**: Use WSL2:
```powershell
wsl --install -d Ubuntu
wsl
# Follow Linux installation instructions
```

## Debugging Techniques

### Enable Debug Mode

```bash
# Maximum verbosity
export WAFFLES_VERBOSE=true
export DEBUG=true
waffles query --verbose --dry-run "debug test"
```

### Check System State

```bash
# Complete system check
waffles deps check --verbose
waffles config validate --verbose
waffles deps info
```

### Test Individual Components

```bash
# Test wheresmyprompt
wheresmyprompt --help

# Test files2prompt  
files2prompt --help

# Test llm
echo "test" | llm -m gpt-3.5-turbo
```

### Export Logs for Analysis

```bash
# Export recent failures
waffles export --days 1 --failures-only --format json --output debug.json

# View detailed execution history
waffles export --include-steps --format markdown --days 1
```

## Getting Help

### Before Reporting Issues

1. **Update to latest version**:
   ```bash
   waffles version
   go install github.com/toozej/waffles/cmd/waffles@latest
   ```

2. **Collect system information**:
   ```bash
   waffles deps info > system-info.txt
   waffles config show --json > config.json
   ```

3. **Create minimal reproduction**:
   ```bash
   # Test with simple query
   waffles query --dry-run --verbose "minimal test case"
   ```

### Report Issues

Include this information when reporting issues:

- **Waffles version**: `waffles version`
- **Operating system**: `uname -a`
- **Go version**: `go version`
- **Dependency status**: `waffles deps check`
- **Configuration**: `waffles config show` (redact sensitive info)
- **Error output**: Full error messages and stack traces
- **Steps to reproduce**: Minimal commands to reproduce the issue

### Community Resources

- **GitHub Issues**: [github.com/toozej/waffles/issues](https://github.com/toozej/waffles/issues)
- **GitHub Discussions**: [github.com/toozej/waffles/discussions](https://github.com/toozej/waffles/discussions)
- **Documentation**: [docs/](https://github.com/toozej/waffles/tree/main/docs)

### Self-Help Resources

1. **[Installation Guide](installation.md)**: Detailed installation instructions
2. **[Configuration Guide](configuration.md)**: Configuration troubleshooting
3. **[Usage Guide](usage.md)**: Best practices and common patterns
4. **[Commands Reference](commands.md)**: Complete command documentation
5. **[Examples](examples.md)**: Working examples for reference

## Next Steps

After resolving issues:

1. **Verify fix**: Test with your original use case
2. **Update documentation**: Consider contributing fixes back
3. **Share solution**: Help others with similar issues
4. **Monitor**: Watch for recurring problems