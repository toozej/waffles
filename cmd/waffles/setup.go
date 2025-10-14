package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/toozej/waffles/internal/setup"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard",
	Long: `Interactive setup wizard to configure Waffles for your environment.

This wizard will guide you through:
- Checking and installing dependencies
- Configuring LLM providers and API keys
- Setting up default models and preferences
- Creating configuration files
- Validating the complete setup

Use --auto for automatic setup with sensible defaults.`,
	Run: setupRun,
}

var (
	interactiveMode bool
	autoMode        bool
)

func setupRun(cmd *cobra.Command, args []string) {
	// Get flags
	autoMode, _ = cmd.Flags().GetBool("auto")
	interactiveMode, _ = cmd.Flags().GetBool("interactive")

	// Auto mode overrides interactive mode
	if autoMode {
		interactiveMode = false
	}

	// Create and run the setup wizard
	wizard := setup.NewSetupWizard(autoMode)

	if err := wizard.Run(); err != nil {
		fmt.Printf("❌ Setup failed: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add flags
	setupCmd.Flags().BoolVar(&interactiveMode, "interactive", true, "Run in interactive mode (default)")
	setupCmd.Flags().BoolVar(&autoMode, "auto", false, "Run in automatic mode with defaults")

	// Add to root command
	rootCmd.AddCommand(setupCmd)
}
