package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/toozej/waffles/pkg/config"
	"github.com/toozej/waffles/pkg/deps"
)

// SetupWizard handles the interactive setup process
type SetupWizard struct {
	ui       *UserInterface
	config   *config.Config
	autoMode bool
}

// NewSetupWizard creates a new setup wizard
func NewSetupWizard(autoMode bool) *SetupWizard {
	return &SetupWizard{
		ui:       NewUserInterface(),
		autoMode: autoMode,
	}
}

// Run executes the complete setup wizard
func (w *SetupWizard) Run() error {
	w.ui.ShowWelcome()

	if !w.autoMode {
		if !w.ui.AskYesNo("Would you like to proceed with the setup?", true) {
			fmt.Println("Setup cancelled.")
			return nil
		}
	}

	// Load existing config or create new one
	cfg, err := config.LoadConfig()
	if err != nil {
		cfg = &config.Config{} // Start with empty config if load fails
	}
	w.config = cfg

	// Step 1: Check Dependencies
	if err := w.checkDependencies(); err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	// Step 2: Configure LLM Providers
	if err := w.configureLLMProviders(); err != nil {
		return fmt.Errorf("LLM provider configuration failed: %w", err)
	}

	// Step 3: Configure Default Settings
	if err := w.configureDefaultSettings(); err != nil {
		return fmt.Errorf("default settings configuration failed: %w", err)
	}

	// Step 4: Configure Tool Arguments
	if err := w.configureToolArguments(); err != nil {
		return fmt.Errorf("tool arguments configuration failed: %w", err)
	}

	// Step 5: Save Configuration
	if err := w.saveConfiguration(); err != nil {
		return fmt.Errorf("configuration save failed: %w", err)
	}

	// Step 6: Final Validation
	if err := w.validateSetup(); err != nil {
		w.ui.ShowWarning("Setup validation failed, but configuration has been saved.")
		w.ui.ShowWarning(err.Error())
	}

	w.showCompletionSummary()
	return nil
}

// checkDependencies handles Step 1: Check and install dependencies
func (w *SetupWizard) checkDependencies() error {
	w.ui.ShowStep(1, 6, "Checking Dependencies")

	w.ui.ShowProgress("Checking required tools...")
	statuses, err := deps.CheckAllDependencies()
	if err != nil {
		w.ui.ShowProgressComplete(false)
		return fmt.Errorf("failed to check dependencies: %w", err)
	}
	w.ui.ShowProgressComplete(true)

	allValid := true
	missingDeps := []string{}

	for _, status := range statuses {
		switch {
		case status.Installed && status.Valid:
			w.ui.ShowSuccess(fmt.Sprintf("%s: %s", status.Name, status.Version))
		case status.Installed && !status.Valid:
			w.ui.ShowWarning(fmt.Sprintf("%s: %s (needs update)", status.Name, status.Version))
			allValid = false
			missingDeps = append(missingDeps, status.Name)
		default:
			w.ui.ShowError(fmt.Sprintf("%s: not installed", status.Name))
			allValid = false
			missingDeps = append(missingDeps, status.Name)
		}
	}

	if !allValid {
		w.ui.ShowInfo(fmt.Sprintf("Found %d missing or outdated dependencies", len(missingDeps)))

		if w.autoMode || w.ui.AskYesNo("Would you like to install missing dependencies automatically?", true) {
			w.ui.ShowProgress("Installing dependencies...")

			if err := deps.AutoInstallAll(); err != nil {
				w.ui.ShowProgressComplete(false)
				w.ui.ShowError("Automatic installation failed: " + err.Error())
				w.ui.ShowInfo("Please install dependencies manually using: waffles deps install --instructions-only")
			} else {
				w.ui.ShowProgressComplete(true)
				w.ui.ShowSuccess("All dependencies installed successfully!")
			}
		} else {
			w.ui.ShowInfo("Skipping dependency installation. You can install them later with: waffles deps install")
		}
	} else {
		w.ui.ShowSuccess("All dependencies are installed and up to date!")
	}

	return nil
}

// configureLLMProviders handles Step 2: Configure LLM providers and API keys
func (w *SetupWizard) configureLLMProviders() error {
	w.ui.ShowStep(2, 6, "Configuring LLM Providers")

	providers := []string{
		"Anthropic (Claude)",
		"OpenAI (GPT)",
		"Google (Gemini)",
		"Ollama (Local)",
		"Skip provider setup",
	}

	if !w.autoMode {
		choice := w.ui.AskChoice("Which LLM provider would you like to configure?", providers, 0)

		switch choice {
		case 0: // Anthropic
			if err := w.configureAnthropic(); err != nil {
				return err
			}
		case 1: // OpenAI
			if err := w.configureOpenAI(); err != nil {
				return err
			}
		case 2: // Google
			if err := w.configureGoogle(); err != nil {
				return err
			}
		case 3: // Ollama
			if err := w.configureOllama(); err != nil {
				return err
			}
		case 4: // Skip
			w.ui.ShowInfo("Skipping provider setup. You can configure providers later.")
		}
	} else {
		// In auto mode, default to Anthropic with Claude
		w.ui.ShowInfo("Auto mode: Configuring default Anthropic provider")
		w.config.DefaultProvider = "anthropic"
		w.config.DefaultModel = "claude-3-sonnet"
	}

	return nil
}

