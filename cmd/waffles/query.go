package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/toozej/waffles/internal/query"
	"github.com/toozej/waffles/pkg/config"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query logged conversations",
	Long: `Query and search through logged LLM conversations and executions.

This command allows you to search through the SQLite database of logged
conversations, filter by various criteria, and analyze usage patterns.`,
	Run: queryRun,
}

func queryRun(cmd *cobra.Command, args []string) {
	// Load configuration to get database path
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("❌ Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Create query engine
	queryEngine, err := query.NewQueryEngine(cfg.LogDBPath)
	if err != nil {
		fmt.Printf("❌ Failed to create query engine: %v\n", err)
		os.Exit(1)
	}

	// Parse flags
	since, _ := cmd.Flags().GetString("since")
	until, _ := cmd.Flags().GetString("until")
	model, _ := cmd.Flags().GetString("model")
	provider, _ := cmd.Flags().GetString("provider")
	language, _ := cmd.Flags().GetString("language")
	search, _ := cmd.Flags().GetString("search")
	limit, _ := cmd.Flags().GetInt("limit")
	stats, _ := cmd.Flags().GetBool("stats")
	format, _ := cmd.Flags().GetString("format")
	success, _ := cmd.Flags().GetString("success")

	// Build filters
	filters := query.QueryFilters{
		Language:  language,
		Model:     model,
		Provider:  provider,
		Search:    search,
		Limit:     limit,
		OrderBy:   "created",
		OrderDesc: true,
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
		filters.Since = sinceTime
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
		filters.Until = untilTime
	}

	// Parse success filter
	if success != "" {
		successBool, err := query.ParseBoolFilter(success)
		if err != nil {
			fmt.Printf("❌ Invalid success value: %v\n", err)
			if closeErr := queryEngine.Close(); closeErr != nil {
				fmt.Printf("Warning: failed to close query engine: %v\n", closeErr)
			}
			os.Exit(1)
		}
		filters.Success = successBool
	}

	// Show statistics if requested
	if stats {
		showStatistics(queryEngine, filters)
		return
	}

	// Query executions
	result, err := queryEngine.QueryExecutions(filters)
	if err != nil {
		fmt.Printf("❌ Query failed: %v\n", err)
		if closeErr := queryEngine.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close query engine: %v\n", closeErr)
		}
		os.Exit(1)
	}

	// Display results
	switch format {
	case "json":
		showJSONResults(result)
	case "csv":
		showCSVResults(result)
	case "table":
		fallthrough
	default:
		showTableResults(result)
	}

	// Clean up query engine
	if closeErr := queryEngine.Close(); closeErr != nil {
		fmt.Printf("Warning: failed to close query engine: %v\n", closeErr)
	}
}

func showStatistics(queryEngine *query.QueryEngine, filters query.QueryFilters) {
	stats, err := queryEngine.GenerateStatistics(filters)
	if err != nil {
		fmt.Printf("❌ Failed to generate statistics: %v\n", err)
		if closeErr := queryEngine.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close query engine: %v\n", closeErr)
		}
		os.Exit(1)
	}

	fmt.Println("📊 Execution Statistics")
	fmt.Println("=======================")
	fmt.Printf("Total Executions: %d\n", stats.TotalExecutions)
	fmt.Printf("Successful: %d\n", stats.SuccessfulRuns)
	fmt.Printf("Failed: %d\n", stats.FailedRuns)
	fmt.Printf("Success Rate: %.1f%%\n", stats.SuccessRate)
	fmt.Printf("Average Duration: %.2fms\n", stats.AverageDuration)
	fmt.Printf("Average Files: %.1f\n", stats.AverageFileCount)
	fmt.Println()

	if len(stats.LanguageBreakdown) > 0 {
		fmt.Println("Language Breakdown:")
		for lang, count := range stats.LanguageBreakdown {
			fmt.Printf("  %s: %d\n", lang, count)
		}
		fmt.Println()
	}

	if len(stats.ModelBreakdown) > 0 {
		fmt.Println("Model Usage:")
		for model, count := range stats.ModelBreakdown {
			fmt.Printf("  %s: %d\n", model, count)
		}
		fmt.Println()
	}

	if len(stats.DailyStats) > 0 {
		fmt.Println("Daily Usage (Last 10 days):")
		count := 0
		for day, usage := range stats.DailyStats {
			if count >= 10 {
				break
			}
			fmt.Printf("  %s: %d\n", day, usage)
			count++
		}
		fmt.Println()
	}
}

func showTableResults(result *query.QueryResult) {
	fmt.Printf("📋 Query Results (%d of %d total)\n", result.FilteredCount, result.TotalCount)
	fmt.Println(formatDivider(80))

	if len(result.Executions) == 0 {
		fmt.Println("No executions found matching the criteria.")
		return
	}

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tCreated\tLanguage\tModel\tDuration\tFiles\tSuccess")
	fmt.Fprintln(w, formatDivider(6)+"\t"+formatDivider(19)+"\t"+formatDivider(8)+"\t"+formatDivider(15)+"\t"+formatDivider(8)+"\t"+formatDivider(5)+"\t"+formatDivider(7))

	for _, exec := range result.Executions {
		successIcon := "✅"
		if !exec.Success {
			successIcon = "❌"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%dms\t%d\t%s\n",
			exec.ID[:8]+"...",
			exec.Created.Format("2006-01-02 15:04"),
			exec.DetectedLanguage,
			exec.ModelUsed,
			exec.ExecutionTimeMS,
			exec.FileCount,
			successIcon,
		)
	}
	_ = w.Flush() // Ignore error from table writer flush

	fmt.Printf("\nQuery completed in %v\n", result.QueryDuration)
}

func showJSONResults(result *query.QueryResult) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fmt.Printf("❌ Failed to encode JSON: %v\n", err)
		// Note: This function doesn't have access to queryEngine, so we can't close it here
		// The caller should handle cleanup
		os.Exit(1)
	}
}

func showCSVResults(result *query.QueryResult) {
	fmt.Println("ID,Created,Language,Model,Provider,Duration,Files,Success,Query")
	for _, exec := range result.Executions {
		fmt.Printf("%s,%s,%s,%s,%s,%d,%d,%t,\"%s\"\n",
			exec.ID,
			exec.Created.Format(time.RFC3339),
			exec.DetectedLanguage,
			exec.ModelUsed,
			exec.ProviderUsed,
			exec.ExecutionTimeMS,
			exec.FileCount,
			exec.Success,
			exec.WheresmypromptQuery,
		)
	}
}

func formatDivider(length int) string {
	divider := ""
	for i := 0; i < length; i++ {
		divider += "-"
	}
	return divider
}

func init() {
	// Add query flags
	queryCmd.Flags().String("since", "", "Show conversations since date (YYYY-MM-DD)")
	queryCmd.Flags().String("until", "", "Show conversations until date (YYYY-MM-DD)")
	queryCmd.Flags().String("model", "", "Filter by LLM model")
	queryCmd.Flags().String("provider", "", "Filter by LLM provider")
	queryCmd.Flags().String("language", "", "Filter by detected language")
	queryCmd.Flags().String("search", "", "Search in prompts and responses")
	queryCmd.Flags().String("success", "", "Filter by success status (true/false)")
	queryCmd.Flags().Int("limit", 10, "Maximum number of results")
	queryCmd.Flags().Bool("stats", false, "Show usage statistics")
	queryCmd.Flags().String("format", "table", "Output format (table, json, csv)")

	// Add to root command
	rootCmd.AddCommand(queryCmd)
}
