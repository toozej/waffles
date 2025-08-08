package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/toozej/waffles/internal/export"
	"github.com/toozej/waffles/internal/query"
	"github.com/toozej/waffles/pkg/config"
	"github.com/toozej/waffles/pkg/logging"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export logged data",
	Long: `Export logged conversation data to various formats.

This command allows you to export your logged LLM conversations and
execution data to different formats for analysis, backup, or sharing.

Supported formats: json, csv, markdown, sql, template`,
	Run: exportRun,
}

func exportRun(cmd *cobra.Command, args []string) {
	// Load configuration to get database path
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("❌ Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Parse flags
	format, _ := cmd.Flags().GetString("format")
	output, _ := cmd.Flags().GetString("output")
	since, _ := cmd.Flags().GetString("since")
	until, _ := cmd.Flags().GetString("until")
	model, _ := cmd.Flags().GetString("model")
	provider, _ := cmd.Flags().GetString("provider")
	language, _ := cmd.Flags().GetString("language")
	compress, _ := cmd.Flags().GetBool("compress")
	includeFiles, _ := cmd.Flags().GetBool("include-files")
	includeSteps, _ := cmd.Flags().GetBool("include-steps")
	includeStats, _ := cmd.Flags().GetBool("include-stats")
	limit, _ := cmd.Flags().GetInt("limit")
	templateFile, _ := cmd.Flags().GetString("template")

	// Validate format
	exportFormat := export.ExportFormat(format)
	switch exportFormat {
	case export.FormatJSON, export.FormatCSV, export.FormatMarkdown, export.FormatSQL, export.FormatTemplate:
		// Valid formats
	default:
		fmt.Printf("❌ Unsupported format: %s\n", format)
		fmt.Println("Supported formats: json, csv, markdown, sql, template")
		os.Exit(1)
	}

	// Create query engine
	queryEngine, err := query.NewQueryEngine(cfg.LogDBPath)
	if err != nil {
		fmt.Printf("❌ Failed to create query engine: %v\n", err)
		os.Exit(1)
	}

	// Build filter
	filter := &logging.ExecutionFilter{
		Language: language,
		Model:    model,
		Provider: provider,
		Limit:    limit,
	}

	// Parse date filters
	if since != "" {
		sinceTime, err := query.ParseDateFilter(since)
		if err != nil {
			fmt.Printf("❌ Invalid since date: %v\n", err)
			if closeErr := queryEngine.Close(); closeErr != nil {
				fmt.Printf("Warning: failed to close query engine: %v\n", closeErr)
			}
			os.Exit(1)
		}
		filter.DateFrom = sinceTime
	}

	if until != "" {
		untilTime, err := query.ParseDateFilter(until)
		if err != nil {
			fmt.Printf("❌ Invalid until date: %v\n", err)
			if closeErr := queryEngine.Close(); closeErr != nil {
				fmt.Printf("Warning: failed to close query engine: %v\n", closeErr)
			}
			os.Exit(1)
		}
		filter.DateTo = untilTime
	}

	// Prepare export options
	options := &export.ExportOptions{
		Format:       exportFormat,
		Filter:       filter,
		IncludeFiles: includeFiles,
		IncludeSteps: includeSteps,
		IncludeStats: includeStats,
		Compress:     compress,
	}

	// Handle template format
	if exportFormat == export.FormatTemplate {
		if templateFile == "" {
			fmt.Printf("❌ Template file is required for template format\n")
			if closeErr := queryEngine.Close(); closeErr != nil {
				fmt.Printf("Warning: failed to close query engine: %v\n", closeErr)
			}
			os.Exit(1)
		}

		templateContent, err := os.ReadFile(templateFile) // #nosec G304 -- Template file from user-specified path
		if err != nil {
			fmt.Printf("❌ Failed to read template file: %v\n", err)
			if closeErr := queryEngine.Close(); closeErr != nil {
				fmt.Printf("Warning: failed to close query engine: %v\n", closeErr)
			}
			os.Exit(1)
		}

		options.Template = string(templateContent)
	}

	// Create exporter
	exporter := export.NewExporter(queryEngine, options)

	// Export to file or stdout
	if output != "" {
		fmt.Printf("🚀 Exporting to file: %s\n", output)
		if err := exporter.ExportToFile(output); err != nil {
			fmt.Printf("❌ Export failed: %v\n", err)
			if closeErr := queryEngine.Close(); closeErr != nil {
				fmt.Printf("Warning: failed to close query engine: %v\n", closeErr)
			}
			os.Exit(1)
		}
		fmt.Printf("✅ Export completed successfully: %s\n", output)
	} else {
		// Use silent progress reporter for stdout output
		exporter.SetProgressReporter(&export.SilentProgressReporter{})

		if err := exporter.Export(os.Stdout); err != nil {
			fmt.Printf("❌ Export failed: %v\n", err)
			if closeErr := queryEngine.Close(); closeErr != nil {
				fmt.Printf("Warning: failed to close query engine: %v\n", closeErr)
			}
			os.Exit(1)
		}
	}

	// Clean up query engine
	if closeErr := queryEngine.Close(); closeErr != nil {
		fmt.Printf("Warning: failed to close query engine: %v\n", closeErr)
	}
}

func init() {
	// Add export flags
	exportCmd.Flags().String("format", "json", "Export format (json, csv, markdown, sql, template)")
	exportCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	exportCmd.Flags().String("since", "", "Export data since date (YYYY-MM-DD)")
	exportCmd.Flags().String("until", "", "Export data until date (YYYY-MM-DD)")
	exportCmd.Flags().String("model", "", "Filter by LLM model")
	exportCmd.Flags().String("provider", "", "Filter by LLM provider")
	exportCmd.Flags().String("language", "", "Filter by detected language")
	exportCmd.Flags().Int("limit", 0, "Maximum number of records (0 = no limit)")
	exportCmd.Flags().Bool("compress", false, "Compress output using gzip")
	exportCmd.Flags().Bool("include-files", false, "Include file information in export")
	exportCmd.Flags().Bool("include-steps", false, "Include pipeline step details in export")
	exportCmd.Flags().Bool("include-stats", false, "Include statistical analysis in export")
	exportCmd.Flags().String("template", "", "Template file path (required for template format)")

	// Add to root command
	rootCmd.AddCommand(exportCmd)
}