// configureAnthropic configures Anthropic API settings
func (w *SetupWizard) configureAnthropic() error {
	w.config.DefaultProvider = "anthropic"

	models := []string{"claude-3-sonnet", "claude-3-haiku", "claude-3-opus"}
	if !w.autoMode {
		choice := w.ui.AskChoice("Select default Anthropic model:", models, 0)
		w.config.DefaultModel = models[choice]
	} else {
		w.config.DefaultModel = "claude-3-sonnet"
	}

	if !w.autoMode {
		w.ui.ShowInfo("To use Anthropic, you need an API key from: https://console.anthropic.com/")
		if w.ui.AskYesNo("Do you have an Anthropic API key?", false) {
			w.ui.ShowInfo("You can set your API key using: llm keys set anthropic")
			w.ui.ShowInfo("Or set the ANTHROPIC_API_KEY environment variable")
		}
	}

	return nil
}

// configureOpenAI configures OpenAI API settings
func (w *SetupWizard) configureOpenAI() error {
	w.config.DefaultProvider = "openai"

	models := []string{"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"}
	choice := w.ui.AskChoice("Select default OpenAI model:", models, 0)
	w.config.DefaultModel = models[choice]

	w.ui.ShowInfo("To use OpenAI, you need an API key from: https://platform.openai.com/api-keys")
	if w.ui.AskYesNo("Do you have an OpenAI API key?", false) {
		w.ui.ShowInfo("You can set your API key using: llm keys set openai")
		w.ui.ShowInfo("Or set the OPENAI_API_KEY environment variable")
	}

	return nil
}

// configureGoogle configures Google Gemini API settings
func (w *SetupWizard) configureGoogle() error {
	w.config.DefaultProvider = "google"

	models := []string{"gemini-pro", "gemini-pro-vision"}
	choice := w.ui.AskChoice("Select default Google model:", models, 0)
	w.config.DefaultModel = models[choice]

	w.ui.ShowInfo("To use Google Gemini, you need an API key from: https://makersuite.google.com/app/apikey")
	if w.ui.AskYesNo("Do you have a Google API key?", false) {
		w.ui.ShowInfo("You can set your API key using: llm keys set google")
		w.ui.ShowInfo("Or set the GOOGLE_API_KEY environment variable")
	}

	return nil
}

// configureOllama configures local Ollama settings
func (w *SetupWizard) configureOllama() error {
	w.config.DefaultProvider = "ollama"

	w.ui.ShowInfo("Ollama runs models locally. Make sure you have Ollama installed and running.")

	model := w.ui.AskString("Enter the Ollama model name", "llama2", true)
	w.config.DefaultModel = model

	return nil
}

// configureDefaultSettings handles Step 3: Configure default settings
func (w *SetupWizard) configureDefaultSettings() error {
	w.ui.ShowStep(3, 6, "Configuring Default Settings")

	if !w.autoMode {
		// Database path
		defaultDBPath := w.config.LogDBPath
		if defaultDBPath == "" {
			defaultDBPath = "./llm-logs.sqlite"
		}
		w.config.LogDBPath = w.ui.AskString("Database path for logging", defaultDBPath, false)

		// Verbose mode
		w.config.Verbose = w.ui.AskYesNo("Enable verbose logging by default?", w.config.Verbose)

		// Auto-install
		w.config.AutoInstall = w.ui.AskYesNo("Enable auto-installation of missing dependencies?", w.config.AutoInstall)

		// Gitignore respect
		w.config.IgnoreGitignore = w.ui.AskYesNo("Ignore .gitignore files when scanning files?", w.config.IgnoreGitignore)
	} else {
		// Auto mode defaults
		if w.config.LogDBPath == "" {
			w.config.LogDBPath = "./llm-logs.sqlite"
		}
		w.config.Verbose = false
		w.config.AutoInstall = true
		w.config.IgnoreGitignore = false
		w.ui.ShowInfo("Auto mode: Using default settings")
	}

	return nil
}

// configureToolArguments handles Step 4: Configure tool-specific arguments
func (w *SetupWizard) configureToolArguments() error {
	w.ui.ShowStep(4, 6, "Configuring Tool Arguments")

	if !w.autoMode {
		if w.ui.AskYesNo("Would you like to configure custom arguments for external tools?", false) {
			w.config.WheresmypromptArgs = w.ui.AskString("Additional wheresmyprompt arguments", w.config.WheresmypromptArgs, false)
			w.config.Files2promptArgs = w.ui.AskString("Additional files2prompt arguments", w.config.Files2promptArgs, false)
			w.config.LLMArgs = w.ui.AskString("Additional llm CLI arguments", w.config.LLMArgs, false)
		} else {
			w.ui.ShowInfo("Skipping tool arguments configuration")
		}
	} else {
		w.ui.ShowInfo("Auto mode: Using default tool arguments")
	}

	return nil
}

