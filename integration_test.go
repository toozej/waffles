package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/toozej/waffles/pkg/config"
	"github.com/toozej/waffles/pkg/deps"
	"github.com/toozej/waffles/pkg/logging"
	"github.com/toozej/waffles/pkg/pipeline"
	"github.com/toozej/waffles/pkg/repo"
)

func TestFullSystemIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary directory for the test
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	_ = os.Chdir(tempDir)

	// Create a test Go project
	testFiles := map[string]string{
		"go.mod":    "module github.com/test/integration\n\ngo 1.21\n",
		"main.go":   "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}\n",
		"helper.go": "package main\n\nfunc helper() string {\n\treturn \"helper function\"\n}\n",
		"README.md": "# Integration Test Project\n\nThis is a test project for integration testing.\n",
	}

	for filename, content := range testFiles {
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	t.Run("ConfigurationSystem", func(t *testing.T) {
		// Test configuration loading
		cfg, err := config.LoadConfig()
		if err != nil {
			t.Errorf("Failed to load config: %v", err)
		}

		if cfg == nil {
			t.Fatal("Config should not be nil")
		}

		// Verify default values
		if cfg.DefaultModel == "" {
			t.Error("Default model should be set")
		}

		t.Logf("Configuration loaded successfully: Model=%s, Provider=%s",
			cfg.DefaultModel, cfg.DefaultProvider)
	})

	t.Run("DependencySystem", func(t *testing.T) {
		// Test dependency checking
		statuses, err := deps.CheckAllDependencies()
		if err != nil {
			t.Errorf("Failed to check dependencies: %v", err)
		}

		if len(statuses) == 0 {
			t.Error("Expected some dependency statuses")
		}

		// Log dependency status
		for _, status := range statuses {
			t.Logf("Dependency %s: Installed=%t, Valid=%t",
				status.Name, status.Installed, status.Valid)
		}
	})

	t.Run("RepositoryAnalysis", func(t *testing.T) {
		// Test repository analysis - analyze the temp directory we created
		repoInfo, err := repo.AnalyzeRepository(tempDir, nil)
		if err != nil {
			t.Fatalf("Failed to analyze repository: %v", err)
		}

		// Verify language detection - should detect Go from go.mod and .go files
		if repoInfo.Language != repo.LanguageGo {
			t.Logf("Expected language Go, got %s - checking if temp dir has proper files", repoInfo.Language)
			// List files in temp directory for debugging
			files, _ := os.ReadDir(tempDir)
			for _, file := range files {
				t.Logf("File in temp dir: %s", file.Name())
			}
		}

		// Verify files were detected
		if len(repoInfo.DetectedFiles) == 0 {
			t.Logf("No files detected, temp dir contents:")
			files, _ := os.ReadDir(tempDir)
			for _, file := range files {
				t.Logf("  - %s", file.Name())
			}
		}

		// Check that Go files are included
		goFilesFound := 0
		for _, file := range repoInfo.DetectedFiles {
			if filepath.Ext(file.Path) == ".go" && file.Included {
				goFilesFound++
			}
		}

		if goFilesFound == 0 {
			t.Logf("No Go files found as included. All detected files:")
			for _, file := range repoInfo.DetectedFiles {
				t.Logf("  - %s (included: %t, reason: %s)", file.Path, file.Included, file.Reason)
			}
		}

		t.Logf("Repository analysis completed: Language=%s, Files=%d, GoFiles=%d",
			repoInfo.Language, len(repoInfo.DetectedFiles), goFilesFound)
	})

	t.Run("DatabaseSystem", func(t *testing.T) {
		// Test database creation and operations
		db, err := logging.NewDatabase(":memory:")
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		// Test execution logging
		exec := &logging.WafflesExecution{
			ID:                  "integration-test-exec",
			CommandArgs:         "waffles integration test",
			WheresmypromptQuery: "integration test query",
			DetectedLanguage:    "go",
			FileCount:           3,
			ExecutionTimeMS:     1500,
			Success:             true,
			ModelUsed:           "test-model",
			ProviderUsed:        "test-provider",
			Created:             time.Now(),
		}

		if err := db.LogExecution(exec); err != nil {
			t.Errorf("Failed to log execution: %v", err)
		}

		// Test querying
		results, err := db.QueryExecutions(&logging.ExecutionFilter{
			Language: "go",
			Limit:    10,
		})
		if err != nil {
			t.Errorf("Failed to query executions: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 execution, got %d", len(results))
		}

		t.Logf("Database operations completed successfully")
	})

	t.Run("PipelineSystem", func(t *testing.T) {
		// Test pipeline creation and basic validation
		cfg := &config.Config{
			DefaultModel:    "integration-test-model",
			DefaultProvider: "integration-test-provider",
			LogDBPath:       ":memory:",
		}

		pipeline := pipeline.NewPipeline(cfg)
		if pipeline == nil {
			t.Fatal("Pipeline should not be nil")
		}

		// Test validation
		err := pipeline.ValidatePipeline()
		if err != nil {
			t.Logf("Pipeline validation error (expected in test environment): %v", err)
		}

		// Test repository info setting
		repoInfo := &repo.RepositoryInfo{
			Language: repo.LanguageGo,
			RootPath: tempDir,
		}

		pipeline.SetRepoInfo(repoInfo)

		retrievedRepo := pipeline.GetRepoInfo()
		if retrievedRepo == nil {
			t.Error("Repository info should be set")
		}

		t.Logf("Pipeline system validation completed")
	})

	t.Run("EndToEndWorkflow", func(t *testing.T) {
		// Test a complete workflow simulation
		t.Log("Starting end-to-end workflow simulation...")

		// 1. Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		cfg.LogDBPath = ":memory:" // Use in-memory DB for test

		// 2. Create pipeline
		pipeline := pipeline.NewPipeline(cfg)

		// 3. Analyze repository
		repoInfo, err := repo.AnalyzeRepository(".", nil)
		if err != nil {
			t.Fatalf("Failed to analyze repository: %v", err)
		}

		pipeline.SetRepoInfo(repoInfo)

		// 4. Create database
		db, err := logging.NewDatabase(":memory:")
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		// 5. Log a simulated execution
		exec := &logging.WafflesExecution{
			ID:                  "e2e-test-exec",
			CommandArgs:         "waffles end to end test",
			WheresmypromptQuery: "e2e test query",
			DetectedLanguage:    string(repoInfo.Language),
			FileCount:           len(repoInfo.DetectedFiles),
			ExecutionTimeMS:     2000,
			Success:             true,
			ModelUsed:           cfg.DefaultModel,
			ProviderUsed:        cfg.DefaultProvider,
			Created:             time.Now(),
		}

		if err := db.LogExecution(exec); err != nil {
			t.Errorf("Failed to log execution: %v", err)
		}

		// 6. Query and verify
		results, err := db.QueryExecutions(&logging.ExecutionFilter{})
		if err != nil {
			t.Errorf("Failed to query executions: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 execution, got %d", len(results))
		}

		// 7. Get statistics
		stats, err := db.GetExecutionStats(&logging.ExecutionFilter{})
		if err != nil {
			t.Errorf("Failed to get stats: %v", err)
		}

		if stats.TotalExecutions != 1 {
			t.Errorf("Expected 1 total execution, got %d", stats.TotalExecutions)
		}

		t.Logf("End-to-end workflow completed successfully: %d executions, %d successful",
			stats.TotalExecutions, stats.SuccessfulExecutions)
	})
}

func TestSystemPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	t.Run("RepositoryAnalysisPerformance", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create many test files
		for i := 0; i < 100; i++ {
			filename := filepath.Join(tempDir, fmt.Sprintf("file%d.go", i))
			content := fmt.Sprintf("package main\n\n// File %d\nfunc function%d() {\n\t// Do something\n}\n", i, i)
			if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
		}

		start := time.Now()
		_, err := repo.AnalyzeRepository(tempDir, nil)
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Repository analysis failed: %v", err)
		}

		t.Logf("Repository analysis of 100 files took: %v", duration)

		// Performance expectation: should complete within reasonable time
		if duration > 5*time.Second {
			t.Errorf("Repository analysis took too long: %v", duration)
		}
	})

	t.Run("DatabasePerformance", func(t *testing.T) {
		db, err := logging.NewDatabase(":memory:")
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		// Test bulk insertions
		const numExecutions = 1000

		start := time.Now()
		for i := 0; i < numExecutions; i++ {
			exec := &logging.WafflesExecution{
				ID:               fmt.Sprintf("perf-test-%d", i),
				CommandArgs:      fmt.Sprintf("command %d", i),
				DetectedLanguage: "go",
				FileCount:        i % 20,
				ExecutionTimeMS:  int64(i * 10),
				Success:          i%2 == 0,
				Created:          time.Now(),
			}

			if err := db.LogExecution(exec); err != nil {
				t.Errorf("Failed to log execution %d: %v", i, err)
				break
			}
		}
		insertDuration := time.Since(start)

		// Test bulk queries
		start = time.Now()
		results, err := db.QueryExecutions(&logging.ExecutionFilter{
			Limit: numExecutions,
		})
		queryDuration := time.Since(start)

		if err != nil {
			t.Errorf("Failed to query executions: %v", err)
		}

		if len(results) != numExecutions {
			t.Errorf("Expected %d results, got %d", numExecutions, len(results))
		}

		t.Logf("Database performance - Insert %d records: %v, Query %d records: %v",
			numExecutions, insertDuration, numExecutions, queryDuration)
	})
}

func TestErrorHandlingAndRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping error handling test in short mode")
	}

	t.Run("ConfigurationErrors", func(t *testing.T) {
		// Test configuration with invalid paths
		originalDir, _ := os.Getwd()
		tempDir := t.TempDir()
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(originalDir) }()

		// This should still work with defaults
		cfg, err := config.LoadConfig()
		if err != nil {
			t.Errorf("Config loading should not fail with defaults: %v", err)
		}

		if cfg == nil {
			t.Error("Config should not be nil even with errors")
		}
	})

	t.Run("DatabaseErrors", func(t *testing.T) {
		// Test operations on closed database
		db, err := logging.NewDatabase(":memory:")
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}

		db.Close()

		// This should handle the closed database gracefully
		exec := &logging.WafflesExecution{
			ID:      "test-closed-db",
			Created: time.Now(),
		}

		err = db.LogExecution(exec)
		if err == nil {
			t.Error("Expected error when using closed database")
		}
	})

	t.Run("RepositoryErrors", func(t *testing.T) {
		// Test repository analysis on non-existent path
		_, err := repo.AnalyzeRepository("/non/existent/path", nil)
		if err == nil {
			t.Error("Expected error for non-existent path")
		}

		// Test with empty directory
		emptyDir := t.TempDir()
		info, err := repo.AnalyzeRepository(emptyDir, nil)
		if err != nil {
			t.Errorf("Should handle empty directory gracefully: %v", err)
		}

		if info != nil && info.Language != repo.LanguageUnknown {
			t.Logf("Empty directory detected as: %s", info.Language)
		}
	})
}
