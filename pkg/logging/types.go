package logging

import (
	"database/sql"
	"time"
)

// Database represents the SQLite logging database
type Database struct {
	path string
	db   *sql.DB
}

// WafflesExecution represents a complete Waffles pipeline execution
type WafflesExecution struct {
	ID                  string    `json:"id"`
	ConversationID      string    `json:"conversation_id"`
	CommandArgs         string    `json:"command_args"`
	WheresmypromptQuery string    `json:"wheresmyprompt_query"`
	Files2promptArgs    string    `json:"files2prompt_args"`
	LLMArgs             string    `json:"llm_args"`
	DetectedLanguage    string    `json:"detected_language"`
	FileCount           int       `json:"file_count"`
	ExecutionTimeMS     int64     `json:"execution_time_ms"`
	Success             bool      `json:"success"`
	ErrorMessage        string    `json:"error_message,omitempty"`
	ModelUsed           string    `json:"model_used"`
	ProviderUsed        string    `json:"provider_used"`
	Created             time.Time `json:"created"`
	Updated             time.Time `json:"updated"`
}

// WafflesFile represents a file that was processed during execution
type WafflesFile struct {
	ID              string    `json:"id"`
	ExecutionID     string    `json:"execution_id"`
	FilePath        string    `json:"file_path"`
	FileSize        int64     `json:"file_size"`
	Included        bool      `json:"included"`
	ExclusionReason string    `json:"exclusion_reason,omitempty"`
	Created         time.Time `json:"created"`
}

// WafflesStep represents individual tool execution within a pipeline
type WafflesStep struct {
	ID          string    `json:"id"`
	ExecutionID string    `json:"execution_id"`
	Tool        string    `json:"tool"`
	Command     string    `json:"command"`
	Output      string    `json:"output"`
	ErrorOutput string    `json:"error_output,omitempty"`
	Success     bool      `json:"success"`
	DurationMS  int64     `json:"duration_ms"`
	StepOrder   int       `json:"step_order"`
	Created     time.Time `json:"created"`
}

// ExecutionFilter represents filtering criteria for querying executions
type ExecutionFilter struct {
	DateFrom    *time.Time `json:"date_from,omitempty"`
	DateTo      *time.Time `json:"date_to,omitempty"`
	Language    string     `json:"language,omitempty"`
	Model       string     `json:"model,omitempty"`
	Provider    string     `json:"provider,omitempty"`
	Success     *bool      `json:"success,omitempty"`
	MinDuration *int64     `json:"min_duration,omitempty"`
	MaxDuration *int64     `json:"max_duration,omitempty"`
	SearchQuery string     `json:"search_query,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
}

// ExecutionStats represents statistics about executions
type ExecutionStats struct {
	TotalExecutions      int            `json:"total_executions"`
	SuccessfulExecutions int            `json:"successful_executions"`
	FailedExecutions     int            `json:"failed_executions"`
	AverageExecutionTime float64        `json:"average_execution_time_ms"`
	TotalFiles           int            `json:"total_files"`
	LanguageBreakdown    map[string]int `json:"language_breakdown"`
	ModelBreakdown       map[string]int `json:"model_breakdown"`
	ProviderBreakdown    map[string]int `json:"provider_breakdown"`
}

// ExportFormat represents supported export formats
type ExportFormat string

const (
	ExportFormatJSON     ExportFormat = "json"
	ExportFormatCSV      ExportFormat = "csv"
	ExportFormatMarkdown ExportFormat = "markdown"
	ExportFormatSQL      ExportFormat = "sql"
)

// ExportOptions represents options for exporting data
type ExportOptions struct {
	Format       ExportFormat     `json:"format"`
	Filter       *ExecutionFilter `json:"filter,omitempty"`
	IncludeFiles bool             `json:"include_files"`
	IncludeSteps bool             `json:"include_steps"`
	Compress     bool             `json:"compress"`
	Template     string           `json:"template,omitempty"`
}
