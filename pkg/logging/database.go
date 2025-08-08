package logging

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// NewDatabase creates a new database instance and initializes the schema
func NewDatabase(path string) (*Database, error) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to ping database: %w (close error: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{
		path: path,
		db:   db,
	}

	// Initialize schema
	if err := database.InitializeSchema(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to initialize schema: %w (close error: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return database, nil
}

// InitializeSchema creates tables and runs migrations
func (d *Database) InitializeSchema() error {
	// Check if this is an existing llm CLI database by looking for conversations table
	var conversationsExists bool
	checkErr := d.db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM sqlite_master
		WHERE type='table' AND name='conversations'
	`).Scan(&conversationsExists)
	if checkErr != nil {
		return fmt.Errorf("failed to check for conversations table: %w", checkErr)
	}

	// Always disable foreign keys for our standalone tables initially
	// We'll only reference conversations table if it exists, but not enforce FK constraint
	if _, err := d.db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("failed to disable foreign keys: %w", err)
	}

	// Check if schema version table exists
	var versionTableExists bool
	err := d.db.QueryRow(`
		SELECT COUNT(*) > 0 
		FROM sqlite_master 
		WHERE type='table' AND name='schema_version'
	`).Scan(&versionTableExists)
	if err != nil {
		return fmt.Errorf("failed to check schema version table: %w", err)
	}

	// Create schema version table if it doesn't exist
	if !versionTableExists {
		_, err := d.db.Exec(`
			CREATE TABLE schema_version (
				version INTEGER PRIMARY KEY,
				applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create schema version table: %w", err)
		}
	}

	// Get current schema version
	currentVersion := 0
	err = d.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	// Apply migrations
	targetVersion := GetCurrentSchemaVersion()
	migrations := GetRequiredMigrations(currentVersion, targetVersion)

	for i, migration := range migrations {
		version := currentVersion + i + 1

		// Begin transaction
		tx, err := d.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin migration transaction: %w", err)
		}

		// Execute migration
		if _, err := tx.Exec(migration); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("failed to execute migration %d: %w (rollback error: %v)", version, err, rollbackErr)
			}
			return fmt.Errorf("failed to execute migration %d: %w", version, err)
		}

		// Record migration
		if _, err := tx.Exec("INSERT INTO schema_version (version) VALUES (?)", version); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("failed to record migration %d: %w (rollback error: %v)", version, err, rollbackErr)
			}
			return fmt.Errorf("failed to record migration %d: %w", version, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", version, err)
		}
	}

	return nil
}

