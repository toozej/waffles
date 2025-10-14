package export

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/toozej/waffles/internal/query"
	"github.com/toozej/waffles/pkg/logging"
)

// Exporter handles the export process with progress indication and streaming
type Exporter struct {
	queryEngine *query.QueryEngine
	options     *ExportOptions
	progress    ProgressReporter
}

// ProgressReporter interface for progress reporting during exports
type ProgressReporter interface {
	Start(total int)
	Update(current int, message string)
	Finish(success bool, message string)
}

// ConsoleProgressReporter implements progress reporting to console
type ConsoleProgressReporter struct {
	startTime time.Time
	total     int
}

// Start begins progress reporting
func (p *ConsoleProgressReporter) Start(total int) {
	p.startTime = time.Now()
	p.total = total
	fmt.Printf("Starting export of %d records...\n", total)
}

// Update reports progress
func (p *ConsoleProgressReporter) Update(current int, message string) {
	if p.total > 0 {
		percentage := float64(current) / float64(p.total) * 100
		elapsed := time.Since(p.startTime)
		fmt.Printf("Progress: %d/%d (%.1f%%) - %s - Elapsed: %v\n", current, p.total, percentage, message, elapsed.Round(time.Second))
	} else {
		fmt.Printf("Progress: %d records - %s\n", current, message)
	}
}

// Finish completes progress reporting
func (p *ConsoleProgressReporter) Finish(success bool, message string) {
	elapsed := time.Since(p.startTime)
	if success {
		fmt.Printf("✅ Export completed successfully: %s (Duration: %v)\n", message, elapsed.Round(time.Second))
	} else {
		fmt.Printf("❌ Export failed: %s (Duration: %v)\n", message, elapsed.Round(time.Second))
	}
}

// SilentProgressReporter implements silent progress reporting
type SilentProgressReporter struct{}

func (p *SilentProgressReporter) Start(total int)                     {}
func (p *SilentProgressReporter) Update(current int, message string)  {}
func (p *SilentProgressReporter) Finish(success bool, message string) {}

// NewExporter creates a new exporter
func NewExporter(queryEngine *query.QueryEngine, options *ExportOptions) *Exporter {
	return &Exporter{
		queryEngine: queryEngine,
		options:     options,
		progress:    &ConsoleProgressReporter{},
	}
}

// SetProgressReporter sets a custom progress reporter
func (e *Exporter) SetProgressReporter(reporter ProgressReporter) {
	e.progress = reporter
}

// Export performs the complete export process
func (e *Exporter) Export(writer io.Writer) error {
	// Prepare the writer (add compression if requested)
	finalWriter, cleanup, err := e.prepareWriter(writer)
	if err != nil {
		return fmt.Errorf("failed to prepare writer: %w", err)
	}
	defer cleanup()

	// Get total count for progress reporting
	totalCount, err := e.getTotalCount()
	if err != nil {
		return fmt.Errorf("failed to get total count: %w", err)
	}

	e.progress.Start(totalCount)

	// Prepare export data
	exportData, err := e.prepareExportData()
	if err != nil {
		e.progress.Finish(false, fmt.Sprintf("Failed to prepare data: %v", err))
		return fmt.Errorf("failed to prepare export data: %w", err)
	}

	e.progress.Update(len(exportData.Executions), "Data prepared, starting format conversion")

	// Format and write data
	if err := e.formatAndWrite(exportData, finalWriter); err != nil {
		e.progress.Finish(false, fmt.Sprintf("Failed to write data: %v", err))
		return fmt.Errorf("failed to format and write data: %w", err)
	}

	e.progress.Finish(true, fmt.Sprintf("Exported %d records in %s format", len(exportData.Executions), e.options.Format))
	return nil
}

// ExportToFile exports data directly to a file
func (e *Exporter) ExportToFile(filename string) error {
	file, err := os.Create(filename) // #nosec G304 -- Filename from user-specified export path
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	return e.Export(file)
}

// prepareWriter sets up the final writer with optional compression
func (e *Exporter) prepareWriter(writer io.Writer) (io.Writer, func(), error) {
	if e.options.Compress {
		gzipWriter := gzip.NewWriter(writer)
		return gzipWriter, func() { _ = gzipWriter.Close() }, nil
	}
	return writer, func() {}, nil
}

