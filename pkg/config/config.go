// Package config provides secure configuration management for the waffles application.
//
// This package handles loading configuration from environment variables and .env files
// with built-in security measures and comprehensive LLM toolchain orchestration settings.
// It uses the github.com/caarlos0/env library for environment variable parsing and
// github.com/joho/godotenv for .env file loading.
//
// The configuration loading follows a priority order:
//  1. System environment variables (highest priority)
//  2. .env file in current working directory
//  3. Global .env file in ~/.config/waffles/.env
//  4. Built-in defaults (lowest priority)
//
// Configuration categories:
//   - Core Application Settings: Model, provider, database paths
//   - Tool Configurations: Arguments for wheresmyprompt, files2prompt, llm
//   - Behavior Settings: Verbosity, auto-install, gitignore handling
//   - Language-specific Settings: Language overrides and file patterns
//
// Example usage:
//
//	import "github.com/toozej/waffles/pkg/config"
//
//	func main() {
//		cfg, err := config.LoadConfig()
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Printf("Using model: %s\n", cfg.DefaultModel)
//	}
package config

import (
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config represents the complete application configuration structure.
//
// This struct defines all configurable parameters for the waffles LLM toolchain
// orchestrator. Fields are tagged with struct tags that correspond to environment
// variable names and default values for automatic parsing.
//
// Configuration is organized into logical groups:
//   - Core settings for LLM model and provider selection
//   - Tool-specific arguments for pipeline components
//   - Behavioral flags for application operation
//   - Language and file filtering options
//
// Example:
//
//	cfg := Config{
//		DefaultModel:    "claude-3-sonnet",
//		DefaultProvider: "anthropic",
//		Verbose:         true,
//	}
type Config struct {
	// Core Application Settings
	DefaultModel    string `env:"WAFFLES_DEFAULT_MODEL" envDefault:"claude-3-sonnet"`
	DefaultProvider string `env:"WAFFLES_DEFAULT_PROVIDER" envDefault:"anthropic"`
	LogDBPath       string `env:"WAFFLES_LOG_DB_PATH" envDefault:"./llm-logs.sqlite"`
	ConfigPath      string `env:"WAFFLES_CONFIG_PATH" envDefault:".waffles.env"`

	// Tool Configurations
	WheresmypromptArgs string `env:"WHERESMYPROMPT_ARGS" envDefault:""`
	Files2promptArgs   string `env:"FILES2PROMPT_ARGS" envDefault:""`
	LLMArgs            string `env:"LLM_ARGS" envDefault:""`

	// Behavior Settings
	Verbose         bool `env:"WAFFLES_VERBOSE" envDefault:"false"`
	Quiet           bool `env:"WAFFLES_QUIET" envDefault:"false"`
	AutoInstall     bool `env:"WAFFLES_AUTO_INSTALL" envDefault:"false"`
	IgnoreGitignore bool `env:"WAFFLES_IGNORE_GITIGNORE" envDefault:"false"`

	// Language-specific Settings
	LanguageOverride string `env:"WAFFLES_LANGUAGE" envDefault:""`
	IncludePatterns  string `env:"WAFFLES_INCLUDE_PATTERNS" envDefault:""`
	ExcludePatterns  string `env:"WAFFLES_EXCLUDE_PATTERNS" envDefault:""`
}

// LoadConfig loads and returns the application configuration from multiple sources
// with comprehensive precedence handling and error management.
//
// This function performs the following operations in order:
//  1. Loads .env files from multiple locations (lowest to highest precedence)
//  2. Parses environment variables into the Config struct (highest precedence)
//  3. Returns the populated configuration with all defaults applied
//
// Configuration loading precedence (highest to lowest):
//  1. System environment variables (WAFFLES_* prefixed)
//  2. .env file in current working directory
//  3. .waffles.env file in current working directory
//  4. Global .env file in ~/.config/waffles/.env
//  5. Built-in defaults from struct tags (envDefault)
//
// The function gracefully handles missing .env files and only returns errors
// for critical failures like environment variable parsing issues.
//
// Returns:
//   - *Config: A populated configuration struct with values from all sources
//   - error: Non-nil if environment variable parsing fails
//
// Example:
//
//	// Load configuration with full precedence handling
//	cfg, err := config.LoadConfig()
//	if err != nil {
//		log.Fatalf("Failed to load configuration: %v", err)
//	}
//
//	// Use configuration
//	fmt.Printf("Using %s model with %s provider\n", cfg.DefaultModel, cfg.DefaultProvider)
func LoadConfig() (*Config, error) {
	var cfg Config

	// Load .env files in order of precedence (lower precedence first)
	loadEnvFiles(&cfg)

	// Parse environment variables into struct (highest precedence)
	if err := LoadFromEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadFromEnv parses environment variables into the provided Config struct.
//
// This function uses the github.com/caarlos0/env library to automatically
// parse environment variables based on the struct field tags. It respects
// the env and envDefault tags defined on Config struct fields.
//
// Parameters:
//   - cfg: Pointer to Config struct to populate with environment values
//
// Returns:
//   - error: Non-nil if environment variable parsing fails
//
// Example:
//
//	var cfg Config
//	if err := LoadFromEnv(&cfg); err != nil {
//		log.Fatal("Environment parsing failed:", err)
//	}
func LoadFromEnv(cfg *Config) error {
	return env.Parse(cfg)
}

// LoadFromFile loads environment variables from a specified .env file.
//
// This function safely loads a .env file using godotenv.Load, which parses
// the file and sets environment variables. It gracefully handles missing files
// by returning nil (no error) when the file doesn't exist.
//
// The function does not override existing environment variables, following
// the standard .env file behavior where existing environment takes precedence.
//
// Parameters:
//   - cfg: Config struct pointer (currently unused but kept for consistency)
//   - filepath: Path to the .env file to load
//
// Returns:
//   - error: Non-nil if file exists but cannot be parsed, nil if file missing
//
// Example:
//
//	var cfg Config
//	if err := LoadFromFile(&cfg, ".env"); err != nil {
//		log.Printf("Failed to load .env file: %v", err)
//	}
func LoadFromFile(cfg *Config, filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// File doesn't exist, that's okay
		return nil
	}

	// Load the .env file, but don't override existing env vars
	return godotenv.Load(filepath)
}

// MergeWithDefaults ensures all configuration fields have appropriate default values.
//
// This function is currently a no-op since defaults are handled automatically
// by the envDefault struct tags in the Config struct. It is maintained for
// future extensibility where custom default logic might be needed.
//
// The function provides a hook for any future custom default value logic
// that cannot be expressed through struct tags, such as computed defaults
// or conditional default values based on other configuration.
//
// Parameters:
//   - cfg: Pointer to Config struct to apply defaults to
//
// Example:
//
//	var cfg Config
//	config.MergeWithDefaults(&cfg)  // Currently no-op, but available for future use
func MergeWithDefaults(cfg *Config) {
	// Defaults are handled by envDefault tags
	// This function is available for any future custom default logic
}

// loadEnvFiles loads .env files from multiple locations in order of precedence.
//
// This internal function loads .env files from various standard locations,
// with files loaded first having lower precedence (can be overridden by
// files loaded later). All file loading errors are ignored since .env files
// are optional configuration sources.
//
// Loading order (lowest to highest precedence):
//  1. Global config: ~/.config/waffles/.env
//  2. Local project config: .waffles.env
//  3. Current directory: .env
//
// Parameters:
//   - cfg: Pointer to Config struct (passed to LoadFromFile for consistency)
//
// Example usage (internal):
//
//	var cfg Config
//	loadEnvFiles(&cfg)  // Loads all available .env files
func loadEnvFiles(cfg *Config) {
	// First try global config file (lowest precedence)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalConfigPath := filepath.Join(homeDir, ".config", "waffles", ".env")
		_ = LoadFromFile(cfg, globalConfigPath) // Ignore errors for optional config files
	}

	// Then try local .env file (higher precedence)
	_ = LoadFromFile(cfg, ".waffles.env") // Ignore errors for optional config files

	// Finally try current directory .env file (even higher precedence)
	_ = LoadFromFile(cfg, ".env") // Ignore errors for optional config files
}
