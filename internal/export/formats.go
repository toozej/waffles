package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/toozej/waffles/pkg/logging"
)

// ExportFormat represents different export formats
type ExportFormat string

const (
	FormatJSON     ExportFormat = "json"
	FormatCSV      ExportFormat = "csv"
	FormatMarkdown ExportFormat = "markdown"
	FormatSQL      ExportFormat = "sql"
	FormatTemplate ExportFormat = "template"
)

// ExportOptions contains options for exporting data
type ExportOptions struct {
	Format       ExportFormat             `json:"format"`
	Filter       *logging.ExecutionFilter `json:"filter,omitempty"`
	IncludeFiles bool                     `json:"include_files"`
	IncludeSteps bool                     `json:"include_steps"`
	IncludeStats bool                     `json:"include_stats"`
	Compress     bool                     `json:"compress"`
	Template     string                   `json:"template,omitempty"`
	TemplateData map[string]interface{}   `json:"template_data,omitempty"`
}

// ExportData represents the complete export data structure
type ExportData struct {
	Metadata   ExportMetadata                   `json:"metadata"`
	Executions []logging.WafflesExecution       `json:"executions"`
	Files      map[string][]logging.WafflesFile `json:"files,omitempty"`
	Steps      map[string][]logging.WafflesStep `json:"steps,omitempty"`
	Statistics *logging.ExecutionStats          `json:"statistics,omitempty"`
}

// ExportMetadata contains information about the export
type ExportMetadata struct {
	ExportedAt    time.Time    `json:"exported_at"`
	Format        ExportFormat `json:"format"`
	RecordCount   int          `json:"record_count"`
	FilterApplied bool         `json:"filter_applied"`
	Version       string       `json:"version"`
}

// JSONFormatter handles JSON export formatting
type JSONFormatter struct {
	Pretty bool
}

