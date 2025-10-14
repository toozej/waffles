package setup

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// UserInterface handles user input and output during setup
type UserInterface struct {
	reader *bufio.Reader
}

// NewUserInterface creates a new user interface
func NewUserInterface() *UserInterface {
	return &UserInterface{
		reader: bufio.NewReader(os.Stdin),
	}
}

// ShowWelcome displays the setup wizard welcome message
func (ui *UserInterface) ShowWelcome() {
	fmt.Println("🧙‍♂️ Welcome to the Waffles Setup Wizard!")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("This wizard will help you:")
	fmt.Println("• Check and install dependencies")
	fmt.Println("• Configure LLM providers and API keys")
	fmt.Println("• Set up default models and preferences")
	fmt.Println("• Create configuration files")
	fmt.Println("• Validate your complete setup")
	fmt.Println()
}

// ShowStep displays a step header
func (ui *UserInterface) ShowStep(step int, total int, title string) {
	fmt.Printf("\n📋 Step %d/%d: %s\n", step, total, title)
	fmt.Println(strings.Repeat("─", len(title)+20))
}

// AskYesNo asks a yes/no question and returns the user's choice
func (ui *UserInterface) AskYesNo(question string, defaultValue bool) bool {
	defaultStr := "y/N"
	if defaultValue {
		defaultStr = "Y/n"
	}

	fmt.Printf("%s [%s]: ", question, defaultStr)

	input, err := ui.reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}

	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return defaultValue
	}

	return input == "y" || input == "yes"
}

// AskString asks for a string input with validation
func (ui *UserInterface) AskString(question, defaultValue string, required bool) string {
	for {
		if defaultValue != "" {
			fmt.Printf("%s [%s]: ", question, defaultValue)
		} else {
			fmt.Printf("%s: ", question)
		}

		input, err := ui.reader.ReadString('\n')
		if err != nil {
			if defaultValue != "" {
				return defaultValue
			}
			continue
		}

		input = strings.TrimSpace(input)

		if input == "" {
			if defaultValue != "" {
				return defaultValue
			}
			if required {
				fmt.Println("⚠️  This field is required. Please enter a value.")
				continue
			}
		}

		return input
	}
}

// AskChoice asks user to choose from a list of options
func (ui *UserInterface) AskChoice(question string, options []string, defaultIndex int) int {
	fmt.Println(question)

	for i, option := range options {
		marker := " "
		if i == defaultIndex {
			marker = ">"
		}
		fmt.Printf("%s %d. %s\n", marker, i+1, option)
	}

	for {
		fmt.Printf("Choose an option [%d]: ", defaultIndex+1)

		input, err := ui.reader.ReadString('\n')
		if err != nil {
			return defaultIndex
		}

		input = strings.TrimSpace(input)

		if input == "" {
			return defaultIndex
		}

		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(options) {
			fmt.Printf("⚠️  Please enter a number between 1 and %d.\n", len(options))
			continue
		}

		return choice - 1
	}
}

// AskPassword asks for a password/API key input (hidden)
func (ui *UserInterface) AskPassword(question string) string {
	fmt.Printf("%s: ", question)

	// For simplicity, we'll use regular input for now
	// In a production version, you'd want to use a library like golang.org/x/crypto/ssh/terminal
	input, err := ui.reader.ReadString('\n')
	if err != nil {
		return ""
	}

	return strings.TrimSpace(input)
}

// ShowSuccess displays a success message
func (ui *UserInterface) ShowSuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}

// ShowError displays an error message
func (ui *UserInterface) ShowError(message string) {
	fmt.Printf("❌ %s\n", message)
}

// ShowWarning displays a warning message
func (ui *UserInterface) ShowWarning(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// ShowInfo displays an informational message
func (ui *UserInterface) ShowInfo(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// ShowProgress displays a progress indicator
func (ui *UserInterface) ShowProgress(message string) {
	fmt.Printf("⏳ %s", message)
}

// ShowProgressComplete completes a progress indicator
func (ui *UserInterface) ShowProgressComplete(success bool) {
	if success {
		fmt.Println(" ✓")
	} else {
		fmt.Println(" ✗")
	}
}

// ShowList displays a list of items
func (ui *UserInterface) ShowList(title string, items []string) {
	fmt.Printf("\n%s:\n", title)
	for _, item := range items {
		fmt.Printf("  • %s\n", item)
	}
	fmt.Println()
}

// Confirm asks for confirmation before proceeding
func (ui *UserInterface) Confirm(message string) bool {
	return ui.AskYesNo(message, false)
}

// ShowSeparator displays a visual separator
func (ui *UserInterface) ShowSeparator() {
	fmt.Println(strings.Repeat("═", 60))
}

// WaitForEnter waits for the user to press Enter
func (ui *UserInterface) WaitForEnter(message string) {
	if message == "" {
		message = "Press Enter to continue..."
	}
	fmt.Printf("\n%s", message)
	_, _ = ui.reader.ReadString('\n') // Ignore errors for user input waiting
}

// ShowConfigSummary displays a configuration summary
func (ui *UserInterface) ShowConfigSummary(config map[string]string) {
	fmt.Println("\n📄 Configuration Summary:")
	fmt.Println("─────────────────────────")

	for key, value := range config {
		// Hide sensitive values
		displayValue := value
		if strings.Contains(strings.ToLower(key), "key") || strings.Contains(strings.ToLower(key), "token") {
			if len(value) > 8 {
				displayValue = value[:4] + "..." + value[len(value)-4:]
			} else if value != "" {
				displayValue = strings.Repeat("*", len(value))
			}
		}

		fmt.Printf("  %s: %s\n", key, displayValue)
	}
	fmt.Println()
}

// ShowNextSteps displays recommended next steps
func (ui *UserInterface) ShowNextSteps(steps []string) {
	fmt.Println("\n🎯 Next Steps:")
	fmt.Println("──────────────")

	for i, step := range steps {
		fmt.Printf("%d. %s\n", i+1, step)
	}
	fmt.Println()
}

// ShowCompletionMessage displays the setup completion message
func (ui *UserInterface) ShowCompletionMessage() {
	fmt.Println()
	fmt.Println("🎉 Setup Complete!")
	fmt.Println("==================")
	fmt.Println()
	fmt.Println("Waffles is now configured and ready to use!")
	fmt.Println()
}
