package deps

// Dependency represents an external tool dependency
type Dependency struct {
	Name         string   // Display name
	Command      string   // Command to check in PATH
	MinVersion   string   // Minimum required version
	CheckCommand string   // Command to check version
	InstallURL   string   // URL with installation instructions
	Plugins      []string // Required plugins (for llm CLI)
}

// DependencyStatus represents the status of a dependency
type DependencyStatus struct {
	Name      string         `json:"name"`
	Installed bool           `json:"installed"`
	Version   string         `json:"version"`
	Valid     bool           `json:"valid"`
	Error     string         `json:"error,omitempty"`
	Plugins   []PluginStatus `json:"plugins,omitempty"`
}

// PluginStatus represents the status of an LLM plugin
type PluginStatus struct {
	Name      string `json:"name"`
	Installed bool   `json:"installed"`
	Error     string `json:"error,omitempty"`
}

// SystemInfo contains information about the host system
type SystemInfo struct {
	OS           string   `json:"os"`
	Architecture string   `json:"architecture"`
	Shell        string   `json:"shell"`
	PathDirs     []string `json:"path_dirs"`
}

// InstallationResult represents the result of an installation attempt
type InstallationResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// RequiredDependencies returns the list of all required dependencies
func RequiredDependencies() []Dependency {
	return []Dependency{
		{
			Name:         "wheresmyprompt",
			Command:      "wheresmyprompt",
			MinVersion:   "0.1.0",
			CheckCommand: "wheresmyprompt version",
			InstallURL:   "https://github.com/toozej/wheresmyprompt#installation",
			Plugins:      []string{},
		},
		{
			Name:         "files2prompt",
			Command:      "files2prompt",
			MinVersion:   "0.1.0",
			CheckCommand: "files2prompt version",
			InstallURL:   "https://github.com/toozej/files2prompt#installation",
			Plugins:      []string{},
		},
		{
			Name:         "llm",
			Command:      "llm",
			MinVersion:   "0.10.0",
			CheckCommand: "llm --version",
			InstallURL:   "https://llm.datasette.io/en/stable/setup.html",
			Plugins: []string{
				"llm-anthropic",
				"llm-ollama",
				"llm-gemini",
				"llm-jq",
				"llm-fragments-github",
				"llm-fragments-go",
				"llm-commit",
			},
		},
	}
}
