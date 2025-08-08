package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string]string // filename -> content
		expected    Language
		expectError bool
	}{
		{
			name: "Go project with go.mod",
			files: map[string]string{
				"go.mod":  "module test\n",
				"main.go": "package main\n",
			},
			expected: LanguageGo,
		},
		{
			name: "Go project with .go files only",
			files: map[string]string{
				"main.go":    "package main\n",
				"helpers.go": "package main\n",
			},
			expected: LanguageGo,
		},
		{
			name: "Python project with requirements.txt",
			files: map[string]string{
				"requirements.txt": "requests==2.25.1\n",
				"main.py":          "import requests\n",
			},
			expected: LanguagePython,
		},
		{
			name: "Python project with .py files only",
			files: map[string]string{
				"main.py":    "import os\n",
				"helpers.py": "def helper():\n    pass\n",
			},
			expected: LanguagePython,
		},
		{
			name: "Mixed project with Go preference",
			files: map[string]string{
				"go.mod":  "module test\n",
				"main.go": "package main\n",
				"main.py": "import os\n",
			},
			expected: LanguageGo,
		},
		{
			name: "Empty directory",
			files: map[string]string{
				".gitkeep": "",
			},
			expected: LanguageUnknown,
		},
		{
			name: "Text files only",
			files: map[string]string{
				"README.md": "# Test\n",
				"data.txt":  "some data\n",
			},
			expected: LanguageUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir := t.TempDir()

			// Create test files
			for filename, content := range tt.files {
				filepath := filepath.Join(tempDir, filename)
				if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create test file %s: %v", filename, err)
				}
			}

			// Test language detection
			detected, err := DetectLanguage(tempDir)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError {
				return
			}

			if detected != tt.expected {
				t.Errorf("Expected language %s, got %s", tt.expected, detected)
			}
		})
	}
}

func TestAnalyzeRepository(t *testing.T) {
	// Create a test repository
	tempDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"go.mod":           "module github.com/test/repo\n",
		"main.go":          "package main\n\nfunc main() {}\n",
		"internal/app.go":  "package internal\n",
		"cmd/tool/main.go": "package main\n",
		"README.md":        "# Test Repository\n",
		"vendor/dep.go":    "package vendor\n",
		".git/config":      "[core]\n",
		"testdata/test.go": "package testdata\n",
		"docs/guide.md":    "# Guide\n",
	}

	for filename, content := range testFiles {
		fullPath := filepath.Join(tempDir, filename)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Test repository analysis
	info, err := AnalyzeRepository(tempDir, nil)
	if err != nil {
		t.Fatalf("Unexpected error analyzing repository: %v", err)
	}

	// Verify results
	if info.Language != LanguageGo {
		t.Errorf("Expected language %s, got %s", LanguageGo, info.Language)
	}

	if info.RootPath != tempDir {
		t.Errorf("Expected root path %s, got %s", tempDir, info.RootPath)
	}

	if len(info.DetectedFiles) == 0 {
		t.Error("Expected some detected files")
	}

	// Check that Go files are included
	goFileFound := false
	for _, file := range info.DetectedFiles {
		if filepath.Ext(file.Path) == ".go" && file.Included {
			goFileFound = true
			break
		}
	}
	if !goFileFound {
		t.Error("Expected at least one included Go file")
	}
}

func TestAnalyzeRepositoryWithOverrides(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"main.go":   "package main\n",
		"helper.go": "package main\n",
		"test.py":   "import os\n",
		"README.md": "# Test\n",
		"data.json": `{"test": true}`,
	}

	for filename, content := range testFiles {
		fullPath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Test with overrides
	pythonLang := LanguagePython
	overrides := &RepositoryOverrides{
		Language:        &pythonLang,
		IncludePatterns: []string{"*.py", "*.json"},
		ExcludePatterns: []string{"README.*"},
	}

	info, err := AnalyzeRepository(tempDir, overrides)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use override language
	if info.Language != LanguagePython {
		t.Errorf("Expected overridden language %s, got %s", LanguagePython, info.Language)
	}

	// Check include patterns are used
	if len(info.IncludePatterns) == 0 {
		t.Error("Expected include patterns to be set")
	}

	// Check that patterns are applied
	jsonFileIncluded := false
	readmeFileIncluded := false
	for _, file := range info.DetectedFiles {
		if filepath.Base(file.Path) == "data.json" && file.Included {
			jsonFileIncluded = true
		}
		if filepath.Base(file.Path) == "README.md" && file.Included {
			readmeFileIncluded = true
		}
	}

	if !jsonFileIncluded {
		t.Error("Expected JSON file to be included due to include patterns")
	}
	if readmeFileIncluded {
		t.Error("Expected README file to be excluded due to exclude patterns")
	}
}

func TestScanFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"main.go":          "package main\n",
		"internal/app.go":  "package internal\n",
		"vendor/deps.go":   "package vendor\n",
		"testdata/test.go": "package testdata\n",
		"README.md":        "# Test\n",
		"go.mod":           "module test\n",
		".git/config":      "[core]\n",
		"docs/example.go":  "package example\n",
	}

	for filename, content := range testFiles {
		fullPath := filepath.Join(tempDir, filename)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Test file scanning with Go patterns
	includePatterns := []string{"*.go", "go.mod", "go.sum"}
	excludePatterns := []string{}
	files, err := ScanFiles(tempDir, includePatterns, excludePatterns, nil)
	if err != nil {
		t.Fatalf("Unexpected error scanning files: %v", err)
	}

	if len(files) == 0 {
		t.Error("Expected some files to be found")
	}

	// Check that Go files are found
	goFileFound := false
	goModFound := false
	for _, file := range files {
		basename := filepath.Base(file.Path)
		if filepath.Ext(basename) == ".go" {
			goFileFound = true
		}
		if basename == "go.mod" {
			goModFound = true
		}
	}

	if !goFileFound {
		t.Error("Expected at least one .go file")
	}
	if !goModFound {
		t.Error("Expected go.mod file")
	}
}

func TestApplyGitignore(t *testing.T) {
	tempDir := t.TempDir()

	// Create .gitignore
	gitignoreContent := `# Ignore patterns
*.log
/build/
node_modules/
*.tmp
!important.tmp
`
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Create test files
	testFiles := map[string]string{
		"main.go":          "package main\n",
		"app.log":          "log entry\n",
		"build/output":     "build output\n",
		"data.tmp":         "temporary data\n",
		"important.tmp":    "important temp file\n",
		"node_modules/dep": "dependency\n",
		"src/helper.go":    "package src\n",
	}

	for filename, content := range testFiles {
		fullPath := filepath.Join(tempDir, filename)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Create initial file list
	var files []FileInfo
	for filename := range testFiles {
		files = append(files, FileInfo{
			Path:     filepath.Join(tempDir, filename),
			Included: true,
		})
	}

	// Apply gitignore
	filteredFiles := ApplyGitignore(files, tempDir)

	// Check results
	includedFiles := make(map[string]bool)
	for _, file := range filteredFiles {
		if file.Included {
			relPath, _ := filepath.Rel(tempDir, file.Path)
			includedFiles[relPath] = true
		}
	}

	// These should be included
	expectedIncluded := []string{"main.go", "important.tmp", "src/helper.go"}
	for _, expected := range expectedIncluded {
		if !includedFiles[expected] {
			t.Errorf("Expected %s to be included", expected)
		}
	}

	// These should be excluded
	expectedExcluded := []string{"app.log", "build/output", "data.tmp", "node_modules/dep"}
	for _, expected := range expectedExcluded {
		if includedFiles[expected] {
			t.Errorf("Expected %s to be excluded", expected)
		}
	}
}

func TestGetPatternsForLanguage(t *testing.T) {
	tests := []struct {
		language Language
		hasGo    bool
		hasPy    bool
	}{
		{LanguageGo, true, false},
		{LanguagePython, false, true},
		{LanguageUnknown, false, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.language), func(t *testing.T) {
			include, exclude := GetPatternsForLanguage(tt.language)

			if len(include) == 0 && tt.language != LanguageUnknown {
				t.Error("Expected include patterns for known language")
			}

			// Check for expected patterns
			includeStr := strings.Join(include, " ")
			excludeStr := strings.Join(exclude, " ")

			if tt.hasGo {
				if !strings.Contains(includeStr, "*.go") {
					t.Error("Expected Go patterns in include list")
				}
				if !strings.Contains(excludeStr, "vendor/") {
					t.Error("Expected vendor exclusion in Go patterns")
				}
			}

			if tt.hasPy {
				if !strings.Contains(includeStr, "*.py") {
					t.Error("Expected Python patterns in include list")
				}
				if !strings.Contains(excludeStr, "__pycache__/") {
					t.Error("Expected __pycache__ exclusion in Python patterns")
				}
			}

			t.Logf("Language %s - Include: %v, Exclude: %v", tt.language, include, exclude)
		})
	}
}

