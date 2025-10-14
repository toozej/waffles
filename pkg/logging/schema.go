package logging

// WafflesSchema contains all SQL statements for Waffles-specific tables
const WafflesSchema = `
-- Waffles execution tracking table
CREATE TABLE IF NOT EXISTS waffles_executions (
    id TEXT PRIMARY KEY,
    conversation_id TEXT,
    command_args TEXT NOT NULL,
    wheresmyprompt_query TEXT,
    files2prompt_args TEXT,
    llm_args TEXT,
    detected_language TEXT,
    file_count INTEGER DEFAULT 0,
    execution_time_ms INTEGER DEFAULT 0,
    success BOOLEAN DEFAULT FALSE,
    error_message TEXT,
    model_used TEXT,
    provider_used TEXT,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP
    
    -- Note: conversation_id can reference llm CLI conversations table when available
    -- But we don't enforce FK constraint to allow standalone operation
);

-- Waffles file tracking table
CREATE TABLE IF NOT EXISTS waffles_files (
    id TEXT PRIMARY KEY,
    execution_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_size INTEGER DEFAULT 0,
    included BOOLEAN DEFAULT TRUE,
    exclusion_reason TEXT,
    created DATETIME DEFAULT CURRENT_TIMESTAMP
    
    -- Note: we don't enforce FK constraint to allow standalone operation
    -- FOREIGN KEY (execution_id) REFERENCES waffles_executions(id) ON DELETE CASCADE
);

-- Waffles step tracking table (individual tool executions)
CREATE TABLE IF NOT EXISTS waffles_steps (
    id TEXT PRIMARY KEY,
    execution_id TEXT NOT NULL,
    tool TEXT NOT NULL,
    command TEXT NOT NULL,
    output TEXT,
    error_output TEXT,
    success BOOLEAN DEFAULT FALSE,
    duration_ms INTEGER DEFAULT 0,
    step_order INTEGER NOT NULL,
    created DATETIME DEFAULT CURRENT_TIMESTAMP
    
    -- Note: we don't enforce FK constraint to allow standalone operation
    -- FOREIGN KEY (execution_id) REFERENCES waffles_executions(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_waffles_executions_created ON waffles_executions(created);
CREATE INDEX IF NOT EXISTS idx_waffles_executions_language ON waffles_executions(detected_language);
CREATE INDEX IF NOT EXISTS idx_waffles_executions_model ON waffles_executions(model_used);
CREATE INDEX IF NOT EXISTS idx_waffles_executions_provider ON waffles_executions(provider_used);
CREATE INDEX IF NOT EXISTS idx_waffles_executions_success ON waffles_executions(success);
CREATE INDEX IF NOT EXISTS idx_waffles_executions_conv_id ON waffles_executions(conversation_id);

CREATE INDEX IF NOT EXISTS idx_waffles_files_execution_id ON waffles_files(execution_id);
CREATE INDEX IF NOT EXISTS idx_waffles_files_included ON waffles_files(included);
CREATE INDEX IF NOT EXISTS idx_waffles_files_path ON waffles_files(file_path);

CREATE INDEX IF NOT EXISTS idx_waffles_steps_execution_id ON waffles_steps(execution_id);
CREATE INDEX IF NOT EXISTS idx_waffles_steps_tool ON waffles_steps(tool);
CREATE INDEX IF NOT EXISTS idx_waffles_steps_success ON waffles_steps(success);
CREATE INDEX IF NOT EXISTS idx_waffles_steps_order ON waffles_steps(execution_id, step_order);

-- Triggers to update timestamps
CREATE TRIGGER IF NOT EXISTS update_waffles_executions_updated
    AFTER UPDATE ON waffles_executions
    FOR EACH ROW
    BEGIN
        UPDATE waffles_executions SET updated = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;
`

// MigrationQueries contains versioned migration queries
var MigrationQueries = map[int]string{
	1: WafflesSchema,
	// Future migrations can be added here
	// 2: "ALTER TABLE waffles_executions ADD COLUMN new_field TEXT;",
}

// GetCurrentSchemaVersion returns the current schema version
func GetCurrentSchemaVersion() int {
	maxVersion := 0
	for version := range MigrationQueries {
		if version > maxVersion {
			maxVersion = version
		}
	}
	return maxVersion
}

// GetMigrationQuery returns the migration query for a specific version
func GetMigrationQuery(version int) (string, bool) {
	query, exists := MigrationQueries[version]
	return query, exists
}

// GetRequiredMigrations returns all migrations needed from current to target version
func GetRequiredMigrations(currentVersion, targetVersion int) []string {
	var migrations []string
	for version := currentVersion + 1; version <= targetVersion; version++ {
		if query, exists := MigrationQueries[version]; exists {
			migrations = append(migrations, query)
		}
	}
	return migrations
}