// getTotalCount gets the total number of records that will be exported
func (e *Exporter) getTotalCount() (int, error) {
	filters := e.convertToQueryFilters()
	result, err := e.queryEngine.QueryExecutions(filters)
	if err != nil {
		return 0, err
	}
	return result.TotalCount, nil
}

// prepareExportData prepares all export data
func (e *Exporter) prepareExportData() (*ExportData, error) {
	// Query executions
	filters := e.convertToQueryFilters()

	var result *query.QueryResult
	var err error

	if e.options.IncludeStats {
		result, err = e.queryEngine.QueryExecutionsWithStats(filters)
	} else {
		result, err = e.queryEngine.QueryExecutions(filters)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query executions: %w", err)
	}

	// Prepare export data structure
	exportData := &ExportData{
		Metadata: ExportMetadata{
			ExportedAt:    time.Now(),
			Format:        e.options.Format,
			RecordCount:   len(result.Executions),
			FilterApplied: e.options.Filter != nil,
			Version:       "1.0",
		},
		Executions: result.Executions,
		Statistics: e.convertStats(result.Statistics),
	}

	// Note: Files and Steps would be included here if the database methods were available
	// For now, we'll just include the basic execution data

	return exportData, nil
}

// formatAndWrite formats the data and writes it to the writer
func (e *Exporter) formatAndWrite(data *ExportData, writer io.Writer) error {
	switch e.options.Format {
	case FormatJSON:
		formatter := &JSONFormatter{Pretty: true}
		return formatter.FormatJSON(data, writer)

	case FormatCSV:
		formatter := &CSVFormatter{IncludeHeaders: true}
		return formatter.FormatCSV(data, writer)

	case FormatMarkdown:
		formatter := &MarkdownFormatter{
			IncludeStats: e.options.IncludeStats,
			IncludeTOC:   true,
		}
		return formatter.FormatMarkdown(data, writer)

	case FormatSQL:
		formatter := &SQLFormatter{
			IncludeSchema: true,
			BatchSize:     100,
		}
		return formatter.FormatSQL(data, writer)

	case FormatTemplate:
		if e.options.Template == "" {
			return fmt.Errorf("template is required for template format")
		}
		formatter := &TemplateFormatter{Template: e.options.Template}
		return formatter.FormatTemplate(data, writer)

	default:
		return fmt.Errorf("unsupported export format: %s", e.options.Format)
	}
}

// convertToQueryFilters converts export options to query filters
func (e *Exporter) convertToQueryFilters() query.QueryFilters {
	filters := query.QueryFilters{}

	if e.options.Filter != nil {
		if e.options.Filter.DateFrom != nil {
			filters.Since = e.options.Filter.DateFrom
		}
		if e.options.Filter.DateTo != nil {
			filters.Until = e.options.Filter.DateTo
		}
		filters.Language = e.options.Filter.Language
		filters.Model = e.options.Filter.Model
		filters.Provider = e.options.Filter.Provider
		filters.Success = e.options.Filter.Success
		filters.MinDuration = e.options.Filter.MinDuration
		filters.MaxDuration = e.options.Filter.MaxDuration
		filters.Search = e.options.Filter.SearchQuery
		filters.Limit = e.options.Filter.Limit
		filters.Offset = e.options.Filter.Offset
	}

	return filters
}

// convertStats converts query.ExecutionStats to logging.ExecutionStats
func (e *Exporter) convertStats(queryStats *query.ExecutionStats) *logging.ExecutionStats {
	if queryStats == nil {
		return nil
	}

	return &logging.ExecutionStats{
		TotalExecutions:      queryStats.TotalExecutions,
		SuccessfulExecutions: queryStats.SuccessfulRuns,
		FailedExecutions:     queryStats.FailedRuns,
		AverageExecutionTime: queryStats.AverageDuration,
		TotalFiles:           int(queryStats.AverageFileCount * float64(max(queryStats.TotalExecutions, 1))),
		LanguageBreakdown:    queryStats.LanguageBreakdown,
		ModelBreakdown:       queryStats.ModelBreakdown,
		ProviderBreakdown:    queryStats.ProviderBreakdown,
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
