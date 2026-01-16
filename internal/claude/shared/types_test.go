package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountInfo(t *testing.T) {
	info := AccountInfo{
		Email:            "test@example.com",
		Organization:     "Test Org",
		SubscriptionType: "pro",
	}
	assert.Equal(t, "test@example.com", info.Email)
	assert.Equal(t, "Test Org", info.Organization)
	assert.Equal(t, "pro", info.SubscriptionType)
}

func TestModelInfo(t *testing.T) {
	info := ModelInfo{
		Value:       "claude-sonnet-4",
		DisplayName: "Claude Sonnet 4",
		Description: "Latest Sonnet model",
	}
	assert.Equal(t, "claude-sonnet-4", info.Value)
	assert.Equal(t, "Claude Sonnet 4", info.DisplayName)
	assert.Equal(t, "Latest Sonnet model", info.Description)
}

func TestSlashCommand(t *testing.T) {
	cmd := SlashCommand{
		Name:         "commit",
		Description:  "Commit changes",
		ArgumentHint: "<message>",
	}
	assert.Equal(t, "commit", cmd.Name)
	assert.Equal(t, "Commit changes", cmd.Description)
	assert.Equal(t, "<message>", cmd.ArgumentHint)
}

func TestRewindFilesResult(t *testing.T) {
	result := RewindFilesResult{
		CanRewind:    true,
		FilesChanged: []string{"file1.go", "file2.go"},
		Insertions:   10,
		Deletions:    5,
	}
	assert.True(t, result.CanRewind)
	assert.Len(t, result.FilesChanged, 2)
	assert.Equal(t, 10, result.Insertions)
	assert.Equal(t, 5, result.Deletions)
}

func TestApiKeySourceConstants(t *testing.T) {
	assert.Equal(t, ApiKeySource("user"), ApiKeySourceUser)
	assert.Equal(t, ApiKeySource("project"), ApiKeySourceProject)
	assert.Equal(t, ApiKeySource("org"), ApiKeySourceOrg)
	assert.Equal(t, ApiKeySource("temporary"), ApiKeySourceTemporary)
}

func TestConfigScopeConstants(t *testing.T) {
	assert.Equal(t, ConfigScope("local"), ConfigScopeLocal)
	assert.Equal(t, ConfigScope("user"), ConfigScopeUser)
	assert.Equal(t, ConfigScope("project"), ConfigScopeProject)
}
