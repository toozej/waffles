package deps

import (
	"os/exec"
	"testing"
)

func TestCheckDependency(t *testing.T) {
	tests := []struct {
		name        string
		dep         Dependency
		expectError bool
		setup       func()
		cleanup     func()
	}{
		{
			name: "Valid dependency with command available",
			dep: Dependency{
				Name:         "test-tool",
				Command:      "echo",
				MinVersion:   "1.0.0",
				CheckCommand: "echo --version",
			},
			expectError: false,
		},
		{
			name: "Invalid dependency with command not available",
			dep: Dependency{
				Name:         "nonexistent-tool",
				Command:      "nonexistent-command-xyz123",
				MinVersion:   "1.0.0",
				CheckCommand: "nonexistent-command-xyz123 --version",
			},
			expectError: false, // CheckDependency doesn't error, it returns status
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			status, err := CheckDependency(tt.dep)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if status == nil {
				t.Error("Expected status to be non-nil")
			}
		})
	}
}

func TestCheckAllDependencies(t *testing.T) {
	statuses, err := CheckAllDependencies()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(statuses) == 0 {
		t.Error("Expected at least some dependencies to be checked")
	}

	// Verify all required tools are checked
	expectedTools := []string{"wheresmyprompt", "files2prompt", "llm"}
	foundTools := make(map[string]bool)

	for _, status := range statuses {
		foundTools[status.Name] = true
	}

	for _, expectedTool := range expectedTools {
		if !foundTools[expectedTool] {
			t.Errorf("Expected tool %s to be checked", expectedTool)
		}
	}
}

func TestCheckVersion(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		minVersion  string
		expectValid bool
		expectError bool
	}{
		{
			name:        "Valid version check with echo",
			command:     "echo 2.0.0",
			minVersion:  "1.0.0",
			expectValid: true,
		},
		{
			name:        "Invalid command",
			command:     "nonexistent-command-xyz123 --version",
			minVersion:  "1.0.0",
			expectError: true,
		},
		{
			name:        "Empty version",
			command:     "echo",
			minVersion:  "1.0.0",
			expectValid: false,
			expectError: true, // Echo with no args produces no version, should error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, version, err := CheckVersion(tt.command, tt.minVersion)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError {
				return
			}

			if valid != tt.expectValid {
				t.Errorf("Expected valid=%t, got valid=%t (version: %s)", tt.expectValid, valid, version)
			}
		})
	}
}

func TestCheckLLMPlugins(t *testing.T) {
	plugins, err := CheckLLMPlugins()

	// This might error if llm is not installed, which is okay in tests
	if err != nil {
		t.Logf("LLM plugin check failed (expected if llm not installed): %v", err)
		return
	}

	// If no error, we should have plugin statuses
	if plugins == nil {
		t.Error("Expected plugin statuses to be non-nil when no error occurred")
	}
}

func TestRequiredDependencies(t *testing.T) {
	deps := RequiredDependencies()

	if len(deps) == 0 {
		t.Error("Expected at least some required dependencies")
	}

	// Check that all expected tools are present
	expectedTools := []string{"wheresmyprompt", "files2prompt", "llm"}
	foundTools := make(map[string]bool)

	for _, dep := range deps {
		foundTools[dep.Name] = true

		// Validate dependency structure
		if dep.Name == "" {
			t.Error("Dependency name should not be empty")
		}
		if dep.Command == "" {
			t.Error("Dependency command should not be empty")
		}
		if dep.CheckCommand == "" {
			t.Error("Dependency check command should not be empty")
		}
	}

	for _, expectedTool := range expectedTools {
		if !foundTools[expectedTool] {
			t.Errorf("Expected tool %s to be in required dependencies", expectedTool)
		}
	}
}

// Helper function to check if a command exists in PATH
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// Integration test that requires actual tools
func TestIntegrationCheckRealDependencies(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Only run if we have some common commands available
	if !commandExists("echo") {
		t.Skip("Skipping integration test - echo command not available")
	}

	// Test with a real dependency that might be installed
	dep := Dependency{
		Name:         "echo",
		Command:      "echo",
		MinVersion:   "1.0.0",
		CheckCommand: "echo test",
	}

	status, err := CheckDependency(dep)
	if err != nil {
		t.Errorf("Unexpected error checking real dependency: %v", err)
	}

	if status == nil {
		t.Error("Expected status to be non-nil")
		return
	}

	if status.Name != "echo" {
		t.Errorf("Expected status name to be 'echo', got %s", status.Name)
	}
}

func TestDependencyTypes(t *testing.T) {
	// Test the type definitions work correctly
	dep := Dependency{
		Name:         "test",
		Command:      "test-cmd",
		MinVersion:   "1.0.0",
		CheckCommand: "test-cmd --version",
		InstallURL:   "https://example.com",
		Plugins:      []string{"plugin1", "plugin2"},
	}

	if dep.Name != "test" {
		t.Errorf("Expected Name to be 'test', got %s", dep.Name)
	}

	if len(dep.Plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(dep.Plugins))
	}

	// Test DependencyStatus
	status := DependencyStatus{
		Name:      "test",
		Installed: true,
		Version:   "1.0.0",
		Valid:     true,
		Plugins: []PluginStatus{
			{Name: "plugin1", Installed: true},
		},
	}

	if !status.Installed {
		t.Error("Expected status to be installed")
	}

	if len(status.Plugins) != 1 {
		t.Errorf("Expected 1 plugin status, got %d", len(status.Plugins))
	}
}
