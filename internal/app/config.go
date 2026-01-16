package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	// "agent-sdk-go/internal/sdk"  // TODO: Implement SDK package
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	APIKey   string `yaml:"api_key"`
	BaseURL  string `yaml:"base_url"`
	Timeout  int    `yaml:"timeout"` // seconds
	Model    string `yaml:"model"`
}

// Load loads configuration from multiple sources with precedence:
// 1. Environment variables
// 2. Local config file (./config.yaml)
// 3. User config file (~/.config/agent-sdk-go/config.yaml)
// 4. System config file (/etc/agent-sdk-go/config.yaml)
//
// Returns the loaded config or an error if loading fails.
func Load() (*Config, error) {
	cfg := &Config{}

	// Try to load from config files
	configPaths := []string{
		"./config.yaml",                              // Local directory
		filepath.Join(userConfigDir(), "config.yaml"), // User config
		"/etc/agent-sdk-go/config.yaml",               // System config
	}

	for _, path := range configPaths {
		if err := loadFromFile(cfg, path); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("load config from %s: %w", path, err)
		}
	}

	// Override with environment variables (highest priority)
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		cfg.APIKey = apiKey
	}

	return cfg, nil
}

// loadFromFile loads configuration from a YAML file.
// If the file doesn't exist, it returns nil (not an error).
func loadFromFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("unmarshal YAML: %w", err)
	}

	return nil
}

// userConfigDir returns the user configuration directory.
func userConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "agent-sdk-go")
}

// NewClientFromConfig creates a new SDK client from the configuration.
// It returns an error if the API key is missing or if client creation fails.
// Deprecated: Use the V2 SDK via internal/claude/v2 package instead.
func NewClientFromConfig(cfg *Config) (any, error) {
	// V2 SDK uses direct functions, not a client interface
	// Use v2.Prompt() or v2.CreateSession() directly
	return nil, fmt.Errorf("use V2 SDK via internal/claude/v2 package instead")
}

// NewClient creates a new SDK client by loading configuration first.
// This is a convenience wrapper around Load() and NewClientFromConfig().
// Deprecated: Use the V2 SDK via internal/claude/v2 package instead.
func NewClient() (any, error) {
	cfg, err := Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return NewClientFromConfig(cfg)
}

// GetModel returns the model from config with a default fallback.
// If config.Model is empty, returns the defaultModel.
func (c *Config) GetModel(defaultModel string) string {
	if c.Model != "" {
		return c.Model
	}
	return defaultModel
}

// GetTimeoutDuration returns the timeout as time.Duration.
// Returns 0 if timeout is <= 0.
func (c *Config) GetTimeoutDuration() time.Duration {
	if c.Timeout <= 0 {
		return 0
	}
	return time.Duration(c.Timeout) * time.Second
}