// LogExecution records a complete Waffles execution
func (d *Database) LogExecution(exec *WafflesExecution) error {
	if exec.ID == "" {
		exec.ID = uuid.New().String()
	}

	now := time.Now()
	if exec.Created.IsZero() {
		exec.Created = now
	}
	if exec.Updated.IsZero() {
		exec.Updated = now
	}

	// Handle null conversation_id
	var conversationID interface{}
	if exec.ConversationID == "" {
		conversationID = nil
	} else {
		conversationID = exec.ConversationID
	}

	_, err := d.db.Exec(`
		INSERT INTO waffles_executions (
			id, conversation_id, command_args, wheresmyprompt_query,
			files2prompt_args, llm_args, detected_language, file_count,
			execution_time_ms, success, error_message, model_used,
			provider_used, created, updated
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		exec.ID, conversationID, exec.CommandArgs, exec.WheresmypromptQuery,
		exec.Files2promptArgs, exec.LLMArgs, exec.DetectedLanguage, exec.FileCount,
		exec.ExecutionTimeMS, exec.Success, exec.ErrorMessage, exec.ModelUsed,
		exec.ProviderUsed, exec.Created, exec.Updated,
	)

	if err != nil {
		return fmt.Errorf("failed to log execution: %w", err)
	}

	return nil
}

// LogFiles records files processed during an execution
func (d *Database) LogFiles(executionID string, files []WafflesFile) error {
	if len(files) == 0 {
		return nil
	}

	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin files transaction: %w", err)
	}
	defer func() {
		// Only rollback if transaction hasn't been committed
		_ = tx.Rollback() // Ignore rollback errors on deferred cleanup
	}()

	stmt, err := tx.Prepare(`
		INSERT INTO waffles_files (
			id, execution_id, file_path, file_size, 
			included, exclusion_reason, created
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare files statement: %w", err)
	}
	defer stmt.Close()

	for _, file := range files {
		if file.ID == "" {
			file.ID = uuid.New().String()
		}
		if file.Created.IsZero() {
			file.Created = time.Now()
		}

		_, err := stmt.Exec(
			file.ID, executionID, file.FilePath, file.FileSize,
			file.Included, file.ExclusionReason, file.Created,
		)
		if err != nil {
			return fmt.Errorf("failed to log file %s: %w", file.FilePath, err)
		}
	}

	return tx.Commit()
}

// LogSteps records individual tool executions within a pipeline
func (d *Database) LogSteps(executionID string, steps []WafflesStep) error {
	if len(steps) == 0 {
		return nil
	}

	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin steps transaction: %w", err)
	}
	defer func() {
		// Only rollback if transaction hasn't been committed
		_ = tx.Rollback() // Ignore rollback errors on deferred cleanup
	}()

	stmt, err := tx.Prepare(`
		INSERT INTO waffles_steps (
			id, execution_id, tool, command, output, error_output,
			success, duration_ms, step_order, created
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare steps statement: %w", err)
	}
	defer stmt.Close()

	for _, step := range steps {
		if step.ID == "" {
			step.ID = uuid.New().String()
		}
		if step.Created.IsZero() {
			step.Created = time.Now()
		}

		_, err := stmt.Exec(
			step.ID, executionID, step.Tool, step.Command, step.Output,
			step.ErrorOutput, step.Success, step.DurationMS, step.StepOrder, step.Created,
		)
		if err != nil {
			return fmt.Errorf("failed to log step %s: %w", step.Tool, err)
		}
	}

	return tx.Commit()
}

// UpdateExecution updates an existing execution record
func (d *Database) UpdateExecution(exec *WafflesExecution) error {
	exec.Updated = time.Now()

	_, err := d.db.Exec(`
		UPDATE waffles_executions SET 
			conversation_id = ?, command_args = ?, wheresmyprompt_query = ?,
			files2prompt_args = ?, llm_args = ?, detected_language = ?,
			file_count = ?, execution_time_ms = ?, success = ?,
			error_message = ?, model_used = ?, provider_used = ?, updated = ?
		WHERE id = ?`,
		exec.ConversationID, exec.CommandArgs, exec.WheresmypromptQuery,
		exec.Files2promptArgs, exec.LLMArgs, exec.DetectedLanguage,
		exec.FileCount, exec.ExecutionTimeMS, exec.Success,
		exec.ErrorMessage, exec.ModelUsed, exec.ProviderUsed,
		exec.Updated, exec.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update execution: %w", err)
	}

	return nil
}

// GetExecution retrieves a single execution by ID
func (d *Database) GetExecution(id string) (*WafflesExecution, error) {
	var exec WafflesExecution
	var conversationID sql.NullString

	err := d.db.QueryRow(`
		SELECT id, conversation_id, command_args, wheresmyprompt_query,
			files2prompt_args, llm_args, detected_language, file_count,
			execution_time_ms, success, error_message, model_used,
			provider_used, created, updated
		FROM waffles_executions WHERE id = ?`, id).Scan(
		&exec.ID, &conversationID, &exec.CommandArgs, &exec.WheresmypromptQuery,
		&exec.Files2promptArgs, &exec.LLMArgs, &exec.DetectedLanguage, &exec.FileCount,
		&exec.ExecutionTimeMS, &exec.Success, &exec.ErrorMessage, &exec.ModelUsed,
		&exec.ProviderUsed, &exec.Created, &exec.Updated,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("execution not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	if conversationID.Valid {
		exec.ConversationID = conversationID.String
	}

	return &exec, nil
}

// GetExecutionFiles retrieves all files for an execution
func (d *Database) GetExecutionFiles(executionID string) ([]WafflesFile, error) {
	rows, err := d.db.Query(`
		SELECT id, execution_id, file_path, file_size, included, 
			exclusion_reason, created
		FROM waffles_files 
		WHERE execution_id = ? 
		ORDER BY file_path`, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query execution files: %w", err)
	}
	defer rows.Close()

	var files []WafflesFile
	for rows.Next() {
		var file WafflesFile
		err := rows.Scan(
			&file.ID, &file.ExecutionID, &file.FilePath, &file.FileSize,
			&file.Included, &file.ExclusionReason, &file.Created,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan execution file: %w", err)
		}
		files = append(files, file)
	}

	return files, rows.Err()
}

// GetExecutionSteps retrieves all steps for an execution
func (d *Database) GetExecutionSteps(executionID string) ([]WafflesStep, error) {
	rows, err := d.db.Query(`
		SELECT id, execution_id, tool, command, output, error_output,
			success, duration_ms, step_order, created
		FROM waffles_steps 
		WHERE execution_id = ? 
		ORDER BY step_order`, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query execution steps: %w", err)
	}
	defer rows.Close()

	var steps []WafflesStep
	for rows.Next() {
		var step WafflesStep
		err := rows.Scan(
			&step.ID, &step.ExecutionID, &step.Tool, &step.Command, &step.Output,
			&step.ErrorOutput, &step.Success, &step.DurationMS, &step.StepOrder, &step.Created,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan execution step: %w", err)
		}
		steps = append(steps, step)
	}

	return steps, rows.Err()
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// GetDB returns the underlying sql.DB connection for advanced operations
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// GetPath returns the database file path
func (d *Database) GetPath() string {
	return d.path
}
