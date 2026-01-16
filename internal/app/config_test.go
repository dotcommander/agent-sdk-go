package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_EnvironmentVariable(t *testing.T) {
	// Set environment variable
	os.Setenv("ANTHROPIC_API_KEY", "env-test-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "env-test-key", cfg.APIKey)
}

func TestLoad_ConfigFile(t *testing.T) {
	// Create temporary directory for config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write test config
	configContent := `
api_key: "file-test-key"
base_url: "https://test.example.com"
timeout: 30
model: "claude-test-model"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Temporarily change directory to read local config
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "file-test-key", cfg.APIKey)
	assert.Equal(t, "https://test.example.com", cfg.BaseURL)
	assert.Equal(t, 30, cfg.Timeout)
	assert.Equal(t, "claude-test-model", cfg.Model)
}

func TestLoad_EnvironmentOverridesFile(t *testing.T) {
	// Create config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
api_key: "file-key"
base_url: "https://file.example.com"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variable (should override)
	os.Setenv("ANTHROPIC_API_KEY", "env-override-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	// Temporarily change directory
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	cfg, err := Load()
	require.NoError(t, err)

	// Environment should override file
	assert.Equal(t, "env-override-key", cfg.APIKey)
	// Other values from file should remain
	assert.Equal(t, "https://file.example.com", cfg.BaseURL)
}

func TestLoad_MissingConfigFiles(t *testing.T) {
	// No environment variable set
	os.Unsetenv("ANTHROPIC_API_KEY")

	// No config files exist - should return empty config
	cfg, err := Load()
	require.NoError(t, err)
	assert.Empty(t, cfg.APIKey)
	assert.Empty(t, cfg.BaseURL)
	assert.Zero(t, cfg.Timeout)
	assert.Empty(t, cfg.Model)
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write invalid YAML
	err := os.WriteFile(configPath, []byte("invalid: yaml: : content"), 0644)
	require.NoError(t, err)

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	cfg, err := Load()
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "unmarshal YAML")
}

func TestNewClientFromConfig_Deprecated(t *testing.T) {
	cfg := &Config{
		APIKey:   "test-api-key",
		BaseURL:  "https://test.example.com",
		Timeout:  30,
	}

	// V1 client is deprecated - should return error directing to V2 SDK
	client, err := NewClientFromConfig(cfg)
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "V2 SDK")
}

func TestNewClient_Deprecated(t *testing.T) {
	// Set environment variable
	os.Setenv("ANTHROPIC_API_KEY", "test-key-for-new-client")
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	// V1 client is deprecated - should return error directing to V2 SDK
	client, err := NewClient()
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "V2 SDK")
}

func TestLoadFromFile_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "does-not-exist.yaml")

	var cfg Config
	err := loadFromFile(&cfg, nonExistentPath)
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err), "Should return IsNotExist error")
}

func TestConfig_GetModel(t *testing.T) {
	tests := []struct {
		name         string
		configModel  string
		defaultModel string
		want         string
	}{
		{"config has model", "claude-config", "claude-default", "claude-config"},
		{"config empty", "", "claude-default", "claude-default"},
		{"both empty", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Model: tt.configModel}
			got := cfg.GetModel(tt.defaultModel)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfig_GetTimeoutDuration(t *testing.T) {
	tests := []struct {
		name     string
		timeout  int
		want     time.Duration
	}{
		{"positive timeout", 30, 30 * time.Second},
		{"zero timeout", 0, 0},
		{"negative timeout", -1, 0},
		{"large timeout", 300, 300 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Timeout: tt.timeout}
			got := cfg.GetTimeoutDuration()
			assert.Equal(t, tt.want, got)
		})
	}
}