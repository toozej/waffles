package deps

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// GetInstallInstructions returns installation instructions for a dependency
func GetInstallInstructions(dep Dependency) string {
	sysInfo := GetSystemInfo()

	switch dep.Name {
	case "wheresmyprompt":
		return getWheresmypromptInstructions(sysInfo)
	case "files2prompt":
		return getFiles2promptInstructions(sysInfo)
	case "llm":
		return getLLMInstructions(sysInfo)
	default:
		return fmt.Sprintf("Visit %s for installation instructions", dep.InstallURL)
	}
}

// InstallLLMPlugin attempts to install an LLM plugin
func InstallLLMPlugin(plugin string) error {
	// Check if llm is available
	if _, err := exec.LookPath("llm"); err != nil {
		return fmt.Errorf("llm command not found - install llm CLI first")
	}

	// Install the plugin
	cmd := exec.Command("llm", "install", plugin)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install plugin %s: %w\nOutput: %s", plugin, err, string(output))
	}

	return nil
}

// AutoInstallAll attempts to auto-install all missing dependencies
func AutoInstallAll() error {
	installer := NewPlatformInstaller()
	deps := RequiredDependencies()

	fmt.Printf("Starting auto-installation for %d dependencies...\n", len(deps))

	var errors []string
	successCount := 0

	for _, dep := range deps {
		fmt.Printf("\nChecking dependency: %s\n", dep.Name)

		// Check current status
		status, err := CheckDependency(dep)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to check %s: %v", dep.Name, err))
			continue
		}

		if status.Installed && status.Valid {
			fmt.Printf("✓ %s is already installed and valid (version %s)\n", dep.Name, status.Version)
			successCount++
			continue
		}

		// Attempt installation
		result, err := InstallDependency(dep)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to install %s: %v", dep.Name, err))
			continue
		}

		if result.Success {
			fmt.Printf("✓ %s\n", result.Message)

			// Verify installation
			if err := installer.VerifyInstallation(dep.Command); err != nil {
				fmt.Printf("⚠ Warning: %s may not be accessible in PATH: %v\n", dep.Name, err)
			} else {
				successCount++
			}

			// Install plugins for llm
			if dep.Name == "llm" && len(dep.Plugins) > 0 {
				fmt.Printf("Installing %d LLM plugins...\n", len(dep.Plugins))
				if err := installLLMPlugins(dep.Plugins); err != nil {
					fmt.Printf("⚠ Warning: Plugin installation issues: %v\n", err)
				}
			}
		} else {
			errors = append(errors, fmt.Sprintf("Installation failed for %s: %s", dep.Name, result.Error))
		}
	}

	// Summary
	fmt.Printf("\nInstallation Summary:\n")
	fmt.Printf("✓ Successfully installed/verified: %d/%d dependencies\n", successCount, len(deps))

	if len(errors) > 0 {
		fmt.Printf("✗ Issues encountered:\n")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		fmt.Printf("\nFor manual installation instructions, run: waffles deps\n")
		return fmt.Errorf("%d installation issues encountered", len(errors))
	}

	fmt.Printf("\n🎉 All dependencies successfully installed!\n")
	return nil
}

// InstallDependency attempts to install a single dependency
func InstallDependency(dep Dependency) (*InstallationResult, error) {
	installer := NewPlatformInstaller()

	switch dep.Name {
	case "wheresmyprompt":
		return installer.InstallWheresmyprompt()
	case "files2prompt":
		return installer.InstallFiles2prompt()
	case "llm":
		return installer.InstallLLM()
	default:
		return &InstallationResult{
			Success: false,
			Error:   fmt.Sprintf("Unknown dependency: %s", dep.Name),
		}, fmt.Errorf("unknown dependency: %s", dep.Name)
	}
}

// getWheresmypromptInstructions returns platform-specific installation instructions for wheresmyprompt
func getWheresmypromptInstructions(sysInfo *SystemInfo) string {
	var instructions strings.Builder

	instructions.WriteString("Install wheresmyprompt:\n\n")

	switch sysInfo.OS {
	case "darwin": // macOS
		if hasHomebrew() {
			instructions.WriteString("Using Homebrew:\n")
			instructions.WriteString("  brew install toozej/tap/wheresmyprompt\n\n")
		}
		instructions.WriteString("Using Go:\n")
		instructions.WriteString("  go install github.com/toozej/wheresmyprompt/cmd/wheresmyprompt@latest\n\n")
		instructions.WriteString("Or download binary from:\n")
		instructions.WriteString("  https://github.com/toozej/wheresmyprompt/releases\n")

	case "linux":
		instructions.WriteString("Using Go:\n")
		instructions.WriteString("  go install github.com/toozej/wheresmyprompt/cmd/wheresmyprompt@latest\n\n")
		instructions.WriteString("Or download binary from:\n")
		instructions.WriteString("  https://github.com/toozej/wheresmyprompt/releases\n")

	case "windows":
		instructions.WriteString("Using Go:\n")
		instructions.WriteString("  go install github.com/toozej/wheresmyprompt/cmd/wheresmyprompt@latest\n\n")
		instructions.WriteString("Or download binary from:\n")
		instructions.WriteString("  https://github.com/toozej/wheresmyprompt/releases\n")

	default:
		instructions.WriteString("Visit https://github.com/toozej/wheresmyprompt#installation\n")
	}

	return instructions.String()
}

