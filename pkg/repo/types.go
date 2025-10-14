package repo

// Language represents a detected programming language
type Language string

const (
	LanguageGo      Language = "go"
	LanguagePython  Language = "python"
	LanguageUnknown Language = "unknown"
)

// String returns the string representation of the language
func (l Language) String() string {
	return string(l)
}

// RepositoryInfo contains information about the analyzed repository
type RepositoryInfo struct {
	Language        Language   `json:"language"`
	RootPath        string     `json:"root_path"`
	IncludePatterns []string   `json:"include_patterns"`
	ExcludePatterns []string   `json:"exclude_patterns"`
	DetectedFiles   []FileInfo `json:"detected_files"`
	GitIgnoreRules  []string   `json:"gitignore_rules,omitempty"`
}

// FileInfo contains information about a single file
type FileInfo struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Included bool   `json:"included"`
	Reason   string `json:"reason,omitempty"`
}

// RepositoryOverrides contains user-specified overrides for repository analysis
type RepositoryOverrides struct {
	Language        *Language `json:"language,omitempty"`
	IncludePatterns []string  `json:"include_patterns,omitempty"`
	ExcludePatterns []string  `json:"exclude_patterns,omitempty"`
	IgnoreGitignore bool      `json:"ignore_gitignore"`
	MaxFiles        int       `json:"max_files,omitempty"`
	MaxFileSize     int64     `json:"max_file_size,omitempty"`
}

// LanguageDetectionResult contains the result of language detection
type LanguageDetectionResult struct {
	Language   Language             `json:"language"`
	Confidence float64              `json:"confidence"`
	Indicators []DetectionIndicator `json:"indicators"`
}

// DetectionIndicator represents evidence for a particular language
type DetectionIndicator struct {
	Type        IndicatorType `json:"type"`
	Value       string        `json:"value"`
	Weight      float64       `json:"weight"`
	Description string        `json:"description"`
}

// IndicatorType represents the type of language detection indicator
type IndicatorType string

const (
	IndicatorManifestFile IndicatorType = "manifest_file"
	IndicatorSourceFile   IndicatorType = "source_file"
	IndicatorConfigFile   IndicatorType = "config_file"
	IndicatorDirectory    IndicatorType = "directory"
	IndicatorFileCount    IndicatorType = "file_count"
)

// SupportedLanguages returns a list of all supported languages
func SupportedLanguages() []Language {
	return []Language{
		LanguageGo,
		LanguagePython,
	}
}

// IsSupported checks if a language is supported
func (l Language) IsSupported() bool {
	for _, supported := range SupportedLanguages() {
		if l == supported {
			return true
		}
	}
	return false
}

// DefaultOverrides returns default repository analysis overrides
func DefaultOverrides() *RepositoryOverrides {
	return &RepositoryOverrides{
		Language:        nil,
		IncludePatterns: []string{},
		ExcludePatterns: []string{},
		IgnoreGitignore: false,
		MaxFiles:        1000,
		MaxFileSize:     1024 * 1024, // 1MB
	}
}
