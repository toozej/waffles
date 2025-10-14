package deps

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

// CheckDependency checks the status of a single dependency
func CheckDependency(dep Dependency) (*DependencyStatus, error) {
	status := &DependencyStatus{
		Name:      dep.Name,
		Installed: false,
		Valid:     false,
	}

	// Check if command exists in PATH
	_, err := exec.LookPath(dep.Command)
	if err != nil {
		status.Error = fmt.Sprintf("Command '%s' not found in PATH", dep.Command)
		return status, nil
	}

	status.Installed = true

	// Check version if available
	if dep.CheckCommand != "" {
		valid, version, err := CheckVersion(dep.CheckCommand, dep.MinVersion)
		if err != nil {
			status.Error = fmt.Sprintf("Failed to check version: %v", err)
			return status, nil
		}
		status.Version = version
		status.Valid = valid

		if !valid && dep.MinVersion != "" {
			status.Error = fmt.Sprintf("Version %s is below minimum required %s", version, dep.MinVersion)
		}
	} else {
		// If no version check command, assume valid if installed
		status.Valid = true
	}

	// Check plugins if this is the llm CLI
	if dep.Name == "llm" && len(dep.Plugins) > 0 {
		pluginStatuses, err := CheckLLMPlugins()
		if err != nil {
			status.Error = fmt.Sprintf("Failed to check plugins: %v", err)
		} else {
			status.Plugins = pluginStatuses
		}
	}

	return status, nil
}

// CheckAllDependencies checks the status of all required dependencies
func CheckAllDependencies() ([]DependencyStatus, error) {
	deps := RequiredDependencies()
	statuses := make([]DependencyStatus, len(deps))

	for i, dep := range deps {
		status, err := CheckDependency(dep)
		if err != nil {
			return nil, fmt.Errorf("error checking dependency %s: %w", dep.Name, err)
		}
		statuses[i] = *status
	}

	return statuses, nil
}

// CheckVersion checks if a command meets minimum version requirements
func CheckVersion(checkCommand, minVersion string) (bool, string, error) {
	if minVersion == "" {
		return true, "", nil
	}

	// Split command and args
	parts := strings.Fields(checkCommand)
	if len(parts) == 0 {
		return false, "", fmt.Errorf("empty check command")
	}

	// Validate command is safe for execution
	commandName := parts[0]
	allowedCommands := map[string]bool{
		"go":             true,
		"python":         true,
		"python3":        true,
		"pip":            true,
		"pip3":           true,
		"llm":            true,
		"files2prompt":   true,
		"wheresmyprompt": true,
		"git":            true,
		"node":           true,
		"npm":            true,
		"yarn":           true,
		"cargo":          true,
		"rustc":          true,
		"echo":           true, // Used in tests
	}

	if !allowedCommands[commandName] {
		return false, "", fmt.Errorf("command '%s' is not in the allowed list for version checking", commandName)
	}

	// Execute version command with timeout
	cmd := exec.Command(commandName, parts[1:]...) // #nosec G204 # nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
	cmd.Env = os.Environ()

	// Set a reasonable timeout
	done := make(chan error, 1)
	var output []byte
	var err error

	go func() {
		output, err = cmd.CombinedOutput()
		done <- err
	}()

	select {
	case <-time.After(10 * time.Second):
		if cmd.Process != nil {
			_ = cmd.Process.Kill() // Ignore error from Kill as process might already be dead
		}
		return false, "", fmt.Errorf("version check timed out")
	case err := <-done:
		if err != nil {
			return false, "", fmt.Errorf("version check failed: %w", err)
		}
	}

	// Extract version from output
	version := extractVersion(string(output))
	if version == "" {
		return false, string(output), fmt.Errorf("could not parse version from output")
	}

	// Handle special version strings like "local" for development builds
	if version == "local" {
		// Local/dev builds are considered valid
		return true, version, nil
	}

	// Compare versions using semver
	currentVer, err := semver.NewVersion(version)
	if err != nil {
		// If semver parsing fails, do string comparison as fallback
		return version >= minVersion, version, nil
	}

	minVer, err := semver.NewVersion(minVersion)
	if err != nil {
		return false, version, fmt.Errorf("invalid minimum version format: %s", minVersion)
	}

	return currentVer.GreaterThan(minVer) || currentVer.Equal(minVer), version, nil
}

// CheckLLMPlugins checks the status of required LLM plugins
func CheckLLMPlugins() ([]PluginStatus, error) {
	deps := RequiredDependencies()
	var llmDep *Dependency

	// Find the llm dependency
	for _, dep := range deps {
		if dep.Name == "llm" {
			llmDep = &dep
			break
		}
	}

	if llmDep == nil || len(llmDep.Plugins) == 0 {
		return []PluginStatus{}, nil
	}

	// Check if llm command is available
	if _, err := exec.LookPath("llm"); err != nil {
		return nil, fmt.Errorf("llm command not found")
	}

	// Get installed plugins
	cmd := exec.Command("llm", "plugins", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}

	installedPlugins := parseInstalledPlugins(string(output))
	statuses := make([]PluginStatus, len(llmDep.Plugins))

	for i, requiredPlugin := range llmDep.Plugins {
		status := PluginStatus{
			Name:      requiredPlugin,
			Installed: contains(installedPlugins, requiredPlugin),
		}

		if !status.Installed {
			status.Error = fmt.Sprintf("Plugin %s is not installed", requiredPlugin)
		}

		statuses[i] = status
	}

	return statuses, nil
}

// GetSystemInfo returns information about the host system
func GetSystemInfo() *SystemInfo {
	info := &SystemInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		Shell:        getShell(),
		PathDirs:     getPathDirs(),
	}
	return info
}

// extractVersion extracts version number from command output
func extractVersion(output string) string {
	// Handle JSON format (like files2prompt)
	if strings.Contains(output, `"Version"`) {
		re := regexp.MustCompile(`"Version"\s*:\s*"([^"]+)"`)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			version := strings.TrimSpace(matches[1])
			// Return local/dev versions as they indicate a working installation
			return version
		}
	}

	// Common version patterns
	patterns := []string{
		`version\s+([0-9]+\.[0-9]+\.[0-9]+(?:-[a-zA-Z0-9\-\.]+)?)`,
		`v?([0-9]+\.[0-9]+\.[0-9]+(?:-[a-zA-Z0-9\-\.]+)?)`,
		`([0-9]+\.[0-9]+\.[0-9]+)`,
		`([0-9]+\.[0-9]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

// parseInstalledPlugins parses the output of 'llm plugins list'
func parseInstalledPlugins(output string) []string {
	var plugins []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Plugin names are typically at the beginning of each line
		// Format might be: "plugin-name: description" or just "plugin-name"
		parts := strings.Fields(line)
		if len(parts) > 0 {
			pluginName := strings.TrimSuffix(parts[0], ":")
			if pluginName != "" && !strings.HasPrefix(pluginName, "=") {
				plugins = append(plugins, pluginName)
			}
		}
	}

	return plugins
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getShell returns the current shell
func getShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		// Fallback for Windows
		shell = os.Getenv("COMSPEC")
	}
	return shell
}

// getPathDirs returns the directories in PATH
func getPathDirs() []string {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return []string{}
	}

	separator := ":"
	if runtime.GOOS == "windows" {
		separator = ";"
	}

	return strings.Split(pathEnv, separator)
}
