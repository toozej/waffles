package pipeline

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/toozej/waffles/pkg/config"
	"github.com/toozej/waffles/pkg/deps"
	"github.com/toozej/waffles/pkg/repo"
)

// NewPipeline creates a new pipeline instance
func NewPipeline(cfg *config.Config) *Pipeline {
	return &Pipeline{
		Config: cfg,
		// RepoInfo will be set during execution
		// Logger will be set when needed
	}
}

// NewPipelineWithLogger creates a new pipeline instance with logging
func NewPipelineWithLogger(cfg *config.Config, logger Logger) *Pipeline {
	return &Pipeline{
		Config: cfg,
		Logger: logger,
	}
}

// Execute runs the complete pipeline: wheresmyprompt → files2prompt → llm
func (p *Pipeline) Execute(promptQuery string, args []string) (*ExecutionContext, error) {
	execCtx := NewExecutionContext(promptQuery)

	// Check dependencies first
	if !p.Config.AutoInstall {
		if err := p.checkDependencies(); err != nil {
			execCtx.Complete(false, "", err)
			return execCtx, err
		}
	}

	// Analyze repository
	repoInfo, err := p.analyzeRepository()
	if err != nil {
		pipelineErr := &PipelineError{
			Phase:   PhaseRepoAnalysis,
			Message: "Failed to analyze repository",
			Err:     err,
		}
		execCtx.Complete(false, "", pipelineErr)
		return execCtx, pipelineErr
	}

	// Store repoInfo with mutex protection for thread safety
	p.mu.Lock()
	p.RepoInfo = repoInfo
	p.mu.Unlock()

	// Step 1: Execute wheresmyprompt
	promptResult, err := p.executeWheresmyprompt(promptQuery, parseCustomArgs(p.Config.WheresmypromptArgs))
	if err != nil {
		execCtx.AddStep(*promptResult)
		execCtx.Complete(false, "", err)
		return execCtx, err
	}
	execCtx.AddStep(*promptResult)

	// Step 2: Execute files2prompt
	contextResult, err := p.executeFiles2prompt(repoInfo, parseCustomArgs(p.Config.Files2promptArgs))
	if err != nil {
		execCtx.AddStep(*contextResult)
		execCtx.Complete(false, "", err)
		return execCtx, err
	}
	execCtx.AddStep(*contextResult)

	// Step 3: Execute LLM
	llmResult, err := p.executeLLM(promptResult.Output, contextResult.Output, p.Config.DefaultModel, parseCustomArgs(p.Config.LLMArgs))
	if err != nil {
		execCtx.AddStep(*llmResult)
		execCtx.Complete(false, "", err)
		return execCtx, err
	}
	execCtx.AddStep(*llmResult)

	// Complete execution successfully
	execCtx.Complete(true, llmResult.Output, nil)
	return execCtx, nil
}

// checkDependencies verifies all required tools are available
func (p *Pipeline) checkDependencies() error {
	statuses, err := deps.CheckAllDependencies()
	if err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	var missing []string
	for _, status := range statuses {
		if !status.Installed || !status.Valid {
			missing = append(missing, status.Name)
		}
	}

	if len(missing) > 0 {
		return &PipelineError{
			Phase:   PhaseDepCheck,
			Message: fmt.Sprintf("Missing or invalid dependencies: %s", strings.Join(missing, ", ")),
		}
	}

	return nil
}

// analyzeRepository analyzes the current repository
func (p *Pipeline) analyzeRepository() (*repo.RepositoryInfo, error) {
	overrides := &repo.RepositoryOverrides{
		IgnoreGitignore: p.Config.IgnoreGitignore,
	}

	if p.Config.LanguageOverride != "" {
		lang := repo.Language(p.Config.LanguageOverride)
		overrides.Language = &lang
	}

	if p.Config.IncludePatterns != "" {
		overrides.IncludePatterns = strings.Split(p.Config.IncludePatterns, ",")
	}
	if p.Config.ExcludePatterns != "" {
		overrides.ExcludePatterns = strings.Split(p.Config.ExcludePatterns, ",")
	}

	return repo.AnalyzeRepository(".", overrides)
}

// executeWheresmyprompt runs wheresmyprompt to retrieve system prompts
func (p *Pipeline) executeWheresmyprompt(query string, args []string) (*StepResult, error) {
	startTime := time.Now()
	result := &StepResult{
		Tool:      "wheresmyprompt",
		StartTime: startTime,
		Success:   false,
	}

	// Build command
	cmdArgs := buildWheresmypromptCommand(query, sanitizeArgs(args))
	result.Command = cmdArgs

	if err := validateCommand(cmdArgs); err != nil {
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)
		return result, &PipelineError{
			Phase:   PhasePromptRetrieval,
			Tool:    "wheresmyprompt",
			Message: "Invalid command",
			Err:     err,
		}
	}

	// Execute command
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Duration(getToolTimeout("wheresmyprompt"))*time.Second)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, cmdArgs[0], cmdArgs[1:]...) // #nosec G204 # nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
	output, err := cmd.CombinedOutput()

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)
	result.Output = parseToolOutput("wheresmyprompt", string(output))

	if err != nil {
		result.Error = err
		return result, &PipelineError{
			Phase:   PhasePromptRetrieval,
			Tool:    "wheresmyprompt",
			Message: fmt.Sprintf("Execution failed: %s", extractErrorFromOutput(string(output))),
			Err:     err,
		}
	}

	result.Success = true
	return result, nil
}