// FormatJSON exports data in JSON format
func (f *JSONFormatter) FormatJSON(data *ExportData, writer io.Writer) error {
	encoder := json.NewEncoder(writer)

	if f.Pretty {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(data)
}

// CSVFormatter handles CSV export formatting
type CSVFormatter struct {
	IncludeHeaders bool
	CustomColumns  []string
}

// FormatCSV exports executions in CSV format
func (f *CSVFormatter) FormatCSV(data *ExportData, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Define columns
	columns := []string{
		"ID", "ConversationID", "CommandArgs", "WheresmypromptQuery",
		"Files2promptArgs", "LLMArgs", "DetectedLanguage", "FileCount",
		"ExecutionTimeMS", "Success", "ErrorMessage", "ModelUsed",
		"ProviderUsed", "Created", "Updated",
	}

	// Use custom columns if provided
	if len(f.CustomColumns) > 0 {
		columns = f.CustomColumns
	}

	// Write header
	if f.IncludeHeaders {
		if err := csvWriter.Write(columns); err != nil {
			return fmt.Errorf("failed to write CSV header: %w", err)
		}
	}

	// Write data rows
	for _, exec := range data.Executions {
		record := []string{
			exec.ID, exec.ConversationID, exec.CommandArgs, exec.WheresmypromptQuery,
			exec.Files2promptArgs, exec.LLMArgs, exec.DetectedLanguage, fmt.Sprintf("%d", exec.FileCount),
			fmt.Sprintf("%d", exec.ExecutionTimeMS), fmt.Sprintf("%t", exec.Success), exec.ErrorMessage, exec.ModelUsed,
			exec.ProviderUsed, exec.Created.Format(time.RFC3339), exec.Updated.Format(time.RFC3339),
		}

		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

// MarkdownFormatter handles Markdown export formatting
type MarkdownFormatter struct {
	Title        string
	IncludeStats bool
	IncludeTOC   bool
}

// FormatMarkdown exports data in Markdown format
func (f *MarkdownFormatter) FormatMarkdown(data *ExportData, writer io.Writer) error {
	title := f.Title
	if title == "" {
		title = "Waffles Execution Report"
	}

	// Write title and metadata
	fmt.Fprintf(writer, "# %s\n\n", title)
	fmt.Fprintf(writer, "**Generated**: %s  \n", data.Metadata.ExportedAt.Format(time.RFC3339))
	fmt.Fprintf(writer, "**Total Records**: %d  \n", data.Metadata.RecordCount)
	fmt.Fprintf(writer, "**Format**: %s  \n\n", data.Metadata.Format)

	// Write table of contents
	if f.IncludeTOC {
		fmt.Fprintf(writer, "## Table of Contents\n\n")
		if f.IncludeStats && data.Statistics != nil {
			fmt.Fprintf(writer, "- [Statistics](#statistics)\n")
		}
		fmt.Fprintf(writer, "- [Executions](#executions)\n\n")
	}

	// Write statistics section
	if f.IncludeStats && data.Statistics != nil {
		if err := f.writeStatisticsSection(data.Statistics, writer); err != nil {
			return fmt.Errorf("failed to write statistics section: %w", err)
		}
	}

	// Write executions section
	fmt.Fprintf(writer, "## Executions\n\n")
	for i, exec := range data.Executions {
		fmt.Fprintf(writer, "### %d. Execution %s\n\n", i+1, exec.ID[:8])

		fmt.Fprintf(writer, "| Field | Value |\n")
		fmt.Fprintf(writer, "|-------|-------|\n")
		fmt.Fprintf(writer, "| **Created** | %s |\n", exec.Created.Format(time.RFC3339))
		fmt.Fprintf(writer, "| **Language** | %s |\n", exec.DetectedLanguage)
		fmt.Fprintf(writer, "| **Model** | %s |\n", exec.ModelUsed)
		fmt.Fprintf(writer, "| **Provider** | %s |\n", exec.ProviderUsed)
		fmt.Fprintf(writer, "| **Success** | %t |\n", exec.Success)
		fmt.Fprintf(writer, "| **Duration** | %dms |\n", exec.ExecutionTimeMS)
		fmt.Fprintf(writer, "| **File Count** | %d |\n", exec.FileCount)

		if exec.WheresmypromptQuery != "" {
			fmt.Fprintf(writer, "| **Query** | %s |\n", exec.WheresmypromptQuery)
		}

		if exec.ErrorMessage != "" {
			fmt.Fprintf(writer, "| **Error** | %s |\n", exec.ErrorMessage)
		}

		fmt.Fprintf(writer, "\n")

		// Include files if available
		if files, exists := data.Files[exec.ID]; exists && len(files) > 0 {
			fmt.Fprintf(writer, "#### Files (%d)\n\n", len(files))
			for _, file := range files {
				status := "✅"
				if !file.Included {
					status = "❌"
				}
				fmt.Fprintf(writer, "- %s %s", status, file.FilePath)
				if file.ExclusionReason != "" {
					fmt.Fprintf(writer, " *(excluded: %s)*", file.ExclusionReason)
				}
				fmt.Fprintf(writer, "\n")
			}
			fmt.Fprintf(writer, "\n")
		}

		// Include steps if available
		if steps, exists := data.Steps[exec.ID]; exists && len(steps) > 0 {
			fmt.Fprintf(writer, "#### Pipeline Steps\n\n")
			for _, step := range steps {
				status := "✅"
				if !step.Success {
					status = "❌"
				}
				fmt.Fprintf(writer, "%d. %s **%s** (%dms)\n", step.StepOrder, status, step.Tool, step.DurationMS)
				if step.ErrorOutput != "" {
					fmt.Fprintf(writer, "   - Error: %s\n", step.ErrorOutput)
				}
			}
			fmt.Fprintf(writer, "\n")
		}
	}

	return nil
}

// writeStatisticsSection writes the statistics section to markdown
func (f *MarkdownFormatter) writeStatisticsSection(stats *logging.ExecutionStats, writer io.Writer) error {
	fmt.Fprintf(writer, "## Statistics\n\n")

	fmt.Fprintf(writer, "### Overview\n\n")
	fmt.Fprintf(writer, "| Metric | Value |\n")
	fmt.Fprintf(writer, "|--------|-------|\n")
	fmt.Fprintf(writer, "| **Total Executions** | %d |\n", stats.TotalExecutions)
	fmt.Fprintf(writer, "| **Successful** | %d |\n", stats.SuccessfulExecutions)
	fmt.Fprintf(writer, "| **Failed** | %d |\n", stats.FailedExecutions)
	fmt.Fprintf(writer, "| **Average Duration** | %.2fms |\n", stats.AverageExecutionTime)
	fmt.Fprintf(writer, "| **Total Files** | %d |\n", stats.TotalFiles)
	fmt.Fprintf(writer, "\n")

	// Language breakdown
	// Create title caser once for reuse
	caser := cases.Title(language.English)

	if len(stats.LanguageBreakdown) > 0 {
		fmt.Fprintf(writer, "### Language Breakdown\n\n")
		for lang, count := range stats.LanguageBreakdown {
			fmt.Fprintf(writer, "- **%s**: %d\n", caser.String(lang), count)
		}
		fmt.Fprintf(writer, "\n")
	}

	// Model breakdown
	if len(stats.ModelBreakdown) > 0 {
		fmt.Fprintf(writer, "### Model Usage\n\n")
		for model, count := range stats.ModelBreakdown {
			fmt.Fprintf(writer, "- **%s**: %d\n", model, count)
		}
		fmt.Fprintf(writer, "\n")
	}

	// Provider breakdown
	if len(stats.ProviderBreakdown) > 0 {
		fmt.Fprintf(writer, "### Provider Usage\n\n")
		for provider, count := range stats.ProviderBreakdown {
			fmt.Fprintf(writer, "- **%s**: %d\n", caser.String(provider), count)
		}
		fmt.Fprintf(writer, "\n")
	}

	return nil
}

// SQLFormatter handles SQL export formatting
type SQLFormatter struct {
	TableName     string
	IncludeSchema bool
	BatchSize     int
}

// FormatSQL exports data as SQL INSERT statements
func (f *SQLFormatter) FormatSQL(data *ExportData, writer io.Writer) error {
	tableName := f.TableName
	if tableName == "" {
		tableName = "waffles_executions"
	}

	// Write header
	fmt.Fprintf(writer, "-- Waffles Execution Data Export\n")
	fmt.Fprintf(writer, "-- Generated: %s\n", data.Metadata.ExportedAt.Format(time.RFC3339))
	fmt.Fprintf(writer, "-- Records: %d\n\n", data.Metadata.RecordCount)

	// Write schema if requested
	if f.IncludeSchema {
		if err := f.writeSchema(writer); err != nil {
			return fmt.Errorf("failed to write schema: %w", err)
		}
	}

	// Write data
	batchSize := f.BatchSize
	if batchSize == 0 {
		batchSize = 100
	}

	for i, exec := range data.Executions {
		if i%batchSize == 0 && i > 0 {
			fmt.Fprintf(writer, "\n-- Batch %d\n", i/batchSize+1)
		}

		fmt.Fprintf(writer,
			"INSERT INTO %s (id, conversation_id, command_args, wheresmyprompt_query, files2prompt_args, llm_args, detected_language, file_count, execution_time_ms, success, error_message, model_used, provider_used, created, updated) VALUES ('%s', %s, %s, %s, %s, %s, '%s', %d, %d, %t, %s, '%s', '%s', '%s', '%s');\n",
			tableName,
			exec.ID,
			f.sqlQuote(exec.ConversationID),
			f.sqlQuote(exec.CommandArgs),
			f.sqlQuote(exec.WheresmypromptQuery),
			f.sqlQuote(exec.Files2promptArgs),
			f.sqlQuote(exec.LLMArgs),
			exec.DetectedLanguage,
			exec.FileCount,
			exec.ExecutionTimeMS,
			exec.Success,
			f.sqlQuote(exec.ErrorMessage),
			exec.ModelUsed,
			exec.ProviderUsed,
			exec.Created.Format(time.RFC3339),
			exec.Updated.Format(time.RFC3339),
		)
	}

	return nil
}

// writeSchema writes the database schema
func (f *SQLFormatter) writeSchema(writer io.Writer) error {
	schema := `-- Table schema
CREATE TABLE IF NOT EXISTS waffles_executions (
    id TEXT PRIMARY KEY,
    conversation_id TEXT,
    command_args TEXT,
    wheresmyprompt_query TEXT,
    files2prompt_args TEXT,
    llm_args TEXT,
    detected_language TEXT,
    file_count INTEGER,
    execution_time_ms INTEGER,
    success BOOLEAN,
    error_message TEXT,
    model_used TEXT,
    provider_used TEXT,
    created DATETIME,
    updated DATETIME
);

-- Data
`
	_, err := writer.Write([]byte(schema))
	return err
}

// sqlQuote properly quotes SQL string values
func (f *SQLFormatter) sqlQuote(value string) string {
	if value == "" {
		return "NULL"
	}
	// Escape single quotes
	escaped := strings.ReplaceAll(value, "'", "''")
	return fmt.Sprintf("'%s'", escaped)
}

// TemplateFormatter handles custom template formatting
type TemplateFormatter struct {
	Template string
	FuncMap  template.FuncMap
}

// FormatTemplate exports data using a custom Go template
func (f *TemplateFormatter) FormatTemplate(data *ExportData, writer io.Writer) error {
	if f.Template == "" {
		return fmt.Errorf("template is required")
	}

	// Create template with helper functions
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
		"duration": func(ms int64) string {
			return fmt.Sprintf("%dms", ms)
		},
		"success": func(success bool) string {
			if success {
				return "✅"
			}
			return "❌"
		},
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		"title": func(s string) string {
			caser := cases.Title(language.English)
			return caser.String(s)
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}

	// Merge with custom functions
	if f.FuncMap != nil {
		for k, v := range f.FuncMap {
			funcMap[k] = v
		}
	}

	tmpl, err := template.New("export").Funcs(funcMap).Parse(f.Template)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl.Execute(writer, data)
}

// GetFormatterForFormat returns the appropriate formatter for the given format
func GetFormatterForFormat(format ExportFormat, options map[string]interface{}) (interface{}, error) {
	switch format {
	case FormatJSON:
		formatter := &JSONFormatter{
			Pretty: true, // Default to pretty formatting
		}
		if pretty, ok := options["pretty"].(bool); ok {
			formatter.Pretty = pretty
		}
		return formatter, nil

	case FormatCSV:
		formatter := &CSVFormatter{
			IncludeHeaders: true, // Default to including headers
		}
		if headers, ok := options["include_headers"].(bool); ok {
			formatter.IncludeHeaders = headers
		}
		if columns, ok := options["columns"].([]string); ok {
			formatter.CustomColumns = columns
		}
		return formatter, nil

	case FormatMarkdown:
		formatter := &MarkdownFormatter{
			IncludeStats: true,
			IncludeTOC:   true,
		}
		if title, ok := options["title"].(string); ok {
			formatter.Title = title
		}
		if stats, ok := options["include_stats"].(bool); ok {
			formatter.IncludeStats = stats
		}
		if toc, ok := options["include_toc"].(bool); ok {
			formatter.IncludeTOC = toc
		}
		return formatter, nil

	case FormatSQL:
		formatter := &SQLFormatter{
			TableName:     "waffles_executions",
			IncludeSchema: false,
			BatchSize:     100,
		}
		if table, ok := options["table_name"].(string); ok {
			formatter.TableName = table
		}
		if schema, ok := options["include_schema"].(bool); ok {
			formatter.IncludeSchema = schema
		}
		if batch, ok := options["batch_size"].(int); ok {
			formatter.BatchSize = batch
		}
		return formatter, nil

	case FormatTemplate:
		templateStr, ok := options["template"].(string)
		if !ok || templateStr == "" {
			return nil, fmt.Errorf("template is required for template format")
		}
		formatter := &TemplateFormatter{
			Template: templateStr,
		}
		if funcMap, ok := options["func_map"].(template.FuncMap); ok {
			formatter.FuncMap = funcMap
		}
		return formatter, nil

	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}
