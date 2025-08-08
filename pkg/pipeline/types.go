package pipeline

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/toozej/waffles/pkg/config"
	"github.com/toozej/waffles/pkg/repo"
)

// Forward declaration to avoid circular imports
type Logger interface {
	LogExecution(exec interface{}) error
	LogFiles(executionID string, files interface{}) error
	LogSteps(executionID string, steps interface{}) error
}

// Pipeline represents the main orchestration pipeline
type Pipeline struct {
	Config   *config.Config
	RepoInfo *repo.RepositoryInfo
	Logger   Logger
	mu       sync.RWMutex // Protects RepoInfo from concurrent access
}

// ExecutionContext contains information about a pipeline execution
type ExecutionContext struct {
	ID             string        `json:"id"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time,omitempty"`
	Duration       time.Duration `json:"duration,omitempty"`
	PromptQuery    string        `json:"prompt_query"`
	Files          []string      `json:"files"`
	ExecutionSteps []StepResult  `json:"execution_steps"`
	Success        bool          `json:"success"`
	Error          string        `json:"error,omitempty"`
	FinalOutput    string        `json:"final_output,omitempty"`
}

// StepResult represents the result of a single pipeline step
type StepResult struct {
	Tool      string        `json:"tool"`
	Command   []string      `json:"command"`
	Output    string        `json:"output"`
	Error     error         `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Success   bool          `json:"success"`
}

// PipelineOptions contains options for pipeline execution
type PipelineOptions struct {
	DryRun         bool     `json:"dry_run"`
	Verbose        bool     `json:"verbose"`
	SkipDependency bool     `json:"skip_dependency_check"`
	TimeoutSeconds int      `json:"timeout_seconds"`
	MaxRetries     int      `json:"max_retries"`
	CustomArgs     []string `json:"custom_args,omitempty"`
}

// ToolConfiguration represents configuration for a specific tool
type ToolConfiguration struct {
	Name        string            `json:"name"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Environment map[string]string `json:"environment,omitempty"`
	Timeout     time.Duration     `json:"timeout"`
	Required    bool              `json:"required"`
}

// ExecutionPhase represents different phases of execution
type ExecutionPhase string

const (
	PhaseInit              ExecutionPhase = "init"
	PhaseDepCheck          ExecutionPhase = "dependency_check"
	PhaseRepoAnalysis      ExecutionPhase = "repo_analysis"
	PhasePromptRetrieval   ExecutionPhase = "prompt_retrieval"
	PhaseContextExtraction ExecutionPhase = "context_extraction"
	PhaseLLMExecution      ExecutionPhase = "llm_execution"
	PhaseComplete          ExecutionPhase = "complete"
	PhaseError             ExecutionPhase = "error"
)

// ExecutionState tracks the current state of pipeline execution
type ExecutionState struct {
	Phase       ExecutionPhase `json:"phase"`
	CurrentStep int            `json:"current_step"`
	TotalSteps  int            `json:"total_steps"`
	Message     string         `json:"message,omitempty"`
	Progress    float64        `json:"progress"` // 0.0 to 1.0
}

// PipelineError represents an error that occurred during pipeline execution
type PipelineError struct {
	Phase   ExecutionPhase `json:"phase"`
	Tool    string         `json:"tool,omitempty"`
	Message string         `json:"message"`
	Err     error          `json:"underlying_error,omitempty"`
}

// Error implements the error interface
func (pe *PipelineError) Error() string {
	if pe.Tool != "" {
		return fmt.Sprintf("pipeline error in %s phase (tool: %s): %s", pe.Phase, pe.Tool, pe.Message)
	}
	return fmt.Sprintf("pipeline error in %s phase: %s", pe.Phase, pe.Message)
}

// Unwrap returns the underlying error
func (pe *PipelineError) Unwrap() error {
	return pe.Err
}

// DefaultOptions returns default pipeline options
func DefaultOptions() *PipelineOptions {
	return &PipelineOptions{
		DryRun:         false,
		Verbose:        false,
		SkipDependency: false,
		TimeoutSeconds: 300, // 5 minutes
		MaxRetries:     3,
		CustomArgs:     []string{},
	}
}

