package app

import (
	"os"
	"path/filepath"
	"testing"
)


func TestExampleIntegration(t *testing.T) {
	// Create a test config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
api_key: "test-integration-key"
base_url: "https://api.test.example.com"
timeout: 60
model: "claude-3-test"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Temporarily change directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Test the config loading flow
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.APIKey != "test-integration-key" {
		t.Errorf("Expected API key 'test-integration-key', got %q", cfg.APIKey)
	}

	if cfg.BaseURL != "https://api.test.example.com" {
		t.Errorf("Expected base URL 'https://api.test.example.com', got %q", cfg.BaseURL)
	}

	if cfg.Timeout != 60 {
		t.Errorf("Expected timeout 60, got %d", cfg.Timeout)
	}

	if cfg.Model != "claude-3-test" {
		t.Errorf("Expected model 'claude-3-test', got %q", cfg.Model)
	}

	// Note: NewClientFromConfig is deprecated - use V2 SDK instead
	// Verify the deprecation error is returned
	_, err = NewClientFromConfig(cfg)
	if err == nil {
		t.Error("Expected error from deprecated NewClientFromConfig")
	}
}