// saveConfiguration handles Step 5: Save configuration to file
func (w *SetupWizard) saveConfiguration() error {
	w.ui.ShowStep(5, 6, "Saving Configuration")

	configPath := w.config.ConfigPath
	if configPath == "" {
		configPath = ".waffles.env"
	}

	// Show configuration summary
	if !w.autoMode {
		configMap := map[string]string{
			"Provider":      w.config.DefaultProvider,
			"Model":         w.config.DefaultModel,
			"Database Path": w.config.LogDBPath,
			"Verbose":       fmt.Sprintf("%v", w.config.Verbose),
			"Auto Install":  fmt.Sprintf("%v", w.config.AutoInstall),
		}
		w.ui.ShowConfigSummary(configMap)

		if !w.ui.AskYesNo("Save configuration to "+configPath+"?", true) {
			w.ui.ShowInfo("Configuration not saved")
			return nil
		}
	}

	w.ui.ShowProgress("Saving configuration...")

	if err := w.writeConfigFile(configPath); err != nil {
		w.ui.ShowProgressComplete(false)
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	w.ui.ShowProgressComplete(true)
	w.ui.ShowSuccess("Configuration saved to " + configPath)

	return nil
}

// writeConfigFile writes the configuration to a .env file
func (w *SetupWizard) writeConfigFile(configPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create config file content
	var content strings.Builder
	content.WriteString("# Waffles Configuration\n")
	content.WriteString("# Generated by setup wizard\n\n")

	if w.config.DefaultProvider != "" {
		content.WriteString(fmt.Sprintf("WAFFLES_DEFAULT_PROVIDER=%s\n", w.config.DefaultProvider))
	}
	if w.config.DefaultModel != "" {
		content.WriteString(fmt.Sprintf("WAFFLES_DEFAULT_MODEL=%s\n", w.config.DefaultModel))
	}
	if w.config.LogDBPath != "" {
		content.WriteString(fmt.Sprintf("WAFFLES_LOG_DB_PATH=%s\n", w.config.LogDBPath))
	}

	content.WriteString(fmt.Sprintf("WAFFLES_VERBOSE=%v\n", w.config.Verbose))
	content.WriteString(fmt.Sprintf("WAFFLES_AUTO_INSTALL=%v\n", w.config.AutoInstall))
	content.WriteString(fmt.Sprintf("WAFFLES_IGNORE_GITIGNORE=%v\n", w.config.IgnoreGitignore))

	if w.config.WheresmypromptArgs != "" {
		content.WriteString(fmt.Sprintf("WHERESMYPROMPT_ARGS=%s\n", w.config.WheresmypromptArgs))
	}
	if w.config.Files2promptArgs != "" {
		content.WriteString(fmt.Sprintf("FILES2PROMPT_ARGS=%s\n", w.config.Files2promptArgs))
	}
	if w.config.LLMArgs != "" {
		content.WriteString(fmt.Sprintf("LLM_ARGS=%s\n", w.config.LLMArgs))
	}

	// Write to file
	return os.WriteFile(configPath, []byte(content.String()), 0600)
}

// validateSetup handles Step 6: Final validation
func (w *SetupWizard) validateSetup() error {
	w.ui.ShowStep(6, 6, "Validating Setup")

	w.ui.ShowProgress("Validating configuration...")

	// Re-load config to validate
	_, err := config.LoadConfig()
	if err != nil {
		w.ui.ShowProgressComplete(false)
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	w.ui.ShowProgressComplete(true)

	w.ui.ShowProgress("Re-checking dependencies...")
	statuses, err := deps.CheckAllDependencies()
	if err != nil {
		w.ui.ShowProgressComplete(false)
		return fmt.Errorf("dependency validation failed: %w", err)
	}

	allValid := true
	for _, status := range statuses {
		if !status.Installed || !status.Valid {
			allValid = false
			break
		}
	}

	w.ui.ShowProgressComplete(allValid)

	if allValid {
		w.ui.ShowSuccess("All dependencies are properly installed")
	} else {
		w.ui.ShowWarning("Some dependencies may still need attention")
	}

	return nil
}

// showCompletionSummary shows the final setup completion summary
func (w *SetupWizard) showCompletionSummary() {
	w.ui.ShowCompletionMessage()

	nextSteps := []string{
		"Try running a query: waffles \"example search term\"",
		"Check your setup anytime: waffles deps",
		"View configuration: cat .waffles.env",
		"Get help: waffles --help",
	}

	if w.config.DefaultProvider != "ollama" {
		nextSteps = append([]string{"Set up your API keys using: llm keys set " + w.config.DefaultProvider}, nextSteps...)
	}

	w.ui.ShowNextSteps(nextSteps)

	w.ui.ShowSeparator()
	w.ui.ShowInfo("Setup wizard completed successfully!")
}
