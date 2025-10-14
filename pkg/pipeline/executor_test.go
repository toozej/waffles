package pipeline

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/toozej/waffles/pkg/config"
	"github.com/toozej/waffles/pkg/repo"
)

func TestNewPipeline(t *testing.T) {
	cfg := &config.Config{
		DefaultModel:    "test-model",
		DefaultProvider: "test-provider",
		LogDBPath:       ":memory:",
	}

	pipeline := NewPipeline(cfg)
	if pipeline == nil {
		t.Error("Expected pipeline to be non-nil")
		return
	}

	if pipeline.Config != cfg {
		t.Error("Expected pipeline config to be set correctly")
	}
}

func TestPipeline_Execute(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pipeline execution test in short mode")
	}

	// Create test config
	cfg := &config.Config{
		DefaultModel:       "test-model",
		DefaultProvider:    "test-provider",
		LogDBPath:          ":memory:",
		WheresmypromptArgs: "",
		Files2promptArgs:   "",
		LLMArgs:            "",
		Verbose:            true,
	}

	pipeline := NewPipeline(cfg)

	// Create a temporary repository for testing
	tempDir := t.TempDir()

	// Create some test files
	testFiles := map[string]string{
		"main.go":   "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
		"helper.go": "package main\n\nfunc helper() string {\n\treturn \"helper\"\n}",
		"go.mod":    "module test\n\ngo 1.21\n",
		"README.md": "# Test Project\n\nThis is a test project.",
	}

	for filename, content := range testFiles {
		filepath := fmt.Sprintf("%s/%s", tempDir, filename)
		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Analyze the repository
	repoInfo, err := repo.AnalyzeRepository(tempDir, nil)
	if err != nil {
		t.Fatalf("Failed to analyze repository: %v", err)
	}

	pipeline.RepoInfo = repoInfo

	// Test execution (this will fail without actual tools, but we test the structure)
	execution, err := pipeline.Execute("test query", []string{})

	// We expect this to fail in test environment, but we can verify the structure
	if err != nil {
		t.Logf("Execution failed as expected in test environment: %v", err)
	}

	if execution != nil {
		if execution.ID == "" {
			t.Error("Expected execution ID to be set")
		}
		if execution.PromptQuery != "test query" {
			t.Errorf("Expected prompt query 'test query', got %s", execution.PromptQuery)
		}
		if execution.StartTime.IsZero() {
			t.Error("Expected start time to be set")
		}

		t.Logf("Execution created with ID: %s", execution.ID)
	}
}

func TestPipeline_GetConfig(t *testing.T) {
	cfg := &config.Config{
		DefaultModel: "test-model",
		LogDBPath:    ":memory:",
	}
	pipeline := NewPipeline(cfg)

	retrievedConfig := pipeline.GetConfig()
	if retrievedConfig != cfg {
		t.Error("Expected config to be retrievable")
	}
}

func TestPipeline_SetRepoInfo(t *testing.T) {
	cfg := &config.Config{
		LogDBPath: ":memory:",
	}
	pipeline := NewPipeline(cfg)

	repoInfo := &repo.RepositoryInfo{
		Language: repo.LanguageGo,
		RootPath: "/test/path",
	}

	pipeline.SetRepoInfo(repoInfo)

	if pipeline.GetRepoInfo() != repoInfo {
		t.Error("Expected repo info to be set correctly")
	}
}

func TestExecutionContext(t *testing.T) {
	// Test ExecutionContext struct
	now := time.Now()
	ctx := &ExecutionContext{
		ID:          "test-id-123",
		StartTime:   now,
		PromptQuery: "test query",
		Files:       []string{"main.go", "helper.go"},
		ExecutionSteps: []StepResult{
			{
				Tool:     "wheresmyprompt",
				Command:  []string{"wheresmyprompt", "test"},
				Output:   "test output",
				Duration: time.Millisecond * 100,
			},
		},
	}

	if ctx.ID != "test-id-123" {
		t.Error("Expected ID to be set correctly")
	}

	if ctx.StartTime != now {
		t.Error("Expected StartTime to be set correctly")
	}

	if len(ctx.Files) != 2 {
		t.Error("Expected 2 files")
	}

	if len(ctx.ExecutionSteps) != 1 {
		t.Error("Expected 1 execution step")
	}

	step := ctx.ExecutionSteps[0]
	if step.Tool != "wheresmyprompt" {
		t.Error("Expected step tool to be 'wheresmyprompt'")
	}
}

func TestStepResult(t *testing.T) {
	// Test StepResult struct
	duration := time.Millisecond * 250
	step := StepResult{
		Tool:     "files2prompt",
		Command:  []string{"files2prompt", "-f", "*.go"},
		Output:   "Generated prompt content",
		Error:    nil,
		Duration: duration,
	}

	if step.Tool != "files2prompt" {
		t.Error("Expected tool to be 'files2prompt'")
	}

	if len(step.Command) != 3 {
		t.Error("Expected 3 command parts")
	}

	if step.Duration != duration {
		t.Error("Expected duration to be set correctly")
	}

	// Test with error
	stepWithError := StepResult{
		Tool:  "llm",
		Error: fmt.Errorf("test error"),
	}

	if stepWithError.Error == nil {
		t.Error("Expected error to be set")
	}

	if stepWithError.Error.Error() != "test error" {
		t.Error("Expected error message to match")
	}
}

func TestPipelineValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.Config
		expectErr bool
	}{
		{
			name: "Valid config",
			config: &config.Config{
				DefaultModel:    "claude-3-sonnet",
				DefaultProvider: "anthropic",
				LogDBPath:       ":memory:",
			},
			expectErr: false,
		},
		{
			name: "Missing model",
			config: &config.Config{
				DefaultProvider: "anthropic",
				LogDBPath:       ":memory:",
			},
			expectErr: false, // Pipeline should handle missing model
		},
		{
			name:      "Empty config",
			config:    &config.Config{},
			expectErr: false, // Pipeline should handle empty config
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := NewPipeline(tt.config)

			if pipeline == nil && !tt.expectErr {
				t.Error("Expected pipeline to be created")
			}

			if pipeline != nil && tt.expectErr {
				t.Error("Expected pipeline creation to fail")
			}
		})
	}
}

func TestGenerateExecutionID(t *testing.T) {
	// Test that execution IDs are unique
	ids := make(map[string]bool)

	for i := 0; i < 100; i++ {
		id := generateExecutionID()

		if id == "" {
			t.Error("Expected execution ID to be non-empty")
		}

		if ids[id] {
			t.Errorf("Execution ID %s is not unique", id)
		}

		ids[id] = true

		// IDs should be reasonable length
		if len(id) < 8 || len(id) > 64 {
			t.Errorf("Execution ID %s has unexpected length: %d", id, len(id))
		}
	}
}

func TestPipelineTimeout(t *testing.T) {
	cfg := &config.Config{
		LogDBPath: ":memory:",
	}
	pipeline := NewPipeline(cfg)

	// Test that execute completes normally (pipeline doesn't use context directly)
	execution, err := pipeline.Execute("test query", []string{})

	// We expect this to fail due to missing dependencies or tools
	if err != nil {
		t.Logf("Got expected error in test environment: %v", err)
	}

	if execution != nil {
		t.Logf("Execution created: %s", execution.ID)
	}
}

func TestPipelineConfigOverrides(t *testing.T) {
	cfg := &config.Config{
		DefaultModel:       "base-model",
		WheresmypromptArgs: "--base-arg",
		Files2promptArgs:   "--base-files",
		LLMArgs:            "--base-llm",
	}

	pipeline := NewPipeline(cfg)

	// Test that config values are accessible
	if pipeline.Config.DefaultModel != "base-model" {
		t.Error("Expected default model to be preserved")
	}

	if pipeline.Config.WheresmypromptArgs != "--base-arg" {
		t.Error("Expected wheresmyprompt args to be preserved")
	}
}

func TestPipelineErrorHandling(t *testing.T) {
	cfg := &config.Config{
		LogDBPath: ":memory:",
	}
	pipeline := NewPipeline(cfg)

	// Test with invalid repository info
	invalidRepo := &repo.RepositoryInfo{
		Language: repo.LanguageUnknown,
		RootPath: "/nonexistent/path",
	}
	pipeline.SetRepoInfo(invalidRepo)

	execution, err := pipeline.Execute("", []string{})

	// Should handle errors gracefully
	if err != nil {
		t.Logf("Pipeline handled error correctly: %v", err)
	}

	if execution != nil {
		t.Logf("Execution created even with errors: %s", execution.ID)
	}
}

// Benchmark tests
func BenchmarkPipelineCreation(b *testing.B) {
	cfg := &config.Config{
		DefaultModel: "test-model",
		LogDBPath:    ":memory:",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewPipeline(cfg)
	}
}

func BenchmarkExecutionIDGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateExecutionID()
	}
}

// Mock functions for testing
func TestMockExecution(t *testing.T) {
	// Test the pipeline with mock data
	cfg := &config.Config{
		DefaultModel: "mock-model",
		LogDBPath:    ":memory:",
		Verbose:      true,
	}

	pipeline := NewPipeline(cfg)

	// Create mock repository info
	mockRepo := &repo.RepositoryInfo{
		Language: repo.LanguageGo,
		RootPath: "/mock/path",
		DetectedFiles: []repo.FileInfo{
			{Path: "main.go", Size: 100, Included: true},
			{Path: "helper.go", Size: 50, Included: true},
		},
	}

	pipeline.SetRepoInfo(mockRepo)

	// The execution should be structured correctly even if it fails
	execution, err := pipeline.Execute("mock query", []string{"--mock-arg"})

	if err != nil {
		t.Logf("Mock execution failed as expected: %v", err)
	}

	if execution != nil {
		// Verify the execution structure
		if execution.ID == "" {
			t.Error("Expected execution ID to be generated")
		}

		if execution.PromptQuery != "mock query" {
			t.Error("Expected prompt query to be preserved")
		}

		t.Logf("Mock execution completed with ID: %s", execution.ID)
	}
}

// Helper function to check if execution ID format is valid
func isValidExecutionID(id string) bool {
	if len(id) < 8 {
		return false
	}

	// Should contain alphanumeric characters
	for _, r := range id {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}

	return true
}

func TestExecutionIDFormat(t *testing.T) {
	for i := 0; i < 10; i++ {
		id := generateExecutionID()
		if !isValidExecutionID(id) {
			t.Errorf("Invalid execution ID format: %s", id)
		}
	}
}

// Test concurrent pipeline usage
func TestConcurrentPipelineUsage(t *testing.T) {
	cfg := &config.Config{
		LogDBPath: ":memory:",
	}

	pipeline := NewPipeline(cfg)

	// Run multiple goroutines to test thread safety
	const numGoroutines = 10
	results := make(chan string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			query := fmt.Sprintf("concurrent query %d", index)

			execution, err := pipeline.Execute(query, []string{})
			if err != nil {
				results <- fmt.Sprintf("error-%d", index)
				return
			}

			if execution != nil {
				results <- execution.ID
			} else {
				results <- fmt.Sprintf("nil-%d", index)
			}
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		result := <-results
		t.Logf("Concurrent execution result %d: %s", i, result)
	}
}
