// Package cmd provides the command-line interface for the waffles LLM toolchain orchestrator.
//
// This package implements the complete CLI application using the cobra library,
// providing a comprehensive interface for LLM workflow automation. The package
// handles command parsing, configuration management, and orchestrates the execution
// of the LLM toolchain pipeline.
//
// Key features:
//   - Root command with comprehensive flag support
//   - Subcommands for setup, querying, exporting, and dependency management
//   - Configuration override via CLI flags
//   - Pipeline execution with wheresmyprompt → files2prompt → llm flow
//   - Comprehensive logging and error handling
//
// The CLI follows standard Unix conventions and provides both interactive
// and scriptable interfaces for LLM workflow automation.
//
// Example usage:
//
//	import "github.com/toozej/waffles/cmd/waffles"
//
//	func main() {
//		cmd.Execute()
//	}
package cmd

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/toozej/waffles/pkg/config"
	"github.com/toozej/waffles/pkg/man"
	"github.com/toozej/waffles/pkg/pipeline"
	"github.com/toozej/waffles/pkg/version"
)

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "waffles [options] [prompt-search-terms...]",
	Short: "LLM toolchain orchestrator for local development workflows",
	Long: `Waffles automates the process of gathering project context, retrieving prompts,
and executing LLM queries while maintaining comprehensive logging for transparency.

Examples:
  waffles "code review" "golang best practices"
  waffles --language python "refactor this code"
  waffles --files2prompt-args="--ignore vendor/" "document this API"`,
	Args:             cobra.ArbitraryArgs,
	PersistentPreRun: rootCmdPreRun,
	Run:              rootCmdRun,
}

func rootCmdRun(cmd *cobra.Command, args []string) {
	fmt.Printf("🧇 Waffles - LLM toolchain orchestrator\n")
	fmt.Printf("Configuration: Model=%s, Provider=%s\n", cfg.DefaultModel, cfg.DefaultProvider)

	// Get prompt query from arguments
	promptQuery := strings.Join(args, " ")
	if promptQuery == "" {
		promptQuery = "help me with this code"
	}
	fmt.Printf("Prompt query: %s\n", promptQuery)
	fmt.Println()

	// Create and execute pipeline
	pipelineInstance := pipeline.NewPipeline(cfg)

	// Check if this is a dry run or dependency check
	checkDeps, _ := cmd.Flags().GetBool("check-deps")
	if checkDeps {
		fmt.Println("🔍 Checking dependencies only...")
		return
	}

	fmt.Println("🚀 Starting Waffles pipeline execution...")
	fmt.Println("Pipeline: wheresmyprompt → files2prompt → llm")
	fmt.Println()

	// Execute pipeline
	execContext, err := pipelineInstance.Execute(promptQuery, []string{})

	if err != nil {
		fmt.Printf("❌ Pipeline execution failed: %v\n", err)
		fmt.Println()

		// Show execution steps that completed (only if execContext is not nil)
		if execContext != nil && len(execContext.ExecutionSteps) > 0 {
			fmt.Println("Steps completed:")
			for i, step := range execContext.ExecutionSteps {
				status := "❌"
				if step.Success {
					status = "✅"
				}
				fmt.Printf("  %d. %s %s (%.2fs)\n", i+1, status, step.Tool, step.Duration.Seconds())
			}
		}

		os.Exit(1)
	}

	// Show successful execution results
	fmt.Println("✅ Pipeline completed successfully!")
	fmt.Printf("⏱️  Total execution time: %.2fs\n", execContext.Duration.Seconds())
	fmt.Println()

	// Show execution steps
	fmt.Println("Execution steps:")
	for i, step := range execContext.ExecutionSteps {
		fmt.Printf("  %d. ✅ %s (%.2fs)\n", i+1, step.Tool, step.Duration.Seconds())
		if cfg.Verbose && step.Output != "" {
			fmt.Printf("     Output: %s\n", truncateOutput(step.Output, 200))
		}
	}

	fmt.Println()
	fmt.Println("🎯 Final Result:")
	fmt.Println("================")
	fmt.Println(execContext.FinalOutput)
}

