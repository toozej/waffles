# Examples

Real-world usage examples for Waffles across different scenarios and project types.

## Table of Contents

- [Getting Started Examples](#getting-started-examples)
- [Development Workflows](#development-workflows)
- [Project Analysis](#project-analysis)
- [Code Review](#code-review)
- [Documentation Generation](#documentation-generation)
- [CI/CD Integration](#cicd-integration)
- [Data Analysis](#data-analysis)
- [Advanced Usage](#advanced-usage)

## Getting Started Examples

### First Setup

```bash
# Complete initial setup
waffles setup --auto-install

# Verify everything works
waffles deps check
waffles query "Hello, what can you tell me about this project?"
```

### Basic Query

```bash
# Simple analysis of current project
waffles query "What is the main purpose of this codebase?"

# Code explanation with specific model
waffles query --model gpt-4 "Explain how the authentication system works"
```

### Configuration Test

```bash
# Test configuration with dry run
waffles query --dry-run --verbose "Test configuration setup"

# Show what files would be analyzed
waffles query --dry-run "Analyze project structure" | grep "detected files"
```

## Development Workflows

### Daily Code Review

```bash
#!/bin/bash
# daily-review.sh - Daily code review script

# Review today's changes
git diff --name-only HEAD~1 | tr '\n' ',' | xargs -I {} \
waffles query --include {} "Review today's changes for potential issues"

# Security check
waffles query --model gpt-4 "Perform security audit on recent changes" \
  --format json --output security-review.json

# Performance analysis  
waffles query --model claude-3-sonnet "Identify performance bottlenecks" \
  --exclude "*_test.go,vendor/*" --format markdown --output perf-analysis.md
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Get staged files
STAGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.(go|py|js|ts)$')

if [ -n "$STAGED_FILES" ]; then
    echo "Running Waffles pre-commit analysis..."
    
    # Create comma-separated list of files
    FILE_LIST=$(echo "$STAGED_FILES" | tr '\n' ',' | sed 's/,$//')
    
    # Analyze staged changes
    waffles query --include "$FILE_LIST" \
        "Review these staged changes for:\n1. Code quality issues\n2. Security vulnerabilities\n3. Performance problems\n4. Best practice violations" \
        --format json --output .waffles-precommit.json
    
    # Check for critical issues (implement your own logic)
    if grep -q "CRITICAL\|SECURITY\|VULNERABILITY" .waffles-precommit.json; then
        echo "❌ Critical issues found. Please review .waffles-precommit.json"
        exit 1
    fi
    
    echo "✅ Pre-commit analysis passed"
    rm -f .waffles-precommit.json
fi
```

### Feature Development

```bash
#!/bin/bash
# feature-analysis.sh

FEATURE_BRANCH=$(git rev-parse --abbrev-ref HEAD)
FEATURE_NAME=${FEATURE_BRANCH#feature/}

echo "Analyzing feature: $FEATURE_NAME"

# Analyze feature changes
git diff main --name-only | tr '\n' ',' | xargs -I {} \
waffles query --include {} \
  "Analyze this feature implementation for:
   1. Completeness and correctness
   2. Integration with existing code
   3. Testing coverage needs
   4. Documentation requirements" \
  --format markdown --output "feature-analysis-$FEATURE_NAME.md"

# Generate test suggestions
waffles query --include "*.go" --exclude "*_test.go" \
  "Suggest comprehensive test cases for this feature" \
  --format markdown --output "test-suggestions-$FEATURE_NAME.md"
```

## Project Analysis

### Architecture Review

```bash
# Comprehensive architecture analysis
waffles query --model gpt-4 \
  "Analyze the overall architecture of this system. Include:
   1. Component relationships and dependencies
   2. Design patterns used
   3. Architectural strengths and weaknesses
   4. Scalability considerations
   5. Recommendations for improvement" \
  --exclude "vendor/*,*_test.go,node_modules/*" \
  --format markdown --output architecture-review.md
```

### Technology Stack Analysis

```bash
# Analyze technology choices
waffles query \
  "Analyze the technology stack used in this project:
   1. List all technologies, frameworks, and libraries
   2. Assess if choices are appropriate for the use case
   3. Identify potential technology debt
   4. Suggest modernization opportunities" \
  --include "*.json,*.yaml,*.toml,go.mod,requirements.txt,package.json" \
  --format json --output tech-stack-analysis.json
```

### Dependencies Audit

```bash
# Security and maintenance audit of dependencies
waffles query --model gpt-4 \
  "Audit project dependencies for:
   1. Security vulnerabilities
   2. Outdated packages
   3. License compliance
   4. Maintenance status
   5. Alternative recommendations" \
  --include "go.mod,go.sum,package.json,requirements.txt,Pipfile" \
  --format markdown --output dependency-audit.md
```

## Code Review

### Pull Request Review

```bash
#!/bin/bash
# pr-review.sh <pr-number>

PR_NUMBER=$1
BRANCH="pr-$PR_NUMBER"

# Fetch PR branch
git fetch origin pull/$PR_NUMBER/head:$BRANCH
git checkout $BRANCH

# Get changed files
CHANGED_FILES=$(git diff main --name-only | tr '\n' ',')

# Comprehensive review
waffles query --model gpt-4 --include "$CHANGED_FILES" \
  "Conduct a thorough code review of this pull request:
   
   ## Code Quality
   - Code clarity and readability
   - Following language idioms and best practices
   - Error handling adequacy
   
   ## Security
   - Potential security vulnerabilities
   - Input validation
   - Authentication/authorization issues
   
   ## Performance
   - Performance implications
   - Resource usage
   - Scalability concerns
   
   ## Testing
   - Test coverage adequacy
   - Edge cases handling
   - Integration test needs
   
   ## Documentation
   - Code documentation quality
   - API documentation needs
   
   Provide specific, actionable feedback with examples." \
  --format markdown --output "pr-$PR_NUMBER-review.md"

# Generate summary
echo "## Pull Request #$PR_NUMBER Review Summary" >> "pr-$PR_NUMBER-review.md"
echo "Generated on: $(date)" >> "pr-$PR_NUMBER-review.md"
```

### Security-Focused Review

```bash
# Security-specific analysis
waffles query --model gpt-4 \
  "Perform a security-focused code review:
   
   1. **Input Validation**: Check for proper input sanitization
   2. **Authentication**: Review auth mechanisms and session handling
   3. **Authorization**: Verify access controls and permissions
   4. **Data Protection**: Check for data exposure and encryption
   5. **Injection Attacks**: Look for SQL, command, or script injection risks
   6. **Cryptography**: Review cryptographic implementations
   7. **Error Handling**: Ensure errors don't leak sensitive information
   8. **Logging**: Check for proper security logging
   
   Rate each area (High/Medium/Low risk) and provide specific recommendations." \
  --exclude "*_test.go,vendor/*" \
  --format markdown --output security-review.md
```

## Documentation Generation

### API Documentation

```bash
# Generate comprehensive API documentation
waffles query --model gpt-4 \
  "Generate comprehensive API documentation for this project:
   
   1. **Overview**: Purpose and scope of the API
   2. **Authentication**: How to authenticate requests
   3. **Endpoints**: List all endpoints with:
      - HTTP methods
      - Parameters and types
      - Request/response examples
      - Error codes and meanings
   4. **Data Models**: Describe all data structures
   5. **Usage Examples**: Common usage patterns
   6. **Rate Limits**: Any limitations or quotas
   7. **SDKs**: Available client libraries" \
  --include "*.go,*.yaml,*.json" --exclude "*_test.go" \
  --format markdown --output api-documentation.md
```

### User Guide Generation

```bash
# Generate user-friendly documentation
waffles query \
  "Create a comprehensive user guide:
   
   ## Getting Started
   - Installation instructions
   - Quick start tutorial
   - Basic configuration
   
   ## Features
   - Complete feature list with examples
   - Use cases and scenarios
   - Best practices
   
   ## Advanced Usage
   - Advanced configuration options
   - Integration examples
   - Troubleshooting guide
   
   ## Reference
   - Command reference
   - Configuration reference
   - FAQ
   
   Make it beginner-friendly with clear examples." \
  --exclude "vendor/*,*_test.go" \
  --format markdown --output user-guide.md
```

### Developer Documentation

```bash
# Generate developer documentation
waffles query --model gpt-4 \
  "Create comprehensive developer documentation:
   
   ## Architecture Overview
   - System architecture and components
   - Data flow and processing
   - Key design decisions and rationale
   
   ## Development Setup
   - Development environment setup
   - Build and test procedures
   - Development workflow
   
   ## Code Organization
   - Directory structure explanation
   - Module responsibilities
   - Interface definitions
   
   ## Contributing Guidelines
   - Code style and standards
   - Testing requirements
   - Pull request process
   
   ## Extension Points
   - How to add new features
   - Plugin architecture (if applicable)
   - Customization options" \
  --exclude "vendor/*,node_modules/*" \
  --format markdown --output developer-guide.md
```

## CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/ai-review.yml
name: AI Code Review

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  ai-review:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
      with:
        fetch-depth: 0
    
    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Install Waffles
      run: go install github.com/toozej/waffles/cmd/waffles@latest
    
    - name: Setup LLM
      run: |
        pip install llm
        echo "${{ secrets.OPENAI_API_KEY }}" | llm keys set openai
    
    - name: Install Dependencies
      run: waffles deps install --non-interactive
    
    - name: Get Changed Files
      id: changed-files
      run: |
        FILES=$(git diff --name-only origin/main...HEAD | grep -E '\.(go|py|js|ts)$' | tr '\n' ',' | sed 's/,$//')
        echo "files=$FILES" >> $GITHUB_OUTPUT
    
    - name: AI Code Review
      if: steps.changed-files.outputs.files
      run: |
        waffles query --include "${{ steps.changed-files.outputs.files }}" \
          --model gpt-4 --format json --output ai-review.json \
          "Review this pull request for code quality, security, and best practices"
    
    - name: Comment PR
      if: steps.changed-files.outputs.files
      uses: actions/github-script@v6
      with:
        script: |
          const fs = require('fs');
          const review = JSON.parse(fs.readFileSync('ai-review.json', 'utf8'));
          
          github.rest.issues.createComment({
            issue_number: context.issue.number,
            owner: context.repo.owner,
            repo: context.repo.repo,
            body: `## 🤖 AI Code Review\n\n${review.final_output}`
          });
```

### Jenkins Pipeline

```groovy
// Jenkinsfile
pipeline {
    agent any
    
    environment {
        WAFFLES_DEFAULT_MODEL = 'gpt-3.5-turbo'
        WAFFLES_DEFAULT_PROVIDER = 'openai'
    }
    
    stages {
        stage('Setup') {
            steps {
                sh 'go install github.com/toozej/waffles/cmd/waffles@latest'
                sh 'pip install llm'
                withCredentials([string(credentialsId: 'openai-api-key', variable: 'OPENAI_API_KEY')]) {
                    sh 'echo $OPENAI_API_KEY | llm keys set openai'
                }
                sh 'waffles deps install --non-interactive'
            }
        }
        
        stage('AI Analysis') {
            steps {
                sh '''
                    waffles query "Analyze this codebase for deployment readiness" \
                        --format json --output deployment-analysis.json
                '''
                
                script {
                    def analysis = readJSON file: 'deployment-analysis.json'
                    if (analysis.final_output.contains('CRITICAL') || analysis.final_output.contains('BLOCKER')) {
                        error("Critical issues found in AI analysis")
                    }
                }
            }
        }
        
        stage('Generate Reports') {
            steps {
                sh '''
                    waffles query "Generate deployment checklist and recommendations" \
                        --format markdown --output deployment-report.md
                '''
                
                publishHTML([
                    allowMissing: false,
                    alwaysLinkToLastBuild: true,
                    keepAll: true,
                    reportDir: '.',
                    reportFiles: 'deployment-report.md',
                    reportName: 'AI Analysis Report'
                ])
            }
        }
    }
}
```

## Data Analysis

### Usage Analytics

```bash
#!/bin/bash
# analytics.sh - Generate usage analytics

echo "# Waffles Usage Analytics Report" > analytics-report.md
echo "Generated on: $(date)" >> analytics-report.md
echo "" >> analytics-report.md

# Export last 30 days of data
waffles export --days 30 --format json --output usage-data.json

# Total executions
TOTAL=$(jq 'length' usage-data.json)
echo "## Summary" >> analytics-report.md
echo "- Total executions (30 days): $TOTAL" >> analytics-report.md

# Success rate
SUCCESS_RATE=$(jq '[.[] | select(.success == true)] | length' usage-data.json)
echo "- Success rate: $(echo "scale=2; $SUCCESS_RATE * 100 / $TOTAL" | bc)%" >> analytics-report.md

# Most used models
echo "" >> analytics-report.md
echo "## Most Used Models" >> analytics-report.md
jq -r '.[] | .model_used' usage-data.json | sort | uniq -c | sort -nr | head -5 | \
    while read count model; do
        echo "- $model: $count uses" >> analytics-report.md
    done

# Language distribution
echo "" >> analytics-report.md
echo "## Language Distribution" >> analytics-report.md
jq -r '.[] | .detected_language' usage-data.json | sort | uniq -c | sort -nr | \
    while read count lang; do
        echo "- $lang: $count projects" >> analytics-report.md
    done

# Average execution time
AVG_TIME=$(jq '[.[] | .execution_time_ms] | add / length' usage-data.json)
echo "- Average execution time: ${AVG_TIME}ms" >> analytics-report.md

echo "Analytics report generated: analytics-report.md"
```

### Performance Analysis

```bash
#!/bin/bash
# performance-analysis.sh

# Export detailed execution data
waffles export --days 7 --include-steps --format json --output perf-data.json

# Analyze step performance
echo "# Performance Analysis" > perf-report.md

echo "## Step Performance (Last 7 Days)" >> perf-report.md
jq -r '.[] | .execution_steps[]? | "\(.tool),\(.duration)"' perf-data.json | \
    awk -F',' '{
        tools[$1] += $2
        counts[$1]++
    }
    END {
        for (tool in tools) {
            printf "- %s: avg %.0fms (%d executions)\n", tool, tools[tool]/counts[tool], counts[tool]
        }
    }' >> perf-report.md

# Find slowest executions
echo "" >> perf-report.md
echo "## Slowest Executions" >> perf-report.md
jq -r '.[] | "\(.execution_time_ms),\(.id),\(.wheresmyprompt_query)"' perf-data.json | \
    sort -nr | head -5 | \
    while IFS=',' read time id query; do
        echo "- ${time}ms: $query (ID: $id)" >> perf-report.md
    done
```

## Advanced Usage

### Custom Workflows

#### Documentation Pipeline

```bash
#!/bin/bash
# docs-pipeline.sh - Automated documentation generation

PROJECT_NAME=$(basename $(pwd))
DOCS_DIR="docs/generated"
mkdir -p $DOCS_DIR

echo "🚀 Starting documentation pipeline for $PROJECT_NAME"

# 1. Architecture documentation
echo "📐 Generating architecture documentation..."
waffles query --model gpt-4 \
    "Create detailed architecture documentation including component diagrams, data flow, and design decisions" \
    --format markdown --output "$DOCS_DIR/architecture.md"

# 2. API documentation
echo "📚 Generating API documentation..."
waffles query --include "*.go,*.yaml" --exclude "*_test.go" \
    "Generate comprehensive API documentation with examples" \
    --format markdown --output "$DOCS_DIR/api.md"

# 3. Deployment guide
echo "🚀 Generating deployment documentation..."
waffles query --include "Dockerfile,*.yaml,*.toml,go.mod" \
    "Create deployment and operations guide" \
    --format markdown --output "$DOCS_DIR/deployment.md"

# 4. Troubleshooting guide
echo "🔧 Generating troubleshooting guide..."
waffles query --include "*.go,README.md" \
    "Generate troubleshooting guide with common issues and solutions" \
    --format markdown --output "$DOCS_DIR/troubleshooting.md"

# 5. Generate table of contents
echo "📖 Generating table of contents..."
cat > "$DOCS_DIR/README.md" << EOF
# $PROJECT_NAME Documentation

Auto-generated documentation using Waffles.

## Contents

- [Architecture](architecture.md)
- [API Documentation](api.md)
- [Deployment Guide](deployment.md)
- [Troubleshooting](troubleshooting.md)

---
*Generated on: $(date)*
EOF

echo "✅ Documentation pipeline complete. See $DOCS_DIR/"
```

### Integration with Development Tools

#### VSCode Tasks

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Waffles: Quick Review",
      "type": "shell",
      "command": "waffles",
      "args": [
        "query",
        "--include", "${file}",
        "Review this file for potential improvements"
      ],
      "group": "build",
      "presentation": {
        "echo": true,
        "reveal": "always",
        "focus": false,
        "panel": "new"
      }
    },
    {
      "label": "Waffles: Security Check",
      "type": "shell",
      "command": "waffles",
      "args": [
        "query",
        "--model", "gpt-4",
        "Perform security review of current project"
      ],
      "group": "build"
    },
    {
      "label": "Waffles: Generate Tests",
      "type": "shell",
      "command": "waffles",
      "args": [
        "query",
        "--include", "${file}",
        "Generate comprehensive unit tests for this file"
      ],
      "group": "test"
    }
  ]
}
```

## Best Practices Summary

### Query Design
- Be specific and detailed in prompts
- Use appropriate models for task complexity
- Include context about what you're trying to achieve

### File Selection
- Use include/exclude patterns effectively
- Consider file size limits for large repositories
- Exclude generated and vendor files

### Performance
- Monitor execution times and optimize patterns
- Use dry-run to test before expensive operations
- Consider using faster models for routine tasks

### Integration
- Set up proper CI/CD integration for team usage
- Create reusable scripts for common workflows
- Document team conventions and standards

### Security
- Be mindful of sensitive data in code
- Use appropriate models and providers for your security requirements
- Regular audit of API key usage and access

This comprehensive set of examples should help users understand how to effectively integrate Waffles into their development workflows across various scenarios and use cases.