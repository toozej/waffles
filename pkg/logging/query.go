package logging

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// QueryExecutions retrieves executions based on filter criteria
func (d *Database) QueryExecutions(filter *ExecutionFilter) ([]WafflesExecution, error) {
	if filter == nil {
		filter = &ExecutionFilter{}
	}

	// Build WHERE clause
	var conditions []string
	var args []interface{}

	if filter.DateFrom != nil {
		conditions = append(conditions, "created >= ?")
		args = append(args, *filter.DateFrom)
	}

	if filter.DateTo != nil {
		conditions = append(conditions, "created <= ?")
		args = append(args, *filter.DateTo)
	}

	if filter.Language != "" {
		conditions = append(conditions, "detected_language = ?")
		args = append(args, filter.Language)
	}

	if filter.Model != "" {
		conditions = append(conditions, "model_used = ?")
		args = append(args, filter.Model)
	}

	if filter.Provider != "" {
		conditions = append(conditions, "provider_used = ?")
		args = append(args, filter.Provider)
	}

	if filter.Success != nil {
		conditions = append(conditions, "success = ?")
		args = append(args, *filter.Success)
	}

	if filter.MinDuration != nil {
		conditions = append(conditions, "execution_time_ms >= ?")
		args = append(args, *filter.MinDuration)
	}

	if filter.MaxDuration != nil {
		conditions = append(conditions, "execution_time_ms <= ?")
		args = append(args, *filter.MaxDuration)
	}

	if filter.SearchQuery != "" {
		conditions = append(conditions, "(wheresmyprompt_query LIKE ? OR command_args LIKE ? OR error_message LIKE ?)")
		searchPattern := "%" + filter.SearchQuery + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	// Build query
	query := `
		SELECT id, conversation_id, command_args, wheresmyprompt_query,
			files2prompt_args, llm_args, detected_language, file_count,
			execution_time_ms, success, error_message, model_used,
			provider_used, created, updated
		FROM waffles_executions`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	// Execute query
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query executions: %w", err)
	}
	defer rows.Close()

	var executions []WafflesExecution
	for rows.Next() {
		var exec WafflesExecution
		var conversationID sql.NullString
		err := rows.Scan(
			&exec.ID, &conversationID, &exec.CommandArgs, &exec.WheresmypromptQuery,
			&exec.Files2promptArgs, &exec.LLMArgs, &exec.DetectedLanguage, &exec.FileCount,
			&exec.ExecutionTimeMS, &exec.Success, &exec.ErrorMessage, &exec.ModelUsed,
			&exec.ProviderUsed, &exec.Created, &exec.Updated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan execution: %w", err)
		}

		if conversationID.Valid {
			exec.ConversationID = conversationID.String
		}

		executions = append(executions, exec)
	}

	return executions, rows.Err()
}

// CountExecutions returns the count of executions matching the filter
func (d *Database) CountExecutions(filter *ExecutionFilter) (int, error) {
	if filter == nil {
		filter = &ExecutionFilter{}
	}

	var conditions []string
	var args []interface{}

	if filter.DateFrom != nil {
		conditions = append(conditions, "created >= ?")
		args = append(args, *filter.DateFrom)
	}

	if filter.DateTo != nil {
		conditions = append(conditions, "created <= ?")
		args = append(args, *filter.DateTo)
	}

	if filter.Language != "" {
		conditions = append(conditions, "detected_language = ?")
		args = append(args, filter.Language)
	}

	if filter.Model != "" {
		conditions = append(conditions, "model_used = ?")
		args = append(args, filter.Model)
	}

	if filter.Provider != "" {
		conditions = append(conditions, "provider_used = ?")
		args = append(args, filter.Provider)
	}

	if filter.Success != nil {
		conditions = append(conditions, "success = ?")
		args = append(args, *filter.Success)
	}

	if filter.SearchQuery != "" {
		conditions = append(conditions, "(wheresmyprompt_query LIKE ? OR command_args LIKE ? OR error_message LIKE ?)")
		searchPattern := "%" + filter.SearchQuery + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	query := "SELECT COUNT(*) FROM waffles_executions"

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int
	err := d.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count executions: %w", err)
	}

	return count, nil
}

