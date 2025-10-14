package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name            string
		envVars         map[string]string
		envFileContent  string
		expectError     bool
		expectedModel   string
		expectedVerbose bool
	}{
		{
			name:            "Default configuration",
			expectedModel:   "claude-3-sonnet",
			expectedVerbose: false,
		},
		{
			name: "Environment variables override defaults",
			envVars: map[string]string{
				"WAFFLES_DEFAULT_MODEL": "gpt-4",
				"WAFFLES_VERBOSE":       "true",
			},
			expectedModel:   "gpt-4",
			expectedVerbose: true,
		},
		{
			name:            "Env file loading",
			envFileContent:  "WAFFLES_DEFAULT_MODEL=gemini-pro\nWAFFLES_VERBOSE=true",
			expectedModel:   "gemini-pro",
			expectedVerbose: true,
		},
		{
			name: "Environment variables override env file",
			envVars: map[string]string{
				"WAFFLES_DEFAULT_MODEL": "claude-3-opus",
			},
			envFileContent:  "WAFFLES_DEFAULT_MODEL=gpt-3.5-turbo\nWAFFLES_VERBOSE=true",
			expectedModel:   "claude-3-opus", // env var should win
			expectedVerbose: true,            // from file
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
			cleanupEnv := []string{
				"WAFFLES_DEFAULT_MODEL",
				"WAFFLES_VERBOSE",
				"WAFFLES_DEFAULT_PROVIDER",
			}
			for _, key := range cleanupEnv {
				os.Unsetenv(key)
			}

			// Set up environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Create temporary directory for test
			tempDir := t.TempDir()
			oldDir, _ := os.Getwd()
			_ = os.Chdir(tempDir)
			defer func() { _ = os.Chdir(oldDir) }()

			// Write .env file if provided
			if tt.envFileContent != "" {
				err := os.WriteFile(".waffles.env", []byte(tt.envFileContent), 0644)
				if err != nil {
					t.Fatalf("Failed to write .env file: %v", err)
				}
			}

			// Test LoadConfig
			cfg, err := LoadConfig()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if tt.expectError {
				return
			}

			// Verify results
			if cfg.DefaultModel != tt.expectedModel {
				t.Errorf("Expected DefaultModel %q, got %q", tt.expectedModel, cfg.DefaultModel)
			}
			if cfg.Verbose != tt.expectedVerbose {
				t.Errorf("Expected Verbose %t, got %t", tt.expectedVerbose, cfg.Verbose)
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, "test.env")

	// Write test env file
	envContent := "WAFFLES_DEFAULT_MODEL=test-model\nWAFFLES_VERBOSE=true"
	err := os.WriteFile(envFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test env file: %v", err)
	}

	var cfg Config
	err = LoadFromFile(&cfg, envFile)
	if err != nil {
		t.Errorf("Unexpected error loading from file: %v", err)
	}

	// Test non-existent file (should not error)
	err = LoadFromFile(&cfg, "non-existent.env")
	if err != nil {
		t.Errorf("Expected no error for non-existent file, got: %v", err)
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("WAFFLES_DEFAULT_MODEL", "env-test-model")
	os.Setenv("WAFFLES_VERBOSE", "true")
	defer func() {
		os.Unsetenv("WAFFLES_DEFAULT_MODEL")
		os.Unsetenv("WAFFLES_VERBOSE")
	}()

	var cfg Config
	err := LoadFromEnv(&cfg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if cfg.DefaultModel != "env-test-model" {
		t.Errorf("Expected DefaultModel 'env-test-model', got %q", cfg.DefaultModel)
	}
	if !cfg.Verbose {
		t.Error("Expected Verbose to be true")
	}
}
