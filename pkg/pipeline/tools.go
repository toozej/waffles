package pipeline

import (
	"fmt"
	"strings"

	"github.com/toozej/waffles/pkg/repo"
)

// buildWheresmypromptCommand builds the command for wheresmyprompt execution
func buildWheresmypromptCommand(query string, args []string) []string {
	cmd := []string{"wheresmyprompt"}

	// Add query terms
	if query != "" {
		// Split query into individual terms for wheresmyprompt
		terms := strings.Fields(query)
		cmd = append(cmd, terms...)
	}

	// Add custom arguments
	cmd = append(cmd, args...)

	return cmd
}

// buildFiles2promptCommand builds the command for files2prompt execution
func buildFiles2promptCommand(repoInfo *repo.RepositoryInfo, args []string) []string {
	cmd := []string{"files2prompt"}

	if repoInfo != nil {
		// Add language-specific arguments
		if repoInfo.Language != repo.LanguageUnknown {
			cmd = append(cmd, "--language", string(repoInfo.Language))
		}

		// Add include patterns
		for _, pattern := range repoInfo.IncludePatterns {
			cmd = append(cmd, "--include", pattern)
		}

		// Add exclude patterns
		for _, pattern := range repoInfo.ExcludePatterns {
			cmd = append(cmd, "--exclude", pattern)
		}

		// Add specific files if we have analyzed files
		includedFiles := getIncludedFiles(repoInfo)
		if len(includedFiles) > 0 && len(includedFiles) < 100 { // Avoid too many args
			cmd = append(cmd, includedFiles...)
		} else {
			// If too many files or no specific files, let files2prompt scan
			cmd = append(cmd, ".")
		}
	} else {
		// Default to current directory
		cmd = append(cmd, ".")
	}

	// Add custom arguments
	cmd = append(cmd, args...)

	return cmd
}

// buildLLMCommand builds the command for LLM execution
func buildLLMCommand(prompt, context, model string, args []string) []string {
	cmd := []string{"llm"}

	// Add model if specified
	if model != "" {
		cmd = append(cmd, "-m", model)
	}

	// Add system prompt from wheresmyprompt if available
	if prompt != "" {
		cmd = append(cmd, "--system", prompt)
	}

	// Add custom arguments
	cmd = append(cmd, args...)

	// Add the context as the final argument
	if context != "" {
		cmd = append(cmd, context)
	}

	return cmd
}

// sanitizeArgs removes potentially dangerous arguments
func sanitizeArgs(args []string) []string {
	var safe []string

	dangerous := []string{
		"--exec", "--execute", "--eval", "--script",
		"rm", "del", "delete", "format", "mkfs",
		"sudo", "su", "chmod", "chown",
		"curl", "wget", "nc", "netcat",
	}

	for _, arg := range args {
		isDangerous := false
		argLower := strings.ToLower(arg)

		for _, danger := range dangerous {
			if strings.Contains(argLower, danger) {
				isDangerous = true
				break
			}
		}

		if !isDangerous {
			safe = append(safe, arg)
		}
	}

	return safe
}

// getIncludedFiles extracts the list of included files from repository info
func getIncludedFiles(repoInfo *repo.RepositoryInfo) []string {
	var files []string

	for _, file := range repoInfo.DetectedFiles {
		if file.Included {
			// Use relative path from repository root
			files = append(files, file.Path)
		}
	}

	return files
}

// validateCommand performs basic validation on a command
func validateCommand(cmd []string) error {
	if len(cmd) == 0 {
		return fmt.Errorf("empty command")
	}

	// Check for path traversal attempts
	for _, arg := range cmd {
		if strings.Contains(arg, "../") || strings.Contains(arg, "..\\") {
			return fmt.Errorf("path traversal detected in argument: %s", arg)
		}
	}

	return nil
}

// parseToolOutput parses and cleans output from external tools
func parseToolOutput(tool, output string) string {
	// Remove common noise from tool outputs
	lines := strings.Split(output, "\n")
	var cleaned []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip common debug/info prefixes
		if strings.HasPrefix(line, "DEBUG:") ||
			strings.HasPrefix(line, "INFO:") ||
			strings.HasPrefix(line, "WARN:") {
			continue
		}

		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, "\n")
}

// estimateTokens provides a rough estimate of tokens in text
func estimateTokens(text string) int {
	// Rough approximation: 1 token ≈ 4 characters
	return len(text) / 4
}

// truncateIfNeeded truncates text if it's too long for the model
func truncateIfNeeded(text string, maxTokens int) string {
	estimatedTokens := estimateTokens(text)

	if estimatedTokens <= maxTokens {
		return text
	}

	// Calculate how much to keep (leave some buffer)
	keepRatio := float64(maxTokens) * 0.9 / float64(estimatedTokens)
	keepChars := int(float64(len(text)) * keepRatio)

	if keepChars < 100 {
		return text[:100] + "...\n[Content truncated due to length]"
	}

	return text[:keepChars] + "...\n[Content truncated due to length]"
}

// formatPromptWithContext combines prompt and context for LLM input
func formatPromptWithContext(prompt, context string) string {
	if prompt == "" && context == "" {
		return ""
	}

	if prompt == "" {
		return context
	}

	if context == "" {
		return prompt
	}

	return fmt.Sprintf("%s\n\nContext:\n%s", prompt, context)
}

// extractErrorFromOutput attempts to extract meaningful error information
func extractErrorFromOutput(output string) string {
	lines := strings.Split(output, "\n")

	// Look for lines that seem like errors
	for _, line := range lines {
		line = strings.TrimSpace(line)
		lineLower := strings.ToLower(line)

		if strings.Contains(lineLower, "error:") ||
			strings.Contains(lineLower, "failed:") ||
			strings.Contains(lineLower, "exception:") {
			return line
		}
	}

	// If no clear error found, return first non-empty line
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}

	return "Unknown error"
}

// getToolTimeout returns appropriate timeout for a tool
func getToolTimeout(toolName string) int {
	timeouts := map[string]int{
		"wheresmyprompt": 30,  // 30 seconds
		"files2prompt":   60,  // 1 minute
		"llm":            300, // 5 minutes
	}

	if timeout, exists := timeouts[toolName]; exists {
		return timeout
	}

	return 60 // Default 1 minute
}
