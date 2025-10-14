# Testing Guide

This document covers testing strategies, best practices, and guidelines for the Waffles project.

## Table of Contents

- [Testing Philosophy](#testing-philosophy)
- [Test Types](#test-types)
- [Test Organization](#test-organization)
- [Writing Tests](#writing-tests)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [Testing Best Practices](#testing-best-practices)
- [Continuous Integration](#continuous-integration)
- [Performance Testing](#performance-testing)
- [Troubleshooting Tests](#troubleshooting-tests)

## Testing Philosophy

Waffles follows a comprehensive testing strategy that emphasizes:

1. **Reliability** - Tests should be deterministic and stable
2. **Coverage** - Critical paths and edge cases must be tested
3. **Performance** - Tests should run quickly and efficiently
4. **Maintainability** - Tests should be easy to understand and modify
5. **Real-world scenarios** - Tests should reflect actual usage patterns

### Testing Pyramid

```
                /\
               /  \
              /E2E \
             /______\
            /        \
           /Integration\
          /__________\
         /            \
        /     Unit     \
       /________________\
```

- **Unit Tests** (70%) - Fast, focused tests for individual functions
- **Integration Tests** (20%) - Tests for component interactions
- **End-to-End Tests** (10%) - Full system tests with real dependencies

## Test Types

### Unit Tests

Test individual functions, methods, and components in isolation.

```go
// Example: pkg/config/config_test.go
func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  *Config
        wantErr bool
        errType error
    }{
        {
            name: "valid config",
            config: &Config{
                DefaultModel:    "gpt-4",
                DefaultProvider: "openai",
                Timeout:         300,
            },
            wantErr: false,
        },
        {
            name:    "missing model",
            config:  &Config{DefaultProvider: "openai"},
            wantErr: true,
            errType: ErrMissingModel,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if tt.wantErr {
                if err == nil {
                    t.Error("expected error but got none")
                }
                if tt.errType != nil && !errors.Is(err, tt.errType) {
                    t.Errorf("expected error type %T, got %T", tt.errType, err)
                }
            } else if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}
```

### Integration Tests

Test interactions between components, typically involving databases, file systems, or external processes.

```go
// Example: pkg/pipeline/integration_test.go
func TestPipelineWithDatabase(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Set up test environment
    tempDir := t.TempDir()
    dbPath := filepath.Join(tempDir, "test.db")
    
    cfg := &config.Config{
        LogDBPath:       dbPath,
        DefaultModel:    "test-model",
        DefaultProvider: "test",
        Timeout:         30,
    }

    // Create pipeline with real database
    db, err := logging.NewDatabase(dbPath)
    if err != nil {
        t.Fatalf("Failed to create database: %v", err)
    }
    defer db.Close()

    pipeline := pipeline.New(cfg, db)

    // Execute pipeline
    ctx := context.Background()
    result, err := pipeline.Execute(ctx, "test query")

    // Verify results
    if err != nil {
        t.Fatalf("Pipeline execution failed: %v", err)
    }

    if result.Success != true {
        t.Error("Expected successful execution")
    }

    // Verify database logging
    logs, err := db.GetExecutions()
    if err != nil {
        t.Fatalf("Failed to retrieve logs: %v", err)
    }

    if len(logs) != 1 {
        t.Errorf("Expected 1 log entry, got %d", len(logs))
    }
}
```

### End-to-End Tests

Test complete workflows with real external dependencies.

```go
// Example: integration_test.go
func TestCompleteWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    // Ensure required tools are available
    tools := []string{"wheresmyprompt", "files2prompt", "llm"}
    for _, tool := range tools {
        if _, err := exec.LookPath(tool); err != nil {
            t.Skipf("Skipping E2E test: %s not available", tool)
        }
    }

    // Set up temporary directory with test files
    testDir := setupTestRepository(t)
    defer os.RemoveAll(testDir)

    // Run waffles command
    cmd := exec.Command("waffles", "query", "What files are in this repository?")
    cmd.Dir = testDir
    cmd.Env = append(os.Environ(),
        "WAFFLES_DEFAULT_MODEL=gpt-3.5-turbo",
        "WAFFLES_DEFAULT_PROVIDER=openai",
    )

    output, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("Command failed: %v\nOutput: %s", err, output)
    }

    // Verify output contains expected information
    if !strings.Contains(string(output), "Repository analysis") {
        t.Error("Expected output to contain repository analysis")
    }
}
```

### Benchmark Tests

Test performance characteristics of critical functions.

```go
// Example: pkg/repo/detector_bench_test.go
func BenchmarkLanguageDetection(b *testing.B) {
    detector := NewDetector()
    testFiles := []string{
        "main.go",
        "app.js",
        "style.css",
        "README.md",
        "Dockerfile",
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        for _, file := range testFiles {
            detector.DetectLanguage(file)
        }
    }
}

func BenchmarkLargeRepoAnalysis(b *testing.B) {
    // Create test repository with many files
    tempDir := createLargeTestRepo(b, 1000)
    defer os.RemoveAll(tempDir)

    analyzer := repo.NewAnalyzer()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := analyzer.Analyze(tempDir)
        if err != nil {
            b.Fatalf("Analysis failed: %v", err)
        }
    }
}
```

## Test Organization

### Directory Structure

```
pkg/
├── config/
│   ├── config.go
│   └── config_test.go       # Unit tests
├── pipeline/
│   ├── pipeline.go
│   ├── pipeline_test.go     # Unit tests
│   └── integration_test.go  # Integration tests
└── repo/
    ├── detector.go
    ├── detector_test.go     # Unit tests
    └── detector_bench_test.go # Benchmark tests

testdata/                    # Test fixtures
├── repositories/
│   ├── go-project/
│   ├── js-project/
│   └── python-project/
└── configs/
    ├── valid.env
    └── invalid.env

integration_test.go          # End-to-end tests
```

### Test File Naming

- **Unit tests**: `*_test.go` in same package
- **Integration tests**: `integration_test.go` or `*_integration_test.go`
- **Benchmark tests**: `*_bench_test.go`
- **Example tests**: `*_example_test.go`

### Test Data

Store test fixtures in `testdata/` directories:

```go
func TestConfigLoading(t *testing.T) {
    configPath := filepath.Join("testdata", "configs", "valid.env")
    cfg, err := LoadConfig(configPath)
    if err != nil {
        t.Fatalf("Failed to load config: %v", err)
    }
    
    if cfg.DefaultModel != "expected-model" {
        t.Errorf("Expected model 'expected-model', got %s", cfg.DefaultModel)
    }
}
```

## Writing Tests

### Test Structure

Follow the Arrange-Act-Assert (AAA) pattern:

```go
func TestFunctionName(t *testing.T) {
    // Arrange - Set up test data and dependencies
    cfg := &Config{
        DefaultModel: "test-model",
        Timeout:      30,
    }
    
    // Act - Execute the function under test
    result, err := ProcessConfig(cfg)
    
    // Assert - Verify the results
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    
    if result.Model != "test-model" {
        t.Errorf("Expected model 'test-model', got %s", result.Model)
    }
}
```

### Table-Driven Tests

Use table-driven tests for testing multiple scenarios:

```go
func TestValidateInput(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    bool
        wantErr bool
    }{
        {"valid input", "valid", true, false},
        {"empty input", "", false, true},
        {"invalid format", "invalid-format", false, true},
        {"special characters", "test@#$", false, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ValidateInput(tt.input)
            
            if tt.wantErr && err == nil {
                t.Error("expected error but got none")
                return
            }
            
            if !tt.wantErr && err != nil {
                t.Errorf("unexpected error: %v", err)
                return
            }
            
            if got != tt.want {
                t.Errorf("ValidateInput() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Test Helpers

Create reusable test helpers:

```go
// testhelpers/setup.go
func SetupTestConfig(t *testing.T) *config.Config {
    t.Helper()
    return &config.Config{
        DefaultModel:    "test-model",
        DefaultProvider: "test",
        LogDBPath:       ":memory:",
        Timeout:         30,
    }
}

func SetupTestDB(t *testing.T) *logging.Database {
    t.Helper()
    db, err := logging.NewDatabase(":memory:")
    if err != nil {
        t.Fatalf("Failed to create test database: %v", err)
    }
    
    t.Cleanup(func() { 
        if err := db.Close(); err != nil {
            t.Errorf("Failed to close test database: %v", err)
        }
    })
    
    return db
}

func SetupTestRepo(t *testing.T, files map[string]string) string {
    t.Helper()
    tempDir := t.TempDir()
    
    for file, content := range files {
        path := filepath.Join(tempDir, file)
        if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
            t.Fatalf("Failed to create directory: %v", err)
        }
        if err := os.WriteFile(path, []byte(content), 0644); err != nil {
            t.Fatalf("Failed to write file %s: %v", file, err)
        }
    }
    
    return tempDir
}
```

### Mocking

Use interfaces and dependency injection for mockable dependencies:

```go
// Define interface
type HTTPClient interface {
    Get(url string) (*http.Response, error)
}

// Production implementation
type RealHTTPClient struct {
    client *http.Client
}

func (c *RealHTTPClient) Get(url string) (*http.Response, error) {
    return c.client.Get(url)
}

// Test mock
type MockHTTPClient struct {
    responses map[string]*http.Response
    errors    map[string]error
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
    if err, exists := m.errors[url]; exists {
        return nil, err
    }
    if resp, exists := m.responses[url]; exists {
        return resp, nil
    }
    return nil, fmt.Errorf("no mock response for URL: %s", url)
}

// Test using mock
func TestAPICall(t *testing.T) {
    mock := &MockHTTPClient{
        responses: map[string]*http.Response{
            "https://api.example.com/test": {
                StatusCode: 200,
                Body: io.NopCloser(strings.NewReader(`{"result": "success"}`)),
            },
        },
    }

    service := NewService(mock)
    result, err := service.MakeAPICall("https://api.example.com/test")
    
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    
    if result != "success" {
        t.Errorf("Expected 'success', got %s", result)
    }
}
```

## Running Tests

### Basic Test Commands

```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run specific package
go test -v ./pkg/config

# Run specific test
go test -v ./pkg/config -run TestConfigValidation

# Run tests matching pattern
go test -v ./... -run "Test.*Validation"
```

### Test Modes

```bash
# Short mode (skips integration tests)
go test -short ./...

# Race detection
go test -race ./...

# With coverage
go test -cover ./...

# Detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Benchmark tests
go test -bench=. ./pkg/repo

# Memory benchmarks
go test -bench=. -benchmem ./pkg/repo

# CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./pkg/repo
```

### Makefile Targets

```makefile
# Run all tests
test:
	go test -short ./...

# Run all tests including integration
test-all:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run integration tests only
test-integration:
	go test -run Integration ./...

# Run benchmarks
test-benchmark:
	go test -bench=. -benchmem ./...

# Test with race detection
test-race:
	go test -race ./...

# Clean test cache
test-clean:
	go clean -testcache
```

## Test Coverage

### Coverage Requirements

- **Overall coverage**: Minimum 80%
- **Critical paths**: Minimum 90%
- **New features**: Must include comprehensive tests
- **Bug fixes**: Must include regression tests

### Coverage Analysis

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage by package
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Show uncovered lines
go tool cover -func=coverage.out | grep -v "100.0%"
```

### Coverage in CI

```yaml
# .github/workflows/test.yml
- name: Test with coverage
  run: |
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

- name: Check coverage threshold
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total: | awk '{print $3}' | sed 's/%//')
    echo "Coverage: ${COVERAGE}%"
    if (( $(echo "$COVERAGE < 80" | bc -l) )); then
      echo "Coverage below 80%"
      exit 1
    fi
```

## Testing Best Practices

### Naming Tests

- Use descriptive test names that explain what is being tested
- Use `Test[Function/Feature]` pattern
- Use subtests for multiple scenarios

```go
func TestConfigValidation(t *testing.T) {
    t.Run("ValidConfig", func(t *testing.T) {
        // Test valid configuration
    })
    
    t.Run("MissingModel", func(t *testing.T) {
        // Test missing model error
    })
    
    t.Run("InvalidTimeout", func(t *testing.T) {
        // Test invalid timeout error
    })
}
```

### Test Independence

- Tests should not depend on each other
- Each test should set up its own data
- Use `t.TempDir()` for temporary files
- Clean up resources with `t.Cleanup()`

```go
func TestFileProcessing(t *testing.T) {
    // Create temporary directory
    tempDir := t.TempDir()
    
    // Set up test files
    testFile := filepath.Join(tempDir, "test.txt")
    if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
        t.Fatalf("Failed to create test file: %v", err)
    }
    
    // Test file processing
    result, err := ProcessFile(testFile)
    // ... assertions
}
```

### Error Testing

- Test both success and error cases
- Use `errors.Is()` for error comparison
- Test error messages when appropriate

```go
func TestProcessFile(t *testing.T) {
    t.Run("Success", func(t *testing.T) {
        result, err := ProcessFile("valid-file.txt")
        if err != nil {
            t.Fatalf("Unexpected error: %v", err)
        }
        // ... verify result
    })
    
    t.Run("FileNotFound", func(t *testing.T) {
        _, err := ProcessFile("nonexistent.txt")
        if err == nil {
            t.Fatal("Expected error for nonexistent file")
        }
        if !errors.Is(err, os.ErrNotExist) {
            t.Errorf("Expected file not found error, got %v", err)
        }
    })
}
```

### Testing Concurrent Code

```go
func TestConcurrentAccess(t *testing.T) {
    cache := NewCache()
    
    var wg sync.WaitGroup
    numGoroutines := 10
    numOperations := 100
    
    // Start multiple goroutines
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            for j := 0; j < numOperations; j++ {
                key := fmt.Sprintf("key-%d-%d", id, j)
                value := fmt.Sprintf("value-%d-%d", id, j)
                
                cache.Set(key, value)
                retrieved := cache.Get(key)
                
                if retrieved != value {
                    t.Errorf("Expected %s, got %s", value, retrieved)
                }
            }
        }(i)
    }
    
    wg.Wait()
}
```

## Continuous Integration

### GitHub Actions Configuration

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.21, 1.22]
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Cache dependencies
      uses: actions/cache@v3
      with:
        path: |
          ~/.cache/go-build
          ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run tests
      run: |
        go test -race -coverprofile=coverage.out ./...
    
    - name: Check coverage
      run: |
        go tool cover -func=coverage.out
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

### Pre-commit Hooks

```bash
#!/bin/sh
# .git/hooks/pre-commit

# Run tests before commit
echo "Running tests..."
if ! make test; then
    echo "Tests failed. Commit aborted."
    exit 1
fi

# Check formatting
echo "Checking formatting..."
if ! go fmt ./...; then
    echo "Code formatting issues found. Please run 'go fmt ./...' and try again."
    exit 1
fi

# Run linting
echo "Running linter..."
if ! golangci-lint run; then
    echo "Linting issues found. Please fix and try again."
    exit 1
fi

echo "All checks passed!"
```

## Performance Testing

### Benchmarking Guidelines

```go
func BenchmarkCriticalFunction(b *testing.B) {
    // Set up test data
    testData := generateTestData(1000)
    
    // Reset timer to exclude setup time
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        CriticalFunction(testData)
    }
}

func BenchmarkWithDifferentSizes(b *testing.B) {
    sizes := []int{10, 100, 1000, 10000}
    
    for _, size := range sizes {
        b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
            data := generateTestData(size)
            b.ResetTimer()
            
            for i := 0; i < b.N; i++ {
                ProcessData(data)
            }
        })
    }
}
```

### Memory Profiling

```bash
# Run benchmarks with memory profiling
go test -bench=. -memprofile=mem.prof ./pkg/repo

# Analyze memory profile
go tool pprof mem.prof
> top10
> list FunctionName
```

### CPU Profiling

```bash
# Run benchmarks with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./pkg/repo

# Analyze CPU profile
go tool pprof cpu.prof
> top10
> web
```

## Troubleshooting Tests

### Common Issues

#### Flaky Tests

```go
// Bad: Time-dependent test
func TestTimeout(t *testing.T) {
    start := time.Now()
    ProcessWithTimeout(100 * time.Millisecond)
    duration := time.Since(start)
    
    if duration > 105*time.Millisecond {
        t.Error("Timeout took too long")
    }
}

// Good: Use channels for synchronization
func TestTimeout(t *testing.T) {
    done := make(chan bool, 1)
    
    go func() {
        ProcessWithTimeout(100 * time.Millisecond)
        done <- true
    }()
    
    select {
    case <-done:
        // Test passed
    case <-time.After(200 * time.Millisecond):
        t.Error("Timeout exceeded")
    }
}
```

#### Resource Leaks

```go
func TestDatabaseOperations(t *testing.T) {
    db := setupTestDatabase(t)
    
    // Ensure cleanup happens even if test fails
    t.Cleanup(func() {
        if err := db.Close(); err != nil {
            t.Errorf("Failed to close database: %v", err)
        }
    })
    
    // Test database operations
}
```

#### Path Issues

```go
func TestFileOperations(t *testing.T) {
    // Use filepath.Join for cross-platform paths
    testFile := filepath.Join("testdata", "config", "test.json")
    
    // Use t.TempDir() for temporary files
    tempDir := t.TempDir()
    outputFile := filepath.Join(tempDir, "output.txt")
    
    // Test operations
}
```

### Debugging Tests

```bash
# Run single test with verbose output
go test -v ./pkg/config -run TestSpecificTest

# Print additional debugging info
go test -v ./pkg/config -run TestSpecificTest -args -debug

# Use dlv debugger
dlv test ./pkg/config -- -test.run TestSpecificTest
```

### Test Environment Variables

```go
func TestWithEnvironment(t *testing.T) {
    // Save original environment
    original := os.Getenv("TEST_ENV_VAR")
    defer os.Setenv("TEST_ENV_VAR", original)
    
    // Set test environment
    os.Setenv("TEST_ENV_VAR", "test-value")
    
    // Run test
    result := FunctionThatUsesEnvVar()
    
    // Verify result
    if result != "expected" {
        t.Errorf("Expected 'expected', got %s", result)
    }
}
```

## Conclusion

Comprehensive testing is crucial for maintaining code quality and preventing regressions. Follow these guidelines to write effective tests that provide confidence in the Waffles codebase while being maintainable and efficient.

Remember:
- Write tests first when possible (TDD)
- Keep tests simple and focused
- Use descriptive names and clear assertions
- Mock external dependencies appropriately
- Maintain good test coverage
- Run tests frequently during development

For questions about testing practices or help with specific testing scenarios, refer to the [Contributing Guide](contributing.md) or open a discussion in the project repository.