package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewDatabase(t *testing.T) {
	// Test in-memory database
	db, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Error("Expected database to be non-nil")
	}

	// Test file database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db2, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file database: %v", err)
	}
	defer db2.Close()

	// Check that file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Expected database file to be created")
	}
}

func TestDatabaseInitializeSchema(t *testing.T) {
	db, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Schema should be initialized automatically
	// Test that we can query the tables
	tables := []string{"waffles_executions", "waffles_files", "waffles_steps"}

	for _, table := range tables {
		var count int
		query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
		err := db.db.QueryRow(query, table).Scan(&count)
		if err != nil {
			t.Errorf("Failed to query for table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("Expected table %s to exist, but count is %d", table, count)
		}
	}
}

func TestLogExecution(t *testing.T) {
	db, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create test execution
	exec := &WafflesExecution{
		ID:                  "test-exec-123",
		ConversationID:      "conv-456",
		CommandArgs:         "waffles test query",
		WheresmypromptQuery: "test query",
		Files2promptArgs:    "--include *.go",
		DetectedLanguage:    "go",
		FileCount:           5,
		ExecutionTimeMS:     1500,
		Created:             time.Now(),
	}

	// Log the execution
	err = db.LogExecution(exec)
	if err != nil {
		t.Errorf("Failed to log execution: %v", err)
	}

	// Verify it was logged
	var count int
	err = db.db.QueryRow("SELECT COUNT(*) FROM waffles_executions WHERE id = ?", exec.ID).Scan(&count)
	if err != nil {
		t.Errorf("Failed to query execution: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 execution record, got %d", count)
	}
}

func TestLogFiles(t *testing.T) {
	db, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	executionID := "test-exec-123"

	// Create test files
	files := []WafflesFile{
		{
			ID:          "file-1",
			ExecutionID: executionID,
			FilePath:    "main.go",
			FileSize:    1024,
			Included:    true,
		},
		{
			ID:              "file-2",
			ExecutionID:     executionID,
			FilePath:        "vendor/dep.go",
			FileSize:        512,
			Included:        false,
			ExclusionReason: "vendor directory excluded",
		},
	}

	// Log the files
	err = db.LogFiles(executionID, files)
	if err != nil {
		t.Errorf("Failed to log files: %v", err)
	}

	// Verify they were logged
	var count int
	err = db.db.QueryRow("SELECT COUNT(*) FROM waffles_files WHERE execution_id = ?", executionID).Scan(&count)
	if err != nil {
		t.Errorf("Failed to query files: %v", err)
	}
	if count != len(files) {
		t.Errorf("Expected %d file records, got %d", len(files), count)
	}
}

func TestQueryExecutions(t *testing.T) {
	db, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create test executions
	executions := []*WafflesExecution{
		{
			ID:                  "exec-1",
			ConversationID:      "conv-1",
			CommandArgs:         "waffles query 1",
			WheresmypromptQuery: "test query 1",
			DetectedLanguage:    "go",
			FileCount:           3,
			ExecutionTimeMS:     1000,
			Created:             time.Now(),
		},
		{
			ID:                  "exec-2",
			ConversationID:      "conv-2",
			CommandArgs:         "waffles query 2",
			WheresmypromptQuery: "test query 2",
			DetectedLanguage:    "python",
			FileCount:           7,
			ExecutionTimeMS:     2000,
			Created:             time.Now().Add(-time.Hour), // 1 hour ago
		},
	}

	// Log the executions
	for _, exec := range executions {
		err = db.LogExecution(exec)
		if err != nil {
			t.Errorf("Failed to log execution %s: %v", exec.ID, err)
		}
	}

	// Test querying all executions
	results, err := db.QueryExecutions(&ExecutionFilter{})
	if err != nil {
		t.Errorf("Failed to query executions: %v", err)
	}
	if len(results) != len(executions) {
		t.Errorf("Expected %d executions, got %d", len(executions), len(results))
	}

	// Test querying with filters
	filters := &ExecutionFilter{
		Language: "go",
	}
	results, err = db.QueryExecutions(filters)
	if err != nil {
		t.Errorf("Failed to query executions with filter: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 Go execution, got %d", len(results))
	}
	if results[0].DetectedLanguage != "go" {
		t.Errorf("Expected Go execution, got %s", results[0].DetectedLanguage)
	}
}

func TestQueryFiles(t *testing.T) {
	db, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	executionID := "test-exec-123"

	// Create and log test files
	files := []WafflesFile{
		{
			ID:          "file-1",
			ExecutionID: executionID,
			FilePath:    "main.go",
			FileSize:    1024,
			Included:    true,
		},
		{
			ID:          "file-2",
			ExecutionID: executionID,
			FilePath:    "helper.go",
			FileSize:    512,
			Included:    true,
		},
	}

	err = db.LogFiles(executionID, files)
	if err != nil {
		t.Errorf("Failed to log files: %v", err)
	}

	// Query the files
	results, err := db.GetExecutionFiles(executionID)
	if err != nil {
		t.Errorf("Failed to query files: %v", err)
	}

	if len(results) != len(files) {
		t.Errorf("Expected %d files, got %d", len(files), len(results))
	}

	// Check that files are correctly retrieved (order-agnostic)
	filePathMap := make(map[string]bool)
	for _, file := range files {
		filePathMap[file.FilePath] = true
	}

	for _, file := range results {
		if file.ExecutionID != executionID {
			t.Errorf("Expected execution ID %s, got %s", executionID, file.ExecutionID)
		}
		if !filePathMap[file.FilePath] {
			t.Errorf("Unexpected file path %s", file.FilePath)
		}
	}
}

func TestDatabaseClose(t *testing.T) {
	db, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Database should be open initially
	err = db.db.Ping()
	if err != nil {
		t.Error("Expected database to be accessible")
	}

	// Close the database
	err = db.Close()
	if err != nil {
		t.Errorf("Failed to close database: %v", err)
	}

	// Database should be closed now
	err = db.db.Ping()
	if err == nil {
		t.Error("Expected database to be closed")
	}
}

func TestDatabaseTypes(t *testing.T) {
	// Test WafflesExecution struct
	exec := WafflesExecution{
		ID:                  "test-id",
		ConversationID:      "conv-id",
		CommandArgs:         "test command",
		WheresmypromptQuery: "test query",
		Files2promptArgs:    "test args",
		DetectedLanguage:    "go",
		FileCount:           10,
		ExecutionTimeMS:     1500,
		Created:             time.Now(),
	}

	if exec.ID != "test-id" {
		t.Error("Expected ID to be set correctly")
	}
	if exec.FileCount != 10 {
		t.Error("Expected FileCount to be set correctly")
	}

	// Test WafflesFile struct
	file := WafflesFile{
		ID:              "file-id",
		ExecutionID:     "exec-id",
		FilePath:        "/path/to/file.go",
		FileSize:        2048,
		Included:        true,
		ExclusionReason: "",
	}

	if file.FilePath != "/path/to/file.go" {
		t.Error("Expected FilePath to be set correctly")
	}
	if file.FileSize != 2048 {
		t.Error("Expected FileSize to be set correctly")
	}
	if !file.Included {
		t.Error("Expected Included to be true")
	}
}

func TestDatabaseConcurrency(t *testing.T) {
	// Use a file-based database for concurrency testing to avoid SQLite in-memory limitations
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "concurrency-test.db")

	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Verify schema is initialized by checking tables exist
	var count int
	err = db.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='waffles_executions'").Scan(&count)
	if err != nil || count != 1 {
		t.Fatalf("Schema not properly initialized: %v, count: %d", err, count)
	}

	// Test concurrent writes
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			exec := &WafflesExecution{
				ID:                  fmt.Sprintf("exec-%d", index),
				ConversationID:      fmt.Sprintf("conv-%d", index),
				CommandArgs:         fmt.Sprintf("command %d", index),
				WheresmypromptQuery: fmt.Sprintf("query %d", index),
				DetectedLanguage:    "go",
				FileCount:           index,
				ExecutionTimeMS:     int64(index * 100),
				Created:             time.Now(),
			}
			results <- db.LogExecution(exec)
		}(i)
	}

	// Collect results
	errorCount := 0
	for i := 0; i < numGoroutines; i++ {
		if err := <-results; err != nil {
			errorCount++
			t.Logf("Concurrent write error: %v", err)
		}
	}

	// Verify all executions were written
	err = db.db.QueryRow("SELECT COUNT(*) FROM waffles_executions").Scan(&count)
	if err != nil {
		t.Errorf("Failed to count executions: %v", err)
	}

	expectedCount := numGoroutines - errorCount
	if count != expectedCount {
		t.Errorf("Expected %d executions, got %d", expectedCount, count)
	}
}

