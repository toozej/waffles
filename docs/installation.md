# Installation Guide

This guide covers all methods for installing Waffles and its dependencies.

## Table of Contents

- [System Requirements](#system-requirements)
- [Installation Methods](#installation-methods)
- [Dependency Installation](#dependency-installation)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)

## System Requirements

### Operating Systems
- **macOS**: 10.15+ (Catalina and later)
- **Linux**: Ubuntu 18.04+, Debian 10+, CentOS 8+, or equivalent
- **Windows**: Windows 10+ (with WSL2 recommended)

### Prerequisites
- **Go**: 1.21+ (if building from source)
- **Python**: 3.8+ (for LLM CLI)
- **Node.js**: 16+ (optional, for some integrations)

## Installation Methods

### Method 1: Homebrew (macOS/Linux - Recommended)

```bash
# Add the tap
brew tap toozej/tap

# Install waffles
brew install waffles

# Verify installation
waffles version
```

### Method 2: Go Install

```bash
# Install latest release
go install github.com/toozej/waffles/cmd/waffles@latest

# Install specific version
go install github.com/toozej/waffles/cmd/waffles@v1.0.0

# Verify installation
waffles version
```

### Method 3: Pre-built Binaries

Download pre-built binaries from the [releases page](https://github.com/toozej/waffles/releases):

```bash
# Example for Linux amd64
curl -L https://github.com/toozej/waffles/releases/latest/download/waffles-linux-amd64.tar.gz | tar xz
sudo mv waffles /usr/local/bin/
```

### Method 4: Build from Source

```bash
# Clone the repository
git clone https://github.com/toozej/waffles.git
cd waffles

# Build and install
make install

# Or just build
make build
./bin/waffles version
```

## Dependency Installation

Waffles requires three core dependencies. You can install them manually or let Waffles handle it automatically.

### Automatic Installation (Recommended)

```bash
# Interactive setup with auto-installation
waffles setup --auto-install

# Or install dependencies only
waffles deps install
```

### Manual Installation

#### wheresmyprompt

```bash
# Using Homebrew
brew install toozej/tap/wheresmyprompt

# Using Go
go install github.com/toozej/wheresmyprompt@latest
```

#### files2prompt

```bash
# Using Homebrew  
brew install toozej/tap/files2prompt

# Using Go
go install github.com/toozej/files2prompt@latest
```

#### llm CLI

```bash
# Using Homebrew
brew install llm

# Using pipx (recommended for Python)
pipx install llm

# Using pip
pip install llm
```

### Optional Dependencies

#### 1Password CLI (for secure credential management)

```bash
# Using Homebrew
brew install 1password-cli

# Or download from https://1password.com/downloads/command-line/
```

## Platform-Specific Instructions

### macOS

1. **Homebrew Installation** (easiest):
   ```bash
   /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
   brew tap toozej/tap
   brew install waffles
   ```

2. **Setup Dependencies**:
   ```bash
   waffles setup --auto-install
   ```

### Linux (Ubuntu/Debian)

1. **Install Prerequisites**:
   ```bash
   sudo apt update
   sudo apt install -y curl git build-essential python3-pip
   ```

2. **Install Go** (if needed):
   ```bash
   curl -L https://go.dev/dl/go1.21.0.linux-amd64.tar.gz | sudo tar -xz -C /usr/local
   echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
   source ~/.bashrc
   ```

3. **Install Waffles**:
   ```bash
   go install github.com/toozej/waffles/cmd/waffles@latest
   ```

4. **Setup Dependencies**:
   ```bash
   waffles setup --auto-install
   ```

### Windows (WSL2)

1. **Install WSL2** (if not already installed):
   ```powershell
   wsl --install -d Ubuntu
   ```

2. **Inside WSL2**, follow the Linux instructions above.

### Method 5: Docker (Complete Environment - Recommended for Teams)

Docker images include all dependencies pre-installed for easy deployment:

```bash
# Pull the complete image with all dependencies
docker pull toozej/waffles:latest

# Quick start - run in current directory
docker run -it --rm \
  -v $(pwd):/workspace \
  -e OPENAI_API_KEY="your-api-key" \
  toozej/waffles:latest query "What does this codebase do?"

# Or use the security-hardened distroless image
docker pull toozej/waffles:distroless
docker run -it --rm \
  -v $(pwd):/workspace \
  --env-file .env \
  toozej/waffles:distroless query "Analyze this project"

# Create alias for easier usage
echo 'alias waffles="docker run -it --rm -v \$(pwd):/workspace --env-file .env toozej/waffles:latest"' >> ~/.bashrc
```

**Docker Images Include:**
- ✅ **waffles**: Main orchestration tool
- ✅ **wheresmyprompt**: Latest version pre-installed
- ✅ **files2prompt**: Latest version pre-installed
- ✅ **llm**: Latest version pre-installed
- ✅ **Python runtime**: Complete environment ready to use

**Available Images:**
- `toozej/waffles:latest` - Full Python runtime (~200MB)
- `toozej/waffles:distroless` - Minimal security-hardened (~150MB)

📖 **See [Docker Guide](docker.md) for complete Docker usage documentation**

## Verification

After installation, verify everything works:

```bash
# Check waffles version
waffles version

# Check all dependencies
waffles deps check

# Run setup wizard
waffles setup

# Test with a simple query
waffles query "test query"
```

## Configuration

### Environment Variables

Set these in your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
# Core configuration
export WAFFLES_DEFAULT_MODEL="claude-3-sonnet"
export WAFFLES_DEFAULT_PROVIDER="anthropic"

# Database location
export WAFFLES_DB_PATH="$HOME/.config/waffles/waffles.db"

# Optional: LLM configuration
export LLM_USER_PATH="$HOME/.config/io.datasette.llm"
```

### Config File

Create `~/.config/waffles/.env`:

```env
WAFFLES_DEFAULT_MODEL=claude-3-sonnet
WAFFLES_DEFAULT_PROVIDER=anthropic
WAFFLES_DB_PATH=/home/user/.config/waffles/waffles.db
WAFFLES_AUTO_INSTALL=true
WAFFLES_VERBOSE=false
```

## Troubleshooting

### Common Issues

#### "waffles: command not found"

**Solution**: Add Go bin to PATH:
```bash
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

#### "Permission denied" when installing

**Solution**: Use sudo or install to user directory:
```bash
# For Go install, ensure GOPATH/bin is writable
mkdir -p $(go env GOPATH)/bin

# Or use --user flag for pip installations
pip install --user llm
```

#### "Dependencies not found"

**Solution**: Run the dependency installer:
```bash
waffles deps install --verbose
```

#### "Database connection failed"

**Solution**: Ensure directory exists and has write permissions:
```bash
mkdir -p ~/.config/waffles
chmod 755 ~/.config/waffles
```

### Platform-Specific Issues

#### macOS: "cannot be opened because it is from an unidentified developer"

**Solution**:
```bash
# Remove quarantine attribute
sudo xattr -rd com.apple.quarantine /usr/local/bin/waffles
```

#### Linux: "libsqlite3 not found"

**Solution**:
```bash
# Ubuntu/Debian
sudo apt install libsqlite3-dev

# CentOS/RHEL
sudo yum install sqlite-devel
```

### Getting Help

If you encounter issues:

1. **Check verbose output**:
   ```bash
   waffles --verbose query "test"
   ```

2. **Check logs**:
   ```bash
   # View recent executions
   waffles export --format json --days 1
   ```

3. **Report issues**: [GitHub Issues](https://github.com/toozej/waffles/issues)

## Next Steps

After successful installation:

1. **[Docker Guide](docker.md)**: Complete Docker usage and deployment guide
2. **[Configuration](configuration.md)**: Customize Waffles for your needs
3. **[Usage Guide](usage.md)**: Learn how to use Waffles effectively
4. **[Examples](examples.md)**: See real-world usage examples