// executeFiles2prompt runs files2prompt to extract repository context
func (p *Pipeline) executeFiles2prompt(repoInfo *repo.RepositoryInfo, args []string) (*StepResult, error) {
	startTime := time.Now()
	result := &StepResult{
		Tool:      "files2prompt",
		StartTime: startTime,
		Success:   false,
	}

	// Build command
	cmdArgs := buildFiles2promptCommand(repoInfo, sanitizeArgs(args))
	result.Command = cmdArgs

	if err := validateCommand(cmdArgs); err != nil {
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)
		return result, &PipelineError{
			Phase:   PhaseContextExtraction,
			Tool:    "files2prompt",
			Message: "Invalid command",
			Err:     err,
		}
	}

	// Execute command
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Duration(getToolTimeout("files2prompt"))*time.Second)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, cmdArgs[0], cmdArgs[1:]...) // #nosec G204 # nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
	output, err := cmd.CombinedOutput()

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)
	result.Output = parseToolOutput("files2prompt", string(output))

	if err != nil {
		result.Error = err
		return result, &PipelineError{
			Phase:   PhaseContextExtraction,
			Tool:    "files2prompt",
			Message: fmt.Sprintf("Execution failed: %s", extractErrorFromOutput(string(output))),
			Err:     err,
		}
	}

	result.Success = true
	return result, nil
}

// executeLLM runs the LLM with the combined prompt and context
func (p *Pipeline) executeLLM(prompt, contextData, model string, args []string) (*StepResult, error) {
	startTime := time.Now()
	result := &StepResult{
		Tool:      "llm",
		StartTime: startTime,
		Success:   false,
	}

	// Combine prompt and context
	finalInput := formatPromptWithContext(prompt, contextData)

	// Truncate if necessary (assuming 8k token limit for safety)
	finalInput = truncateIfNeeded(finalInput, 6000) // Leave buffer for response

	// Build command
	cmdArgs := buildLLMCommand(prompt, contextData, model, sanitizeArgs(args))
	result.Command = cmdArgs

	if err := validateCommand(cmdArgs); err != nil {
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)
		return result, &PipelineError{
			Phase:   PhaseLLMExecution,
			Tool:    "llm",
			Message: "Invalid command",
			Err:     err,
		}
	}

	// Execute command with input
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Duration(getToolTimeout("llm"))*time.Second)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, cmdArgs[0], cmdArgs[1:]...) // #nosec G204 # nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command

	// Provide the context as stdin input
	if finalInput != "" {
		cmd.Stdin = strings.NewReader(finalInput)
	}

	output, err := cmd.CombinedOutput()

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)
	result.Output = parseToolOutput("llm", string(output))

	if err != nil {
		result.Error = err
		return result, &PipelineError{
			Phase:   PhaseLLMExecution,
			Tool:    "llm",
			Message: fmt.Sprintf("Execution failed: %s", extractErrorFromOutput(string(output))),
			Err:     err,
		}
	}

	result.Success = true
	return result, nil
}

// ExecuteWithOptions runs the pipeline with custom options
func (p *Pipeline) ExecuteWithOptions(promptQuery string, options *PipelineOptions) (*ExecutionContext, error) {
	if options == nil {
		options = DefaultOptions()
	}

	// Update config with options
	if options.Verbose {
		p.Config.Verbose = true
	}

	execCtx := NewExecutionContext(promptQuery)

	if options.DryRun {
		// For dry run, just validate and return
		if err := p.checkDependencies(); err != nil && !options.SkipDependency {
			execCtx.Complete(false, "", err)
			return execCtx, err
		}

		execCtx.Complete(true, "[DRY RUN] Pipeline would execute successfully", nil)
		return execCtx, nil
	}

	// Execute normally
	return p.Execute(promptQuery, options.CustomArgs)
}

// GetProgress returns current execution progress
func (p *Pipeline) GetProgress(ctx *ExecutionContext) ExecutionState {
	return ctx.GetState()
}

// parseCustomArgs parses a string of custom arguments
func parseCustomArgs(argsStr string) []string {
	if argsStr == "" {
		return []string{}
	}

	// Simple space-based splitting - could be enhanced with proper shell parsing
	return strings.Fields(argsStr)
}

// ValidatePipeline checks if the pipeline is properly configured
func (p *Pipeline) ValidatePipeline() error {
	if p.Config == nil {
		return fmt.Errorf("pipeline configuration is nil")
	}

	// Check for required configuration
	if p.Config.DefaultModel == "" {
		return fmt.Errorf("default model not specified")
	}

	return nil
}

// SetRepoInfo allows setting repository information externally
func (p *Pipeline) SetRepoInfo(repoInfo *repo.RepositoryInfo) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.RepoInfo = repoInfo
}

// GetConfig returns the pipeline configuration
func (p *Pipeline) GetConfig() *config.Config {
	return p.Config
}

// GetRepoInfo returns the repository information
func (p *Pipeline) GetRepoInfo() *repo.RepositoryInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.RepoInfo
}