// DefaultToolConfigurations returns default tool configurations
func DefaultToolConfigurations() map[string]ToolConfiguration {
	return map[string]ToolConfiguration{
		"wheresmyprompt": {
			Name:     "wheresmyprompt",
			Command:  "wheresmyprompt",
			Args:     []string{},
			Timeout:  30 * time.Second,
			Required: true,
		},
		"files2prompt": {
			Name:     "files2prompt",
			Command:  "files2prompt",
			Args:     []string{},
			Timeout:  60 * time.Second,
			Required: true,
		},
		"llm": {
			Name:     "llm",
			Command:  "llm",
			Args:     []string{},
			Timeout:  180 * time.Second,
			Required: true,
		},
	}
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(promptQuery string) *ExecutionContext {
	return &ExecutionContext{
		ID:             generateExecutionID(),
		StartTime:      time.Now(),
		PromptQuery:    promptQuery,
		Files:          []string{},
		ExecutionSteps: []StepResult{},
		Success:        false,
	}
}

// AddStep adds a step result to the execution context
func (ctx *ExecutionContext) AddStep(step StepResult) {
	ctx.ExecutionSteps = append(ctx.ExecutionSteps, step)
}

// Complete marks the execution as complete
func (ctx *ExecutionContext) Complete(success bool, output string, err error) {
	ctx.EndTime = time.Now()
	ctx.Duration = ctx.EndTime.Sub(ctx.StartTime)
	ctx.Success = success
	ctx.FinalOutput = output

	if err != nil {
		ctx.Error = err.Error()
	}
}

// GetState returns the current execution state
func (ctx *ExecutionContext) GetState() ExecutionState {
	totalSteps := 4 // wheresmyprompt, files2prompt, llm, complete
	currentStep := len(ctx.ExecutionSteps)

	phase := PhaseInit
	message := "Initializing pipeline"

	if currentStep > 0 {
		lastStep := ctx.ExecutionSteps[currentStep-1]
		switch lastStep.Tool {
		case "wheresmyprompt":
			if lastStep.Success {
				phase = PhaseContextExtraction
				message = "Retrieving file context"
			} else {
				phase = PhaseError
				message = "Failed to retrieve prompt"
			}
		case "files2prompt":
			if lastStep.Success {
				phase = PhaseLLMExecution
				message = "Executing LLM query"
			} else {
				phase = PhaseError
				message = "Failed to extract context"
			}
		case "llm":
			if lastStep.Success {
				phase = PhaseComplete
				message = "Pipeline completed successfully"
			} else {
				phase = PhaseError
				message = "LLM execution failed"
			}
		}
	}

	if ctx.EndTime.IsZero() {
		if currentStep == 0 {
			phase = PhasePromptRetrieval
			message = "Retrieving prompt"
		}
	} else {
		if ctx.Success {
			phase = PhaseComplete
			message = "Pipeline completed successfully"
		} else {
			phase = PhaseError
			if ctx.Error != "" {
				message = ctx.Error
			} else {
				message = "Pipeline execution failed"
			}
		}
	}

	progress := float64(currentStep) / float64(totalSteps)
	if progress > 1.0 {
		progress = 1.0
	}

	return ExecutionState{
		Phase:       phase,
		CurrentStep: currentStep,
		TotalSteps:  totalSteps,
		Message:     message,
		Progress:    progress,
	}
}

// generateExecutionID generates a unique execution ID
func generateExecutionID() string {
	// Use a combination of timestamp and a cryptographically secure random component for better uniqueness
	nano := time.Now().UnixNano()
	randomBig, err := rand.Int(rand.Reader, big.NewInt(9223372036854775807)) // max int64
	if err != nil {
		// Fallback to just timestamp if crypto/rand fails
		return fmt.Sprintf("exec-%d", nano)
	}
	random := randomBig.Int64()
	return fmt.Sprintf("exec-%d-%d", nano, random)
}
