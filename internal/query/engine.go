package query

import (
	"fmt"
	"strconv"
	"time"

	"github.com/toozej/waffles/pkg/logging"
)

// QueryEngine handles complex queries against the logging database
type QueryEngine struct {
	db *logging.Database
}

// NewQueryEngine creates a new query engine
func NewQueryEngine(dbPath string) (*QueryEngine, error) {
	db, err := logging.NewDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &QueryEngine{
		db: db,
	}, nil
}

// Close closes the database connection
func (qe *QueryEngine) Close() error {
	return qe.db.Close()
}

// QueryFilters represents filters for querying executions
type QueryFilters struct {
	Since        *time.Time `json:"since,omitempty"`
	Until        *time.Time `json:"until,omitempty"`
	Model        string     `json:"model,omitempty"`
	Provider     string     `json:"provider,omitempty"`
	Language     string     `json:"language,omitempty"`
	Success      *bool      `json:"success,omitempty"`
	MinDuration  *int64     `json:"min_duration,omitempty"`
	MaxDuration  *int64     `json:"max_duration,omitempty"`
	MinFileCount *int       `json:"min_file_count,omitempty"`
	MaxFileCount *int       `json:"max_file_count,omitempty"`
	Search       string     `json:"search,omitempty"`
	Limit        int        `json:"limit"`
	Offset       int        `json:"offset"`
	OrderBy      string     `json:"order_by"`
	OrderDesc    bool       `json:"order_desc"`
}

// QueryResult represents the results of a query
type QueryResult struct {
	Executions    []logging.WafflesExecution `json:"executions"`
	TotalCount    int                        `json:"total_count"`
	FilteredCount int                        `json:"filtered_count"`
	Statistics    *ExecutionStats            `json:"statistics,omitempty"`
	QueryDuration time.Duration              `json:"query_duration"`
}

// ExecutionStats provides statistical analysis of executions
type ExecutionStats struct {
	TotalExecutions   int            `json:"total_executions"`
	SuccessfulRuns    int            `json:"successful_runs"`
	FailedRuns        int            `json:"failed_runs"`
	SuccessRate       float64        `json:"success_rate"`
	AverageFileCount  float64        `json:"average_file_count"`
	AverageDuration   float64        `json:"average_duration_ms"`
	TotalDuration     int64          `json:"total_duration_ms"`
	LanguageBreakdown map[string]int `json:"language_breakdown"`
	ModelBreakdown    map[string]int `json:"model_breakdown"`
	ProviderBreakdown map[string]int `json:"provider_breakdown"`
	DailyStats        map[string]int `json:"daily_stats"`
	HourlyStats       map[int]int    `json:"hourly_stats"`
	FileCountBuckets  map[string]int `json:"file_count_buckets"`
	DurationBuckets   map[string]int `json:"duration_buckets"`
}

