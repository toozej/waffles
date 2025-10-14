package repo

import (
	"path/filepath"
	"strings"
)

// LanguagePatterns defines file patterns for each supported language
type LanguagePatterns struct {
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`
}

// languagePatterns defines the default patterns for each language
var languagePatterns = map[Language]LanguagePatterns{
	LanguageGo: {
		Include: []string{
			"*.go",
			"go.mod",
			"go.sum",
		},
		Exclude: []string{
			"*_test.go",
			"pkg/version/*",
			"pkg/man/*",
			"vendor/*",
			".git/*",
			"*.pb.go",
			"*_gen.go",
			"*_generated.go",
		},
	},
	LanguagePython: {
		Include: []string{
			"*.py",
			"requirements.txt",
			"pyproject.toml",
			"setup.py",
			"setup.cfg",
			"Pipfile",
		},
		Exclude: []string{
			"*test*.py",
			"__init__.py",
			"__pycache__/*",
			".pytest_cache/*",
			"*.pyc",
			"*.pyo",
			"*.pyd",
			".git/*",
			"venv/*",
			"env/*",
			".env/*",
			".venv/*",
			"build/*",
			"dist/*",
			"*.egg-info/*",
		},
	},
}

// GetPatternsForLanguage returns the default include and exclude patterns for a language
func GetPatternsForLanguage(lang Language) (include, exclude []string) {
	if patterns, exists := languagePatterns[lang]; exists {
		return patterns.Include, patterns.Exclude
	}
	// Return generic patterns for unknown languages
	return []string{"*"}, []string{".git/*", "node_modules/*", "*.log"}
}

// ApplyPatterns filters files based on include and exclude patterns
func ApplyPatterns(files []string, include, exclude []string) []string {
	if len(include) == 0 && len(exclude) == 0 {
		return files
	}

	var result []string
	for _, file := range files {
		if shouldIncludeFile(file, include, exclude) {
			result = append(result, file)
		}
	}
	return result
}

// shouldIncludeFile determines if a file should be included based on patterns
func shouldIncludeFile(file string, include, exclude []string) bool {
	// Normalize path separators
	file = filepath.ToSlash(file)

	// Check exclude patterns first (they take precedence)
	for _, pattern := range exclude {
		if matchPattern(file, pattern) {
			return false
		}
	}

	// If no include patterns, include by default (unless excluded)
	if len(include) == 0 {
		return true
	}

	// Check include patterns
	for _, pattern := range include {
		if matchPattern(file, pattern) {
			return true
		}
	}

	return false
}

// matchPattern checks if a file matches a glob-style pattern
func matchPattern(file, pattern string) bool {
	// Normalize path separators
	file = filepath.ToSlash(file)
	pattern = filepath.ToSlash(pattern)

	// Handle directory patterns (ending with /*)
	if strings.HasSuffix(pattern, "/*") {
		dirPattern := strings.TrimSuffix(pattern, "/*")
		if strings.HasPrefix(file, dirPattern+"/") {
			return true
		}
		// Also check if file is exactly in the directory
		if filepath.Dir(file) == dirPattern {
			return true
		}
	}

	// Handle exact directory match
	if strings.HasSuffix(pattern, "/") {
		dirPattern := strings.TrimSuffix(pattern, "/")
		if strings.HasPrefix(file, dirPattern+"/") {
			return true
		}
	}

	// Use filepath.Match for glob patterns
	matched, err := filepath.Match(pattern, filepath.Base(file))
	if err != nil {
		// If pattern matching fails, fall back to string contains
		return strings.Contains(file, strings.Trim(pattern, "*"))
	}

	if matched {
		return true
	}

	// Also check against the full path for patterns with directories
	if strings.Contains(pattern, "/") {
		matched, err := filepath.Match(pattern, file)
		if err == nil && matched {
			return true
		}
		// Check if pattern matches any parent directory
		parts := strings.Split(file, "/")
		for i := 1; i <= len(parts); i++ {
			subPath := strings.Join(parts[:i], "/")
			if matched, err := filepath.Match(pattern, subPath); err == nil && matched {
				return true
			}
		}
	}

	return false
}

// MergePatterns combines multiple pattern sets
func MergePatterns(patternSets ...LanguagePatterns) LanguagePatterns {
	var merged LanguagePatterns

	for _, patterns := range patternSets {
		merged.Include = append(merged.Include, patterns.Include...)
		merged.Exclude = append(merged.Exclude, patterns.Exclude...)
	}

	// Remove duplicates
	merged.Include = removeDuplicates(merged.Include)
	merged.Exclude = removeDuplicates(merged.Exclude)

	return merged
}

// removeDuplicates removes duplicate strings from a slice
func removeDuplicates(strings []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, str := range strings {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}

	return result
}

// ParseGitignorePatterns parses .gitignore patterns into exclude patterns
func ParseGitignorePatterns(gitignoreContent string) []string {
	var patterns []string
	var negationPatterns []string
	lines := strings.Split(gitignoreContent, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle negation patterns (!)
		if strings.HasPrefix(line, "!") {
			negPattern := strings.TrimPrefix(line, "!")
			pattern := convertGitignorePattern(negPattern)
			if pattern != "" {
				negationPatterns = append(negationPatterns, pattern)
			}
			continue
		}

		// Convert gitignore pattern to filepath pattern
		pattern := convertGitignorePattern(line)
		if pattern != "" {
			patterns = append(patterns, pattern)
		}
	}

	// Remove negated patterns from the exclude list
	var finalPatterns []string
	for _, pattern := range patterns {
		isNegated := false
		for _, negPattern := range negationPatterns {
			if pattern == negPattern {
				isNegated = true
				break
			}
		}
		if !isNegated {
			finalPatterns = append(finalPatterns, pattern)
		}
	}

	return finalPatterns
}

// convertGitignorePattern converts a .gitignore pattern to a filepath pattern
func convertGitignorePattern(gitignorePattern string) string {
	pattern := gitignorePattern

	// Handle directory patterns (ending with /)
	if strings.HasSuffix(pattern, "/") {
		dirPattern := strings.TrimSuffix(pattern, "/")
		// Handle patterns starting with / (root-relative)
		dirPattern = strings.TrimPrefix(dirPattern, "/")
		return dirPattern + "/*"
	}

	// Handle patterns starting with / (root-relative)
	if strings.HasPrefix(pattern, "/") {
		return strings.TrimPrefix(pattern, "/")
	}

	// If pattern contains no path separator, it matches files in any directory
	if !strings.Contains(pattern, "/") {
		return pattern
	}

	return pattern
}

// ValidatePatterns checks if patterns are valid glob patterns
func ValidatePatterns(patterns []string) error {
	for _, pattern := range patterns {
		// Test the pattern with filepath.Match
		_, err := filepath.Match(pattern, "test")
		if err != nil {
			return err
		}
	}
	return nil
}

// ExpandPatterns expands pattern aliases to actual patterns
func ExpandPatterns(patterns []string, lang Language) []string {
	var expanded []string

	aliases := map[string][]string{
		"source": getSourcePatterns(lang),
		"config": getConfigPatterns(lang),
		"docs":   getDocPatterns(),
		"tests":  getTestPatterns(lang),
	}

	for _, pattern := range patterns {
		if expandedPatterns, exists := aliases[pattern]; exists {
			expanded = append(expanded, expandedPatterns...)
		} else {
			expanded = append(expanded, pattern)
		}
	}

	return expanded
}

// getSourcePatterns returns source file patterns for a language
func getSourcePatterns(lang Language) []string {
	patterns, _ := GetPatternsForLanguage(lang)
	var sourcePatterns []string

	for _, pattern := range patterns {
		// Only include actual source file patterns, not config files
		if strings.Contains(pattern, ".") && !isConfigFile(pattern) {
			sourcePatterns = append(sourcePatterns, pattern)
		}
	}

	return sourcePatterns
}

// getConfigPatterns returns configuration file patterns for a language
func getConfigPatterns(lang Language) []string {
	switch lang {
	case LanguageGo:
		return []string{"go.mod", "go.sum", ".goreleaser.yml", "Makefile"}
	case LanguagePython:
		return []string{"requirements.txt", "pyproject.toml", "setup.py", "setup.cfg", "Pipfile", "Pipfile.lock"}
	default:
		return []string{"*.json", "*.yaml", "*.yml", "*.toml", "*.ini"}
	}
}

// getDocPatterns returns documentation file patterns
func getDocPatterns() []string {
	return []string{"*.md", "*.rst", "*.txt", "docs/*", "*.adoc"}
}

// getTestPatterns returns test file patterns for a language
func getTestPatterns(lang Language) []string {
	switch lang {
	case LanguageGo:
		return []string{"*_test.go"}
	case LanguagePython:
		return []string{"*test*.py", "test_*.py", "tests/*"}
	default:
		return []string{"*test*", "test/*", "tests/*"}
	}
}

// isConfigFile checks if a pattern represents a configuration file
func isConfigFile(pattern string) bool {
	configExtensions := []string{".mod", ".sum", ".txt", ".toml", ".cfg", ".yml", ".yaml", ".json", ".ini"}

	for _, ext := range configExtensions {
		if strings.HasSuffix(pattern, ext) {
			return true
		}
	}

	configFiles := []string{"Makefile", "Dockerfile", "Pipfile"}
	for _, file := range configFiles {
		if strings.Contains(pattern, file) {
			return true
		}
	}

	return false
}