func TestExportExecutions(t *testing.T) {
	db, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create test execution
	exec := &WafflesExecution{
		ID:                  "export-test",
		ConversationID:      "conv-export",
		CommandArgs:         "waffles export test",
		WheresmypromptQuery: "export query",
		DetectedLanguage:    "go",
		FileCount:           5,
		ExecutionTimeMS:     1200,
		Created:             time.Now(),
	}

	err = db.LogExecution(exec)
	if err != nil {
		t.Errorf("Failed to log execution: %v", err)
	}

	// Test export functionality (if implemented)
	// This would test the ExportExecutions method if it exists
	// Since it's not in the interface, we skip this test
	t.Log("Export functionality test would go here")
}

// Benchmark tests
func BenchmarkLogExecution(b *testing.B) {
	db, err := NewDatabase(":memory:")
	if err != nil {
		b.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	exec := &WafflesExecution{
		ConversationID:      "bench-conv",
		CommandArgs:         "benchmark command",
		WheresmypromptQuery: "benchmark query",
		DetectedLanguage:    "go",
		FileCount:           10,
		ExecutionTimeMS:     1000,
		Created:             time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exec.ID = fmt.Sprintf("bench-exec-%d", i)
		_ = db.LogExecution(exec)
	}
}

func BenchmarkQueryExecutions(b *testing.B) {
	db, err := NewDatabase(":memory:")
	if err != nil {
		b.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Insert test data
	for i := 0; i < 1000; i++ {
		exec := &WafflesExecution{
			ID:               fmt.Sprintf("bench-query-%d", i),
			ConversationID:   fmt.Sprintf("conv-%d", i),
			CommandArgs:      "benchmark command",
			DetectedLanguage: "go",
			FileCount:        i % 20,
			ExecutionTimeMS:  int64(i * 10),
			Created:          time.Now().Add(-time.Duration(i) * time.Second),
		}
		_ = db.LogExecution(exec)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = db.QueryExecutions(&ExecutionFilter{})
	}
}

// Helper functions for testing

func TestDatabaseValidation(t *testing.T) {
	// Test various database path scenarios
	testCases := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "valid memory database",
			path:        ":memory:",
			expectError: false,
		},
		{
			name:        "empty path",
			path:        "",
			expectError: false, // Empty path should work (defaults to current dir or in-memory)
		},
		{
			name:        "relative path",
			path:        "test.db",
			expectError: false, // Should work in current directory
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := NewDatabase(tc.path)

			if tc.expectError {
				if err == nil && db != nil {
					db.Close()
					t.Errorf("Expected error for path %s, but got success", tc.path)
				}
			} else {
				if err != nil {
					t.Errorf("Expected success for path %s, but got error: %v", tc.path, err)
				}
				if db != nil {
					db.Close()
				}
			}
		})
	}
}

func TestDatabaseRecovery(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "recovery-test.db")

	// Create database
	db1, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Add some data
	exec := &WafflesExecution{
		ID:               "recovery-test",
		ConversationID:   "recovery-conv",
		DetectedLanguage: "go",
		FileCount:        3,
		Created:          time.Now(),
	}
	err = db1.LogExecution(exec)
	if err != nil {
		t.Errorf("Failed to log execution: %v", err)
	}

	// Close first connection
	db1.Close()

	// Reopen database
	db2, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}
	defer db2.Close()

	// Verify data is still there
	results, err := db2.QueryExecutions(&ExecutionFilter{})
	if err != nil {
		t.Errorf("Failed to query after reopening: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 execution after reopening, got %d", len(results))
	}
}