// QueryExecutions performs a filtered query for executions
func (qe *QueryEngine) QueryExecutions(filters QueryFilters) (*QueryResult, error) {
	startTime := time.Now()

	// Convert filters to ExecutionFilter
	execFilter := qe.convertToExecutionFilter(filters)

	// Execute the query
	executions, err := qe.db.QueryExecutions(execFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to query executions: %w", err)
	}

	// Get total count
	totalCount, err := qe.db.CountExecutions(execFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	result := &QueryResult{
		Executions:    executions,
		TotalCount:    totalCount,
		FilteredCount: len(executions),
		QueryDuration: time.Since(startTime),
	}

	return result, nil
}

// QueryExecutionsWithStats performs a query and includes statistical analysis
func (qe *QueryEngine) QueryExecutionsWithStats(filters QueryFilters) (*QueryResult, error) {
	result, err := qe.QueryExecutions(filters)
	if err != nil {
		return nil, err
	}

	// Generate statistics
	stats, err := qe.GenerateStatistics(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to generate statistics: %w", err)
	}

	result.Statistics = stats
	return result, nil
}

// GenerateStatistics generates comprehensive statistics for executions
func (qe *QueryEngine) GenerateStatistics(filters QueryFilters) (*ExecutionStats, error) {
	// Convert filters to ExecutionFilter for stats
	execFilter := qe.convertToExecutionFilter(filters)
	execFilter.Limit = 0 // No limit for stats
	execFilter.Offset = 0

	// Use existing GetExecutionStats method
	existingStats, err := qe.db.GetExecutionStats(execFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution stats: %w", err)
	}

	// Convert to our stats format
	stats := &ExecutionStats{
		TotalExecutions:   existingStats.TotalExecutions,
		SuccessfulRuns:    existingStats.SuccessfulExecutions,
		FailedRuns:        existingStats.FailedExecutions,
		SuccessRate:       float64(existingStats.SuccessfulExecutions) / float64(existingStats.TotalExecutions) * 100,
		AverageFileCount:  float64(existingStats.TotalFiles) / float64(existingStats.TotalExecutions),
		AverageDuration:   existingStats.AverageExecutionTime,
		TotalDuration:     int64(existingStats.AverageExecutionTime * float64(existingStats.TotalExecutions)),
		LanguageBreakdown: existingStats.LanguageBreakdown,
		ModelBreakdown:    existingStats.ModelBreakdown,
		ProviderBreakdown: existingStats.ProviderBreakdown,
		DailyStats:        make(map[string]int),
		HourlyStats:       make(map[int]int),
		FileCountBuckets:  make(map[string]int),
		DurationBuckets:   make(map[string]int),
	}

	// Get additional detailed stats by querying executions
	executions, err := qe.db.QueryExecutions(execFilter)
	if err != nil {
		return stats, nil // Return basic stats if detailed query fails
	}

	// Add time-based and bucket breakdowns
	for _, exec := range executions {
		// Daily stats
		dayKey := exec.Created.Format("2006-01-02")
		stats.DailyStats[dayKey]++

		// Hourly stats
		stats.HourlyStats[exec.Created.Hour()]++

		// File count buckets
		fileCountBucket := qe.getFileCountBucket(exec.FileCount)
		stats.FileCountBuckets[fileCountBucket]++

		// Duration buckets
		durationBucket := qe.getDurationBucket(exec.ExecutionTimeMS)
		stats.DurationBuckets[durationBucket]++
	}

	return stats, nil
}

// SearchFullText performs full-text search across prompts and responses
func (qe *QueryEngine) SearchFullText(searchTerm string, filters QueryFilters) (*QueryResult, error) {
	filters.Search = searchTerm
	return qe.QueryExecutions(filters)
}

// GetUsagePatterns analyzes usage patterns over time
func (qe *QueryEngine) GetUsagePatterns(days int) (map[string]interface{}, error) {
	since := time.Now().AddDate(0, 0, -days)
	filters := QueryFilters{
		Since: &since,
		Limit: 0, // No limit
	}

	stats, err := qe.GenerateStatistics(filters)
	if err != nil {
		return nil, err
	}

	patterns := make(map[string]interface{})
	patterns["daily_usage"] = stats.DailyStats
	patterns["hourly_usage"] = stats.HourlyStats
	patterns["language_preference"] = stats.LanguageBreakdown
	patterns["model_usage"] = stats.ModelBreakdown
	patterns["success_rate"] = stats.SuccessRate
	patterns["average_files"] = stats.AverageFileCount

	return patterns, nil
}

// convertToExecutionFilter converts QueryFilters to logging.ExecutionFilter
func (qe *QueryEngine) convertToExecutionFilter(filters QueryFilters) *logging.ExecutionFilter {
	execFilter := &logging.ExecutionFilter{
		Limit:  filters.Limit,
		Offset: filters.Offset,
	}

	if filters.Since != nil {
		execFilter.DateFrom = filters.Since
	}

	if filters.Until != nil {
		execFilter.DateTo = filters.Until
	}

	if filters.Language != "" {
		execFilter.Language = filters.Language
	}

	if filters.Model != "" {
		execFilter.Model = filters.Model
	}

	if filters.Provider != "" {
		execFilter.Provider = filters.Provider
	}

	if filters.Success != nil {
		execFilter.Success = filters.Success
	}

	if filters.MinDuration != nil {
		execFilter.MinDuration = filters.MinDuration
	}

	if filters.MaxDuration != nil {
		execFilter.MaxDuration = filters.MaxDuration
	}

	if filters.Search != "" {
		execFilter.SearchQuery = filters.Search
	}

	return execFilter
}

// getFileCountBucket categorizes file counts into buckets
func (qe *QueryEngine) getFileCountBucket(fileCount int) string {
	switch {
	case fileCount == 0:
		return "0"
	case fileCount <= 5:
		return "1-5"
	case fileCount <= 10:
		return "6-10"
	case fileCount <= 25:
		return "11-25"
	case fileCount <= 50:
		return "26-50"
	case fileCount <= 100:
		return "51-100"
	default:
		return "100+"
	}
}

// getDurationBucket categorizes execution durations into buckets
func (qe *QueryEngine) getDurationBucket(durationMS int64) string {
	seconds := durationMS / 1000
	switch {
	case seconds < 1:
		return "<1s"
	case seconds <= 5:
		return "1-5s"
	case seconds <= 15:
		return "6-15s"
	case seconds <= 30:
		return "16-30s"
	case seconds <= 60:
		return "31-60s"
	case seconds <= 120:
		return "1-2min"
	case seconds <= 300:
		return "2-5min"
	default:
		return "5min+"
	}
}

// ParseDateFilter parses date strings into time.Time
func ParseDateFilter(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}

	// Try different date formats
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("unable to parse date: %s", dateStr)
}

// ParseBoolFilter parses string bool filters
func ParseBoolFilter(boolStr string) (*bool, error) {
	if boolStr == "" {
		return nil, nil
	}

	val, err := strconv.ParseBool(boolStr)
	if err != nil {
		return nil, fmt.Errorf("invalid boolean value: %s", boolStr)
	}

	return &val, nil
}

// ParseIntFilter parses string int filters
func ParseIntFilter(intStr string) (*int, error) {
	if intStr == "" {
		return nil, nil
	}

	val, err := strconv.Atoi(intStr)
	if err != nil {
		return nil, fmt.Errorf("invalid integer value: %s", intStr)
	}

	return &val, nil
}

// ParseInt64Filter parses string int64 filters
func ParseInt64Filter(intStr string) (*int64, error) {
	if intStr == "" {
		return nil, nil
	}

	val, err := strconv.ParseInt(intStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid integer value: %s", intStr)
	}

	return &val, nil
}