// truncateOutput truncates output for display
func truncateOutput(output string, maxLen int) string {
	if len(output) <= maxLen {
		return output
	}
	return output[:maxLen] + "..."
}

func rootCmdPreRun(cmd *cobra.Command, args []string) {
	// Load configuration
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Override config with CLI flags
	overrideConfigWithFlags(cmd)

	// Set debug level if requested
	debug, _ := cmd.Flags().GetBool("debug")
	if debug || cfg.Verbose {
		log.SetLevel(log.DebugLevel)
	}
	if cfg.Quiet {
		log.SetLevel(log.ErrorLevel)
	}
}

// overrideConfigWithFlags applies CLI flag values to override configuration
func overrideConfigWithFlags(cmd *cobra.Command) {
	if cmd.Flags().Changed("model") {
		cfg.DefaultModel, _ = cmd.Flags().GetString("model")
	}
	if cmd.Flags().Changed("provider") {
		cfg.DefaultProvider, _ = cmd.Flags().GetString("provider")
	}
	if cmd.Flags().Changed("wheresmyprompt-args") {
		cfg.WheresmypromptArgs, _ = cmd.Flags().GetString("wheresmyprompt-args")
	}
	if cmd.Flags().Changed("files2prompt-args") {
		cfg.Files2promptArgs, _ = cmd.Flags().GetString("files2prompt-args")
	}
	if cmd.Flags().Changed("llm-args") {
		cfg.LLMArgs, _ = cmd.Flags().GetString("llm-args")
	}
	if cmd.Flags().Changed("language") {
		cfg.LanguageOverride, _ = cmd.Flags().GetString("language")
	}
	if cmd.Flags().Changed("include") {
		cfg.IncludePatterns, _ = cmd.Flags().GetString("include")
	}
	if cmd.Flags().Changed("exclude") {
		cfg.ExcludePatterns, _ = cmd.Flags().GetString("exclude")
	}
	if cmd.Flags().Changed("ignore-gitignore") {
		cfg.IgnoreGitignore, _ = cmd.Flags().GetBool("ignore-gitignore")
	}
	if cmd.Flags().Changed("log-db") {
		cfg.LogDBPath, _ = cmd.Flags().GetString("log-db")
	}
	if cmd.Flags().Changed("verbose") {
		cfg.Verbose, _ = cmd.Flags().GetBool("verbose")
	}
	if cmd.Flags().Changed("quiet") {
		cfg.Quiet, _ = cmd.Flags().GetBool("quiet")
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func init() {
	// create rootCmd-level flags
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug-level logging")

	// Model and Provider Options
	rootCmd.Flags().StringP("model", "m", "", "LLM model to use (overrides config)")
	rootCmd.Flags().String("provider", "", "LLM provider override")

	// Tool Configuration
	rootCmd.Flags().String("wheresmyprompt-args", "", "Pass arguments to wheresmyprompt")
	rootCmd.Flags().String("files2prompt-args", "", "Pass arguments to files2prompt")
	rootCmd.Flags().String("llm-args", "", "Pass arguments to llm CLI")

	// Repository Options
	rootCmd.Flags().String("language", "", "Override auto-detected language")
	rootCmd.Flags().String("include", "", "Additional file patterns to include")
	rootCmd.Flags().String("exclude", "", "File patterns to exclude")
	rootCmd.Flags().Bool("ignore-gitignore", false, "Don't respect .gitignore rules")

	// Installation and Setup
	rootCmd.Flags().Bool("install", false, "Auto-install missing dependencies")
	rootCmd.Flags().Bool("check-deps", false, "Only check dependencies")

	// Output and Logging
	rootCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	rootCmd.Flags().String("log-db", "", "SQLite database path")
	rootCmd.Flags().BoolP("quiet", "q", false, "Suppress progress output")
	rootCmd.Flags().BoolP("verbose", "v", false, "Detailed execution logging")

	// Configuration
	rootCmd.Flags().String("config", "", "Configuration file path (.env)")
	rootCmd.Flags().String("env-file", "", "Alternative .env file path")

	// add sub-commands
	rootCmd.AddCommand(
		man.NewManCmd(),
		version.Command(),
	)
}
