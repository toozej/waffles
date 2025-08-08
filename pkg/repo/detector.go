package repo

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// DetectLanguage analyzes a directory to determine the primary programming language
func DetectLanguage(path string) (Language, error) {
	result, err := DetectLanguageWithDetails(path)
	if err != nil {
		return LanguageUnknown, err
	}
	return result.Language, nil
}

// DetectLanguageWithDetails analyzes a directory and returns detailed detection results
func DetectLanguageWithDetails(path string) (*LanguageDetectionResult, error) {
	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("path does not exist: %s", path)
	}

	indicators := []DetectionIndicator{}
	languageScores := make(map[Language]float64)

	// Initialize all supported languages with zero scores
	for _, lang := range SupportedLanguages() {
		languageScores[lang] = 0
	}

	// Walk the directory tree (but not too deep to avoid performance issues)
	err := filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Continue on errors
		}

		// Skip hidden directories and common ignore patterns
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") ||
				name == "node_modules" ||
				name == "vendor" ||
				name == "__pycache__" {
				return filepath.SkipDir
			}
		}

		// Limit depth to avoid deep recursion
		relPath, _ := filepath.Rel(path, filePath)
		if strings.Count(relPath, string(os.PathSeparator)) > 3 {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() {
			// Analyze files for language indicators
			analyzeFileForLanguage(filePath, path, &indicators, languageScores)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	// Determine the language with the highest score
	var detectedLang Language = LanguageUnknown
	var maxScore float64 = 0

	for lang, score := range languageScores {
		if score > maxScore {
			maxScore = score
			detectedLang = lang
		}
	}

	// Calculate confidence based on score
	confidence := calculateConfidence(maxScore, languageScores)

	return &LanguageDetectionResult{
		Language:   detectedLang,
		Confidence: confidence,
		Indicators: indicators,
	}, nil
}

// analyzeFileForLanguage examines a single file for language indicators
func analyzeFileForLanguage(filePath, rootPath string, indicators *[]DetectionIndicator, scores map[Language]float64) {
	fileName := filepath.Base(filePath)
	relPath, _ := filepath.Rel(rootPath, filePath)

	// Go language detection
	switch fileName {
	case "go.mod":
		addIndicator(indicators, IndicatorManifestFile, relPath, 50.0, "Go module file")
		scores[LanguageGo] += 50.0
	case "go.sum":
		addIndicator(indicators, IndicatorManifestFile, relPath, 20.0, "Go checksums file")
		scores[LanguageGo] += 20.0
	case "main.go":
		addIndicator(indicators, IndicatorSourceFile, relPath, 15.0, "Go main file")
		scores[LanguageGo] += 15.0
	default:
		if strings.HasSuffix(fileName, ".go") {
			addIndicator(indicators, IndicatorSourceFile, relPath, 10.0, "Go source file")
			scores[LanguageGo] += 10.0
		}
	}

	// Python language detection
	switch fileName {
	case "requirements.txt":
		addIndicator(indicators, IndicatorManifestFile, relPath, 40.0, "Python requirements file")
		scores[LanguagePython] += 40.0
	case "pyproject.toml":
		addIndicator(indicators, IndicatorManifestFile, relPath, 45.0, "Python project file")
		scores[LanguagePython] += 45.0
	case "setup.py":
		addIndicator(indicators, IndicatorManifestFile, relPath, 35.0, "Python setup file")
		scores[LanguagePython] += 35.0
	case "Pipfile":
		addIndicator(indicators, IndicatorManifestFile, relPath, 30.0, "Python Pipfile")
		scores[LanguagePython] += 30.0
	case "__init__.py":
		addIndicator(indicators, IndicatorSourceFile, relPath, 5.0, "Python package init file")
		scores[LanguagePython] += 5.0
	default:
		if strings.HasSuffix(fileName, ".py") {
			addIndicator(indicators, IndicatorSourceFile, relPath, 10.0, "Python source file")
			scores[LanguagePython] += 10.0
		}
	}

	// Directory-based detection
	if strings.Contains(relPath, "cmd/") && strings.HasSuffix(fileName, ".go") {
		addIndicator(indicators, IndicatorDirectory, "cmd/", 5.0, "Go command directory structure")
		scores[LanguageGo] += 5.0
	}
	if strings.Contains(relPath, "pkg/") && strings.HasSuffix(fileName, ".go") {
		addIndicator(indicators, IndicatorDirectory, "pkg/", 5.0, "Go package directory structure")
		scores[LanguageGo] += 5.0
	}
	if strings.Contains(relPath, "__pycache__") {
		addIndicator(indicators, IndicatorDirectory, "__pycache__/", 3.0, "Python cache directory")
		scores[LanguagePython] += 3.0
	}
}

// addIndicator adds a detection indicator to the list
func addIndicator(indicators *[]DetectionIndicator, iType IndicatorType, value string, weight float64, description string) {
	*indicators = append(*indicators, DetectionIndicator{
		Type:        iType,
		Value:       value,
		Weight:      weight,
		Description: description,
	})
}

// calculateConfidence calculates confidence based on scores
func calculateConfidence(maxScore float64, scores map[Language]float64) float64 {
	if maxScore == 0 {
		return 0.0
	}

	// Calculate total score
	var totalScore float64
	for _, score := range scores {
		totalScore += score
	}

	if totalScore == 0 {
		return 0.0
	}

	// Confidence is the percentage of the max score
	confidence := maxScore / totalScore
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// AnalyzeRepository analyzes a repository and returns comprehensive information
func AnalyzeRepository(path string, overrides *RepositoryOverrides) (*RepositoryInfo, error) {
	if overrides == nil {
		overrides = DefaultOverrides()
	}

	// Detect language (or use override)
	var language Language
	if overrides.Language != nil {
		language = *overrides.Language
	} else {
		var err error
		language, err = DetectLanguage(path)
		if err != nil {
			return nil, fmt.Errorf("failed to detect language: %w", err)
		}
	}

	// Get base patterns for the detected language
	includePatterns, excludePatterns := GetPatternsForLanguage(language)

	// Apply overrides
	if len(overrides.IncludePatterns) > 0 {
		includePatterns = append(includePatterns, overrides.IncludePatterns...)
	}
	if len(overrides.ExcludePatterns) > 0 {
		excludePatterns = append(excludePatterns, overrides.ExcludePatterns...)
	}

	// Load gitignore rules if not ignored
	var gitignoreRules []string
	if !overrides.IgnoreGitignore {
		gitignoreRules = loadGitignoreRules(path)
		excludePatterns = append(excludePatterns, gitignoreRules...)
	}

	// Scan files
	detectedFiles, err := ScanFiles(path, includePatterns, excludePatterns, overrides)
	if err != nil {
		return nil, fmt.Errorf("failed to scan files: %w", err)
	}

	return &RepositoryInfo{
		Language:        language,
		RootPath:        path,
		IncludePatterns: includePatterns,
		ExcludePatterns: excludePatterns,
		DetectedFiles:   detectedFiles,
		GitIgnoreRules:  gitignoreRules,
	}, nil
}

// ScanFiles scans a directory for files matching the given patterns
func ScanFiles(path string, includePatterns, excludePatterns []string, overrides *RepositoryOverrides) ([]FileInfo, error) {
	var files []FileInfo
	maxFiles := 1000
	maxFileSize := int64(1024 * 1024) // 1MB

	if overrides != nil {
		if overrides.MaxFiles > 0 {
			maxFiles = overrides.MaxFiles
		}
		if overrides.MaxFileSize > 0 {
			maxFileSize = overrides.MaxFileSize
		}
	}

	fileCount := 0
	err := filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Continue on errors
		}

		// Skip directories
		if d.IsDir() {
			// Skip hidden and common ignore directories
			name := d.Name()
			if strings.HasPrefix(name, ".") ||
				name == "node_modules" ||
				name == "vendor" ||
				name == "__pycache__" ||
				name == "build" ||
				name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file count limit
		if fileCount >= maxFiles {
			return fmt.Errorf("too many files (limit: %d)", maxFiles)
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			return nil // Continue on errors
		}

		// Check file size limit
		if info.Size() > maxFileSize {
			relPath, _ := filepath.Rel(path, filePath)
			files = append(files, FileInfo{
				Path:     relPath,
				Size:     info.Size(),
				Included: false,
				Reason:   fmt.Sprintf("File too large (%d bytes > %d bytes)", info.Size(), maxFileSize),
			})
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(path, filePath)
		if err != nil {
			return nil // Continue on errors
		}

		// Check patterns
		included, reason := shouldIncludeFileWithReason(relPath, includePatterns, excludePatterns)

		files = append(files, FileInfo{
			Path:     relPath,
			Size:     info.Size(),
			Included: included,
			Reason:   reason,
		})

		if included {
			fileCount++
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// shouldIncludeFileWithReason determines if a file should be included and provides a reason
func shouldIncludeFileWithReason(file string, include, exclude []string) (bool, string) {
	// Normalize path separators
	file = filepath.ToSlash(file)

	// Check exclude patterns first (they take precedence)
	for _, pattern := range exclude {
		if matchPattern(file, pattern) {
			return false, fmt.Sprintf("Excluded by pattern: %s", pattern)
		}
	}

	// If no include patterns, include by default (unless excluded)
	if len(include) == 0 {
		return true, "No include patterns specified"
	}

	// Check include patterns
	for _, pattern := range include {
		if matchPattern(file, pattern) {
			return true, fmt.Sprintf("Included by pattern: %s", pattern)
		}
	}

	return false, "No matching include pattern"
}

// ApplyGitignore applies .gitignore rules to filter files
func ApplyGitignore(files []FileInfo, path string) []FileInfo {
	gitignoreRules, negationRules := loadGitignoreRulesWithNegations(path)
	if len(gitignoreRules) == 0 && len(negationRules) == 0 {
		return files
	}

	var result []FileInfo
	for _, file := range files {
		excluded := false
		var excludeReason string

		// Convert to relative path for pattern matching
		var checkPath string
		if filepath.IsAbs(file.Path) {
			relPath, err := filepath.Rel(path, file.Path)
			if err != nil {
				// If we can't get relative path, use the original path
				checkPath = file.Path
			} else {
				checkPath = filepath.ToSlash(relPath)
			}
		} else {
			checkPath = filepath.ToSlash(file.Path)
		}

		// Check exclude rules first
		for _, rule := range gitignoreRules {
			if matchPattern(checkPath, rule) {
				excluded = true
				excludeReason = fmt.Sprintf("Excluded by .gitignore rule: %s", rule)
				break
			}
		}

		// Check negation rules (they override exclusions)
		if excluded {
			for _, rule := range negationRules {
				if matchPattern(checkPath, rule) {
					excluded = false
					excludeReason = ""
					break
				}
			}
		}

		if excluded {
			file.Included = false
			if file.Reason == "" || strings.Contains(file.Reason, "Included") {
				file.Reason = excludeReason
			}
		}

		result = append(result, file)
	}

	return result
}

// loadGitignoreRules loads patterns from .gitignore files
func loadGitignoreRules(path string) []string {
	var allRules []string

	// Check for .gitignore in the root directory
	gitignorePath := filepath.Join(path, ".gitignore")
	// #nosec G304 -- Reading .gitignore from trusted directory
	if content, err := os.ReadFile(gitignorePath); err == nil {
		rules := ParseGitignorePatterns(string(content))
		allRules = append(allRules, rules...)
	}

	// Could also check parent directories and global gitignore, but keeping it simple for now

	return allRules
}

// loadGitignoreRulesWithNegations loads patterns from .gitignore files, separating regular rules and negations
func loadGitignoreRulesWithNegations(path string) (excludeRules, negationRules []string) {
	// Check for .gitignore in the root directory
	gitignorePath := filepath.Join(path, ".gitignore")
	content, err := os.ReadFile(gitignorePath) // #nosec G304 -- Reading .gitignore from trusted directory
	if err != nil {
		return nil, nil
	}

	lines := strings.Split(string(content), "\n")

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
				negationRules = append(negationRules, pattern)
			}
			continue
		}

		// Convert gitignore pattern to filepath pattern
		pattern := convertGitignorePattern(line)
		if pattern != "" {
			excludeRules = append(excludeRules, pattern)
		}
	}

	return excludeRules, negationRules
}