// GetExecutionStats calculates statistics about executions
func (d *Database) GetExecutionStats(filter *ExecutionFilter) (*ExecutionStats, error) {
	stats := &ExecutionStats{
		LanguageBreakdown: make(map[string]int),
		ModelBreakdown:    make(map[string]int),
		ProviderBreakdown: make(map[string]int),
	}

	// Build WHERE clause for filtering
	var conditions []string
	var args []interface{}

	if filter != nil {
		if filter.DateFrom != nil {
			conditions = append(conditions, "created >= ?")
			args = append(args, *filter.DateFrom)
		}

		if filter.DateTo != nil {
			conditions = append(conditions, "created <= ?")
			args = append(args, *filter.DateTo)
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Get basic counts and averages
	query := `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as successful,
			SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END) as failed,
			AVG(execution_time_ms) as avg_time,
			SUM(file_count) as total_files
		FROM waffles_executions`

	if whereClause != "" {
		query += whereClause
	}

	var avgTime sql.NullFloat64
	err := d.db.QueryRow(query, args...).Scan(
		&stats.TotalExecutions,
		&stats.SuccessfulExecutions,
		&stats.FailedExecutions,
		&avgTime,
		&stats.TotalFiles,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic stats: %w", err)
	}

	if avgTime.Valid {
		stats.AverageExecutionTime = avgTime.Float64
	}

	// Get language breakdown
	query = `
		SELECT detected_language, COUNT(*)
		FROM waffles_executions`

	if whereClause != "" {
		query += whereClause
	}

	query += " GROUP BY detected_language"

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get language breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var language string
		var count int
		if err := rows.Scan(&language, &count); err != nil {
			return nil, fmt.Errorf("failed to scan language breakdown: %w", err)
		}
		if language != "" {
			stats.LanguageBreakdown[language] = count
		}
	}

	// Get model breakdown
	query = `
		SELECT model_used, COUNT(*)
		FROM waffles_executions`

	if whereClause != "" {
		query += whereClause
	}

	query += " GROUP BY model_used"

	rows, err = d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get model breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var model string
		var count int
		if err := rows.Scan(&model, &count); err != nil {
			return nil, fmt.Errorf("failed to scan model breakdown: %w", err)
		}
		if model != "" {
			stats.ModelBreakdown[model] = count
		}
	}

	// Get provider breakdown
	query = `
		SELECT provider_used, COUNT(*)
		FROM waffles_executions`

	if whereClause != "" {
		query += whereClause
	}

	query += " GROUP BY provider_used"

	rows, err = d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var provider string
		var count int
		if err := rows.Scan(&provider, &count); err != nil {
			return nil, fmt.Errorf("failed to scan provider breakdown: %w", err)
		}
		if provider != "" {
			stats.ProviderBreakdown[provider] = count
		}
	}

	return stats, nil
}

// ExportExecutions exports execution data to various formats
func (d *Database) ExportExecutions(options *ExportOptions, writer io.Writer) error {
	if options == nil {
		options = &ExportOptions{
			Format: ExportFormatJSON,
		}
	}

	// Query executions
	executions, err := d.QueryExecutions(options.Filter)
	if err != nil {
		return fmt.Errorf("failed to query executions for export: %w", err)
	}

	switch options.Format {
	case ExportFormatJSON:
		return d.exportJSON(executions, options, writer)
	case ExportFormatCSV:
		return d.exportCSV(executions, options, writer)
	case ExportFormatMarkdown:
		return d.exportMarkdown(executions, options, writer)
	case ExportFormatSQL:
		return d.exportSQL(executions, options, writer)
	default:
		return fmt.Errorf("unsupported export format: %s", options.Format)
	}
}

