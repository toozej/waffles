# Docker Guide

Complete guide to using Waffles with Docker, including pre-built images with all dependencies included.

## Table of Contents

- [Overview](#overview)
- [Available Images](#available-images)
- [Quick Start](#quick-start)
- [Image Details](#image-details)
- [Usage Examples](#usage-examples)
- [Configuration](#configuration)
- [Building Images](#building-images)
- [Troubleshooting](#troubleshooting)

## Overview

Waffles provides two Docker images that include all necessary dependencies for running the complete LLM pipeline:

- **Standard Image (`Dockerfile`)**: Full Python runtime with all tools
- **Distroless Image (`Dockerfile.distroless`)**: Minimal security-hardened image

Both images include:
- **waffles**: The main orchestration tool
- **wheresmyprompt**: Context extraction from various sources
- **files2prompt**: File content aggregation  
- **llm**: LLM CLI for model interactions

## Available Images

### Docker Hub Images

```bash
# Pull the latest standard image
docker pull toozej/waffles:latest

# Pull the latest distroless image  
docker pull toozej/waffles:distroless

# Pull specific version
docker pull toozej/waffles:v1.0.0
docker pull toozej/waffles:v1.0.0-distroless
```

### Image Variants

| Image | Base | Size | Use Case |
|-------|------|------|----------|
| `toozej/waffles:latest` | `python:3.11-slim` | ~200MB | Development, CI/CD |
| `toozej/waffles:distroless` | `gcr.io/distroless/python3` | ~150MB | Production, Security-focused |

## Quick Start

### Basic Usage

```bash
# Run in current directory
docker run -it --rm \
  -v $(pwd):/workspace \
  -e OPENAI_API_KEY="your-api-key" \
  toozej/waffles:latest query "What does this codebase do?"
```

### With Configuration File

```bash
# Create local config
cat > .env << EOF
WAFFLES_DEFAULT_MODEL=claude-3-sonnet
WAFFLES_DEFAULT_PROVIDER=anthropic
ANTHROPIC_API_KEY=your-api-key-here
EOF

# Run with config
docker run -it --rm \
  -v $(pwd):/workspace \
  --env-file .env \
  toozej/waffles:latest query "Analyze this project structure"
```

### Interactive Setup

```bash
# Run setup wizard
docker run -it --rm \
  -v $(pwd):/workspace \
  -v ~/.config/waffles:/root/.config/waffles \
  toozej/waffles:latest setup --auto-install
```

## Image Details

### Included Dependencies

Both images include the latest versions of:

#### Go Tools (Static Binaries)
- **wheresmyprompt**: `github.com/toozej/wheresmyprompt@latest`
  - Location: `/usr/local/bin/wheresmyprompt`
  - Purpose: Extract prompts from various sources
  
- **files2prompt**: `github.com/toozej/files2prompt@latest`
  - Location: `/usr/local/bin/files2prompt`  
  - Purpose: Aggregate file contents for LLM context

#### Python Tools
- **llm**: Latest version via pip/uv
  - Location: `/usr/local/bin/llm`
  - Purpose: LLM model interactions
  - Supports: OpenAI, Anthropic, local models

### Runtime Environment

#### Standard Image (`Dockerfile`)
```dockerfile
# Base: python:3.11-slim-bookworm
# Runtime: Full Python 3.11 environment
# Package Manager: uv (faster than pip)
# Dependencies: ca-certificates
# Working Directory: /workspace
```

#### Distroless Image (`Dockerfile.distroless`)  
```dockerfile
# Base: gcr.io/distroless/python3-debian12
# Runtime: Minimal Python environment
# Security: No shell, minimal attack surface
# Dependencies: Only essential runtime libraries
# Working Directory: /workspace
```

## Usage Examples

### Development Workflow

```bash
#!/bin/bash
# dev-with-docker.sh

# Set up Docker alias for easier usage
alias waffles='docker run -it --rm \
  -v $(pwd):/workspace \
  -v ~/.config/waffles:/root/.config/waffles \
  --env-file .env \
  toozej/waffles:latest'

# Now use as normal
waffles query "Review this code for potential issues"
waffles deps check
waffles export --format json --days 7
```

### CI/CD Integration

#### GitHub Actions
```yaml
name: AI Code Review
on: [pull_request]

jobs:
  ai-review:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: AI Code Review
      run: |
        docker run --rm \
          -v ${{ github.workspace }}:/workspace \
          -e OPENAI_API_KEY="${{ secrets.OPENAI_API_KEY }}" \
          -e WAFFLES_DEFAULT_MODEL="gpt-4" \
          toozej/waffles:distroless \
          query "Review this pull request for code quality and security issues" \
          --format json --output /workspace/ai-review.json
    
    - name: Upload Results
      uses: actions/upload-artifact@v3
      with:
        name: ai-review
        path: ai-review.json
```

#### Jenkins Pipeline
```groovy
pipeline {
    agent any
    stages {
        stage('AI Analysis') {
            steps {
                script {
                    docker.image('toozej/waffles:distroless').inside('-v $WORKSPACE:/workspace') {
                        sh '''
                            waffles query "Analyze codebase for deployment readiness" \
                                --format markdown --output deployment-analysis.md
                        '''
                    }
                }
            }
        }
    }
}
```

### Batch Processing

```bash
#!/bin/bash
# batch-analysis.sh

DOCKER_CMD="docker run --rm \
  -v $(pwd):/workspace \
  --env-file .env \
  toozej/waffles:latest"

# Multiple analyses
queries=(
    "What is the overall architecture?"
    "Identify security vulnerabilities"
    "Suggest performance optimizations"
    "Review code quality and best practices"
)

for i, query in "${queries[@]}"; do
    echo "Running analysis $((i+1)): $query"
    $DOCKER_CMD query "$query" \
        --format json \
        --output "analysis-$((i+1)).json"
done

# Combine results
$DOCKER_CMD export --format markdown --output combined-report.md
```

## Configuration

### Environment Variables

```bash
# Core configuration
WAFFLES_DEFAULT_MODEL=claude-3-sonnet
WAFFLES_DEFAULT_PROVIDER=anthropic
WAFFLES_VERBOSE=true
WAFFLES_AUTO_INSTALL=false  # Dependencies pre-installed in image

# API Keys
OPENAI_API_KEY=your-openai-key
ANTHROPIC_API_KEY=your-anthropic-key

# File processing
WAFFLES_MAX_FILES=1000
WAFFLES_MAX_FILE_SIZE=1048576

# Tool arguments
WAFFLES_WHERESMYPROMPT_ARGS="--exclude .git,vendor,node_modules"
WAFFLES_FILES2PROMPT_ARGS="--max-tokens 4000"
WAFFLES_LLM_ARGS="--temperature 0.1"
```

### Volume Mounts

```bash
# Essential mounts
-v $(pwd):/workspace                    # Project files
-v ~/.config/waffles:/root/.config/waffles  # Persistent config

# Optional mounts
-v ~/.llm:/root/.llm                    # LLM CLI config
-v /var/run/docker.sock:/var/run/docker.sock  # Docker-in-Docker (if needed)
```

### Network Configuration

```bash
# For API access
docker run --rm \
  --network host \
  -v $(pwd):/workspace \
  toozej/waffles:latest query "Your prompt"

# With custom network
docker network create waffles-net
docker run --rm \
  --network waffles-net \
  -v $(pwd):/workspace \
  toozej/waffles:latest query "Your prompt"
```

## Building Images

### Build Standard Image

```bash
# Build from source
git clone https://github.com/toozej/waffles.git
cd waffles

# Build with default settings
docker build -t waffles:local .

# Build with custom LDFLAGS
docker build --build-arg LDFLAGS="-X main.version=custom" -t waffles:custom .

# Multi-platform build
docker buildx build --platform linux/amd64,linux/arm64 -t waffles:multiarch .
```

### Build Distroless Image

```bash
# Build distroless variant
docker build -f Dockerfile.distroless -t waffles:distroless .

# Build and push
docker build -f Dockerfile.distroless -t myregistry/waffles:distroless .
docker push myregistry/waffles:distroless
```

### Build Arguments

| Argument | Description | Default |
|----------|-------------|---------|
| `LDFLAGS` | Go linker flags for versioning | _(empty)_ |

### Multi-stage Build Benefits

1. **Smaller Images**: Only runtime dependencies in final image
2. **Security**: No build tools in production image  
3. **Caching**: Efficient layer caching for faster rebuilds
4. **Separation**: Clear separation between build and runtime

## Troubleshooting

### Common Issues

#### "Permission denied" errors
```bash
# Fix with proper user mapping
docker run --rm \
  -v $(pwd):/workspace \
  --user $(id -u):$(id -g) \
  toozej/waffles:latest query "Your prompt"
```

#### "API key not found"
```bash
# Verify environment variables are passed
docker run --rm \
  -v $(pwd):/workspace \
  -e OPENAI_API_KEY="$OPENAI_API_KEY" \
  toozej/waffles:latest config show
```

#### "wheresmyprompt not found"
```bash
# Verify tool availability
docker run --rm toozej/waffles:latest wheresmyprompt --version
docker run --rm toozej/waffles:latest files2prompt --version
docker run --rm toozej/waffles:latest llm --version
```

#### Container exits immediately
```bash
# Check with interactive mode
docker run -it --rm \
  -v $(pwd):/workspace \
  toozej/waffles:latest /bin/sh

# Or with debug output
docker run --rm \
  -v $(pwd):/workspace \
  -e WAFFLES_VERBOSE=true \
  toozej/waffles:latest query "debug test" --dry-run
```

### Debugging

#### Check Dependencies
```bash
docker run --rm toozej/waffles:latest deps check --verbose
```

#### Inspect Image
```bash
# List installed tools
docker run --rm toozej/waffles:latest which waffles wheresmyprompt files2prompt llm

# Check Python environment (standard image)
docker run --rm toozej/waffles:latest python3 -c "import sys; print(sys.path)"

# Check file system
docker run --rm toozej/waffles:latest find /usr/local/bin -name "*prompt*" -o -name "llm"
```

#### Resource Usage
```bash
# Monitor resource usage
docker run --rm \
  --memory 512m \
  --cpus 1.0 \
  -v $(pwd):/workspace \
  toozej/waffles:latest query "Resource-constrained analysis"
```

### Performance Optimization

#### Image Size Reduction
```bash
# Use distroless for production
docker run --rm toozej/waffles:distroless query "Minimal footprint analysis"

# Multi-stage builds already optimize layer usage
docker history toozej/waffles:latest
```

#### Caching Strategy
```bash
# Pre-pull images for faster startup
docker pull toozej/waffles:latest

# Use local registry for team sharing
docker tag toozej/waffles:latest localhost:5000/waffles:latest
docker push localhost:5000/waffles:latest
```

## Integration Examples

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  waffles:
    image: toozej/waffles:latest
    volumes:
      - .:/workspace
      - ~/.config/waffles:/root/.config/waffles
    environment:
      - WAFFLES_DEFAULT_MODEL=claude-3-sonnet
      - WAFFLES_DEFAULT_PROVIDER=anthropic
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    working_dir: /workspace
    command: query "Analyze this project"

  # Analysis with specific model
  gpt4-analysis:
    image: toozej/waffles:distroless
    volumes:
      - .:/workspace
    environment:
      - WAFFLES_DEFAULT_MODEL=gpt-4
      - WAFFLES_DEFAULT_PROVIDER=openai
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    command: query "Detailed architectural analysis"
```

### Kubernetes Deployment

```yaml
# waffles-job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: waffles-analysis
spec:
  template:
    spec:
      containers:
      - name: waffles
        image: toozej/waffles:distroless
        command: ["waffles", "query", "Analyze project for deployment readiness"]
        env:
        - name: WAFFLES_DEFAULT_MODEL
          value: "gpt-4"
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: waffles-secrets
              key: openai-api-key
        volumeMounts:
        - name: project-source
          mountPath: /workspace
        - name: results
          mountPath: /results
      volumes:
      - name: project-source
        configMap:
          name: project-files
      - name: results
        emptyDir: {}
      restartPolicy: Never
```

## Next Steps

- **[Installation Guide](installation.md)**: Alternative installation methods
- **[Usage Guide](usage.md)**: Detailed usage patterns  
- **[Configuration](configuration.md)**: Advanced configuration options
- **[Examples](examples.md)**: Real-world usage examples
- **[Troubleshooting](troubleshooting.md)**: Common issues and solutions

## Contributing

To contribute improvements to the Docker images:

1. Fork the repository
2. Modify the Dockerfiles
3. Test builds locally
4. Submit a pull request with clear description of changes

For image-specific issues, please include:
- Docker version: `docker --version`
- Image tag and digest
- Full command used
- Error output and logs