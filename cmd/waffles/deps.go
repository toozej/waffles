package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/toozej/waffles/pkg/deps"
)

var depsCmd = &cobra.Command{
	Use:   "deps",
	Short: "Dependency management",
	Long: `Check, install, and manage external tool dependencies.

The deps command helps you verify that all required tools are installed
and available in your system PATH:
- wheresmyprompt (Go-based prompt retrieval)
- files2prompt (Go-based context extraction)  
- llm (Python-based LLM CLI with SQLite logging)
- Required LLM plugins (llm-anthropic, llm-ollama, etc.)`,
	Run: depsRun,
}

var depsCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check dependency status",
	Long:  `Check if all required dependencies are installed and available.`,
	Run:   depsCheckRun,
}

var depsInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install missing dependencies",
	Long: `Automatically install missing dependencies where possible.

This command will attempt to install missing dependencies using the best
available method for your platform (Homebrew, Go, pip, pipx, etc.).

Use --instructions-only to see installation commands without executing them.`,
	Run: depsInstallRun,
}

var instructionsOnly bool

func depsRun(cmd *cobra.Command, args []string) {
	fmt.Println("🔍 Checking Waffles Dependencies")
	fmt.Println("================================")
	fmt.Println()

	statuses, err := deps.CheckAllDependencies()
	if err != nil {
		fmt.Printf("❌ Error checking dependencies: %v\n", err)
		os.Exit(1)
	}

	allValid := true
	for _, status := range statuses {
		switch {
		case status.Installed && status.Valid:
			fmt.Printf("✅ %s: %s", status.Name, status.Version)
			if status.Version == "" {
				fmt.Printf("installed")
			}
			fmt.Println()
		case status.Installed && !status.Valid:
			fmt.Printf("⚠️  %s: %s (needs update - %s)\n", status.Name, status.Version, status.Error)
			allValid = false
		default:
			fmt.Printf("❌ %s: not installed\n", status.Name)
			allValid = false
		}

		// Show plugin status for llm
		if status.Name == "llm" && len(status.Plugins) > 0 {
			fmt.Println("   Plugins:")
			for _, plugin := range status.Plugins {
				if plugin.Installed {
					fmt.Printf("   ✅ %s\n", plugin.Name)
				} else {
					fmt.Printf("   ❌ %s: %s\n", plugin.Name, plugin.Error)
					allValid = false
				}
			}
		}
	}

	fmt.Println()
	if allValid {
		fmt.Println("🎉 All dependencies are installed and up to date!")
	} else {
		fmt.Println("❗ Some dependencies are missing or need updates.")
		fmt.Println("   Run 'waffles deps install' to install missing dependencies")
		fmt.Println("   Or run 'waffles setup' for guided installation")
	}
}

func depsCheckRun(cmd *cobra.Command, args []string) {
	statuses, err := deps.CheckAllDependencies()
	if err != nil {
		fmt.Printf("Error checking dependencies: %v\n", err)
		os.Exit(1)
	}

	// Output in a more structured format for scripting
	fmt.Println("Dependency Status Report:")
	fmt.Println("========================")

	allValid := true
	for _, status := range statuses {
		fmt.Printf("- %s: ", status.Name)
		switch {
		case status.Installed && status.Valid:
			fmt.Printf("OK (%s)\n", status.Version)
		case status.Installed && !status.Valid:
			fmt.Printf("OUTDATED (%s - %s)\n", status.Version, status.Error)
			allValid = false
		default:
			fmt.Printf("MISSING\n")
			allValid = false
		}
	}

	if !allValid {
		os.Exit(1)
	}
}

func depsInstallRun(cmd *cobra.Command, args []string) {
	if instructionsOnly {
		depsInstallInstructionsOnly()
		return
	}

	fmt.Println("📦 Auto-Installing Missing Dependencies")
	fmt.Println("======================================")
	fmt.Println()

	// Check current status first
	statuses, err := deps.CheckAllDependencies()
	if err != nil {
		fmt.Printf("❌ Error checking dependencies: %v\n", err)
		os.Exit(1)
	}

	hasMissing := false
	for _, status := range statuses {
		if !status.Installed || !status.Valid {
			hasMissing = true
			break
		}
	}

	if !hasMissing {
		fmt.Println("✅ All dependencies are already installed!")
		return
	}

	// Show platform info
	installer := deps.NewPlatformInstaller()
	fmt.Printf("🖥️  Platform: %s %s\n", installer.OS, installer.Architecture)
	fmt.Printf("📦 Package Manager: %s\n", installer.PackageManager)
	fmt.Println()

	// Attempt auto-installation
	if err := deps.AutoInstallAll(); err != nil {
		fmt.Printf("❌ Auto-installation completed with issues: %v\n", err)
		fmt.Println()
		fmt.Println("💡 For manual installation instructions, run:")
		fmt.Println("   waffles deps install --instructions-only")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ Installation completed! Run 'waffles deps check' to verify.")
}

func depsInstallInstructionsOnly() {
	fmt.Println("📋 Installation Instructions")
	fmt.Println("===========================")
	fmt.Println()

	statuses, err := deps.CheckAllDependencies()
	if err != nil {
		fmt.Printf("❌ Error checking dependencies: %v\n", err)
		os.Exit(1)
	}

	hasMissing := false
	for _, status := range statuses {
		if !status.Installed || !status.Valid {
			hasMissing = true
			break
		}
	}

	if !hasMissing {
		fmt.Println("✅ All dependencies are already installed!")
		return
	}

	fmt.Println("� Manual Installation Instructions:")
	fmt.Println()

	dependencies := deps.RequiredDependencies()
	for i, status := range statuses {
		if !status.Installed || !status.Valid {
			fmt.Printf("📋 %s:\n", status.Name)
			instructions := deps.GetInstallInstructions(dependencies[i])
			fmt.Println(instructions)
			fmt.Println(strings.Repeat("-", 50))
		}
	}

	fmt.Println("💡 Tip: Run 'waffles deps install' for automatic installation")
	fmt.Println("    or use 'waffles setup' for an interactive guide")
}

func init() {
	// Add flags to install command
	depsInstallCmd.Flags().BoolVar(&instructionsOnly, "instructions-only", false, "Show installation instructions without executing them")

	// Add subcommands
	depsCmd.AddCommand(depsCheckCmd)
	depsCmd.AddCommand(depsInstallCmd)

	// Add to root command
	rootCmd.AddCommand(depsCmd)
}