// getFiles2promptInstructions returns platform-specific installation instructions for files2prompt
func getFiles2promptInstructions(sysInfo *SystemInfo) string {
	var instructions strings.Builder

	instructions.WriteString("Install files2prompt:\n\n")

	switch sysInfo.OS {
	case "darwin": // macOS
		if hasHomebrew() {
			instructions.WriteString("Using Homebrew:\n")
			instructions.WriteString("  brew install toozej/tap/files2prompt\n\n")
		}
		instructions.WriteString("Using Go:\n")
		instructions.WriteString("  go install github.com/toozej/files2prompt/cmd/files2prompt@latest\n\n")
		instructions.WriteString("Or download binary from:\n")
		instructions.WriteString("  https://github.com/toozej/files2prompt/releases\n")

	case "linux":
		instructions.WriteString("Using Go:\n")
		instructions.WriteString("  go install github.com/toozej/files2prompt/cmd/files2prompt@latest\n\n")
		instructions.WriteString("Or download binary from:\n")
		instructions.WriteString("  https://github.com/toozej/files2prompt/releases\n")

	case "windows":
		instructions.WriteString("Using Go:\n")
		instructions.WriteString("  go install github.com/toozej/files2prompt/cmd/files2prompt@latest\n\n")
		instructions.WriteString("Or download binary from:\n")
		instructions.WriteString("  https://github.com/toozej/files2prompt/releases\n")

	default:
		instructions.WriteString("Visit https://github.com/toozej/files2prompt#installation\n")
	}

	return instructions.String()
}

// getLLMInstructions returns platform-specific installation instructions for llm CLI
func getLLMInstructions(sysInfo *SystemInfo) string {
	var instructions strings.Builder

	instructions.WriteString("Install llm CLI:\n\n")

	switch sysInfo.OS {
	case "darwin": // macOS
		if hasHomebrew() {
			instructions.WriteString("Using Homebrew:\n")
			instructions.WriteString("  brew install llm\n\n")
		}
		instructions.WriteString("Using pipx (recommended):\n")
		instructions.WriteString("  pipx install llm\n\n")
		instructions.WriteString("Using pip:\n")
		instructions.WriteString("  pip install llm\n\n")

	case "linux":
		instructions.WriteString("Using pipx (recommended):\n")
		instructions.WriteString("  pipx install llm\n\n")
		instructions.WriteString("Using pip:\n")
		instructions.WriteString("  pip install llm\n\n")
		if hasApt() {
			instructions.WriteString("On Ubuntu/Debian, first install pipx:\n")
			instructions.WriteString("  sudo apt install pipx\n\n")
		}

	case "windows":
		instructions.WriteString("Using pipx (recommended):\n")
		instructions.WriteString("  pipx install llm\n\n")
		instructions.WriteString("Using pip:\n")
		instructions.WriteString("  pip install llm\n\n")

	default:
		instructions.WriteString("Using pip:\n")
		instructions.WriteString("  pip install llm\n\n")
	}

	instructions.WriteString("After installation, install required plugins:\n")
	for _, plugin := range RequiredDependencies()[2].Plugins { // llm is the 3rd dependency
		instructions.WriteString(fmt.Sprintf("  llm install %s\n", plugin))
	}

	instructions.WriteString("\nFor more details: https://llm.datasette.io/en/stable/setup.html\n")

	return instructions.String()
}

// hasHomebrew checks if Homebrew is installed (macOS/Linux)
func hasHomebrew() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

// hasApt checks if apt package manager is available (Debian/Ubuntu)
func hasApt() bool {
	_, err := exec.LookPath("apt")
	return err == nil
}

// DetectPackageManager detects the available package manager on the system
func DetectPackageManager() string {
	// Check for Go first as it's preferred for Go binaries
	if _, err := exec.LookPath("go"); err == nil {
		return "go"
	}

	// Platform-specific package managers
	if runtime.GOOS == "darwin" && hasHomebrew() {
		return "homebrew"
	}
	if runtime.GOOS == "linux" {
		if hasApt() {
			return "apt"
		}
		if _, err := exec.LookPath("yum"); err == nil {
			return "yum"
		}
		if _, err := exec.LookPath("pacman"); err == nil {
			return "pacman"
		}
	}

	// Python package managers
	if _, err := exec.LookPath("pipx"); err == nil {
		return "pipx"
	}
	if _, err := exec.LookPath("pip"); err == nil {
		return "pip"
	}

	return "unknown"
}

// installLLMPlugins installs multiple LLM plugins
func installLLMPlugins(plugins []string) error {
	var errors []string
	successCount := 0

	for _, plugin := range plugins {
		fmt.Printf("  Installing plugin: %s...", plugin)

		if err := InstallLLMPlugin(plugin); err != nil {
			fmt.Printf(" ✗\n")
			errors = append(errors, fmt.Sprintf("Failed to install %s: %v", plugin, err))
		} else {
			fmt.Printf(" ✓\n")
			successCount++
		}
	}

	fmt.Printf("Plugin installation completed: %d/%d successful\n", successCount, len(plugins))

	if len(errors) > 0 {
		return fmt.Errorf("plugin installation issues: %v", errors)
	}

	return nil
}