// exportJSON exports executions in JSON format
func (d *Database) exportJSON(executions []WafflesExecution, options *ExportOptions, writer io.Writer) error {
	type ExportData struct {
		Executions []WafflesExecution       `json:"executions"`
		Files      map[string][]WafflesFile `json:"files,omitempty"`
		Steps      map[string][]WafflesStep `json:"steps,omitempty"`
	}

	data := ExportData{
		Executions: executions,
	}

	if options.IncludeFiles {
		data.Files = make(map[string][]WafflesFile)
		for _, exec := range executions {
			files, err := d.GetExecutionFiles(exec.ID)
			if err != nil {
				return fmt.Errorf("failed to get files for execution %s: %w", exec.ID, err)
			}
			data.Files[exec.ID] = files
		}
	}

	if options.IncludeSteps {
		data.Steps = make(map[string][]WafflesStep)
		for _, exec := range executions {
			steps, err := d.GetExecutionSteps(exec.ID)
			if err != nil {
				return fmt.Errorf("failed to get steps for execution %s: %w", exec.ID, err)
			}
			data.Steps[exec.ID] = steps
		}
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// exportCSV exports executions in CSV format
func (d *Database) exportCSV(executions []WafflesExecution, options *ExportOptions, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{
		"ID", "ConversationID", "CommandArgs", "WheresmypromptQuery",
		"Files2promptArgs", "LLMArgs", "DetectedLanguage", "FileCount",
		"ExecutionTimeMS", "Success", "ErrorMessage", "ModelUsed",
		"ProviderUsed", "Created", "Updated",
	}

	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	for _, exec := range executions {
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

// exportMarkdown exports executions in Markdown format
func (d *Database) exportMarkdown(executions []WafflesExecution, options *ExportOptions, writer io.Writer) error {
	fmt.Fprintf(writer, "# Waffles Execution Report\n\n")
	fmt.Fprintf(writer, "Generated: %s\n\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(writer, "Total Executions: %d\n\n", len(executions))

	for _, exec := range executions {
		fmt.Fprintf(writer, "## Execution %s\n\n", exec.ID)
		fmt.Fprintf(writer, "- **Created**: %s\n", exec.Created.Format(time.RFC3339))
		fmt.Fprintf(writer, "- **Language**: %s\n", exec.DetectedLanguage)
		fmt.Fprintf(writer, "- **Model**: %s\n", exec.ModelUsed)
		fmt.Fprintf(writer, "- **Provider**: %s\n", exec.ProviderUsed)
		fmt.Fprintf(writer, "- **Success**: %t\n", exec.Success)
		fmt.Fprintf(writer, "- **Duration**: %dms\n", exec.ExecutionTimeMS)
		fmt.Fprintf(writer, "- **File Count**: %d\n", exec.FileCount)

		if exec.WheresmypromptQuery != "" {
			fmt.Fprintf(writer, "- **Query**: %s\n", exec.WheresmypromptQuery)
		}

		if exec.ErrorMessage != "" {
			fmt.Fprintf(writer, "- **Error**: %s\n", exec.ErrorMessage)
		}

		fmt.Fprintf(writer, "\n")
	}

	return nil
}

// exportSQL exports executions as SQL INSERT statements
func (d *Database) exportSQL(executions []WafflesExecution, options *ExportOptions, writer io.Writer) error {
	fmt.Fprintf(writer, "-- Waffles Execution Data Export\n")
	fmt.Fprintf(writer, "-- Generated: %s\n\n", time.Now().Format(time.RFC3339))

	for _, exec := range executions {
		fmt.Fprintf(writer,
			"INSERT INTO waffles_executions (id, conversation_id, command_args, wheresmyprompt_query, files2prompt_args, llm_args, detected_language, file_count, execution_time_ms, success, error_message, model_used, provider_used, created, updated) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s', %d, %d, %t, '%s', '%s', '%s', '%s', '%s');\n",
			exec.ID, exec.ConversationID, exec.CommandArgs, exec.WheresmypromptQuery,
			exec.Files2promptArgs, exec.LLMArgs, exec.DetectedLanguage, exec.FileCount,
			exec.ExecutionTimeMS, exec.Success, exec.ErrorMessage, exec.ModelUsed,
			exec.ProviderUsed, exec.Created.Format(time.RFC3339), exec.Updated.Format(time.RFC3339),
		)
	}

	return nil
}