func TestApplyPatterns(t *testing.T) {
	files := []string{
		"main.go",
		"helper.go",
		"test.py",
		"data.json",
		"README.md",
		"vendor/dep.go",
		"__pycache__/cache.pyc",
		"node_modules/lib.js",
	}

	tests := []struct {
		name     string
		include  []string
		exclude  []string
		expected []string
	}{
		{
			name:     "Go patterns",
			include:  []string{"*.go", "go.mod"},
			exclude:  []string{"vendor/*"},
			expected: []string{"main.go", "helper.go"},
		},
		{
			name:     "Python patterns",
			include:  []string{"*.py"},
			exclude:  []string{"__pycache__/*"},
			expected: []string{"test.py"},
		},
		{
			name:     "All files",
			include:  []string{"*"},
			exclude:  []string{},
			expected: files,
		},
		{
			name:     "Specific exclusions",
			include:  []string{"*"},
			exclude:  []string{"*.md", "*.json", "node_modules/*"},
			expected: []string{"main.go", "helper.go", "test.py", "vendor/dep.go", "__pycache__/cache.pyc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyPatterns(files, tt.include, tt.exclude)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d files, got %d", len(tt.expected), len(result))
			}

			// Convert to map for easy lookup
			resultMap := make(map[string]bool)
			for _, file := range result {
				resultMap[file] = true
			}

			for _, expected := range tt.expected {
				if !resultMap[expected] {
					t.Errorf("Expected file %s to be included", expected)
				}
			}
		})
	}
}

func TestLanguageString(t *testing.T) {
	tests := []struct {
		lang     Language
		expected string
	}{
		{LanguageGo, "go"},
		{LanguagePython, "python"},
		{LanguageUnknown, "unknown"},
	}

	for _, tt := range tests {
		if string(tt.lang) != tt.expected {
			t.Errorf("Expected language string %s, got %s", tt.expected, string(tt.lang))
		}
	}
}

func TestRepositoryInfo(t *testing.T) {
	// Test RepositoryInfo struct
	info := &RepositoryInfo{
		Language:        LanguageGo,
		RootPath:        "/test/path",
		IncludePatterns: []string{"*.go"},
		ExcludePatterns: []string{"vendor/*"},
		DetectedFiles: []FileInfo{
			{Path: "main.go", Size: 100, Included: true},
			{Path: "vendor/dep.go", Size: 50, Included: false, Reason: "excluded by pattern"},
		},
	}

	if info.Language != LanguageGo {
		t.Error("Expected Language to be Go")
	}

	if len(info.DetectedFiles) != 2 {
		t.Error("Expected 2 detected files")
	}

	includedCount := 0
	for _, file := range info.DetectedFiles {
		if file.Included {
			includedCount++
		}
	}

	if includedCount != 1 {
		t.Errorf("Expected 1 included file, got %d", includedCount)
	}
}

func TestFileInfo(t *testing.T) {
	info := FileInfo{
		Path:     "/test/main.go",
		Size:     1024,
		Included: true,
		Reason:   "matches pattern",
	}

	if info.Path != "/test/main.go" {
		t.Error("Expected Path to be set correctly")
	}

	if info.Size != 1024 {
		t.Error("Expected Size to be 1024")
	}

	if !info.Included {
		t.Error("Expected Included to be true")
	}

	if info.Reason != "matches pattern" {
		t.Error("Expected Reason to be set correctly")
	}
}

// Benchmark tests
func BenchmarkDetectLanguage(b *testing.B) {
	tempDir := b.TempDir()

	// Create test files
	testFiles := map[string]string{
		"go.mod":  "module test\n",
		"main.go": "package main\n",
		"app.go":  "package main\n",
	}

	for filename, content := range testFiles {
		fullPath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DetectLanguage(tempDir)
	}
}

func BenchmarkScanFiles(b *testing.B) {
	tempDir := b.TempDir()

	// Create many test files
	for i := 0; i < 100; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("file%d.go", i))
		content := fmt.Sprintf("package file%d\n", i)
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
	}

	includePatterns := []string{"*.go"}
	excludePatterns := []string{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ScanFiles(tempDir, includePatterns, excludePatterns, nil)
	}
}
