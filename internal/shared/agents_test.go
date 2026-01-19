package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentModelConstants(t *testing.T) {
	assert.Equal(t, AgentModel("sonnet"), AgentModelSonnet)
	assert.Equal(t, AgentModel("opus"), AgentModelOpus)
	assert.Equal(t, AgentModel("haiku"), AgentModelHaiku)
	assert.Equal(t, AgentModel("inherit"), AgentModelInherit)
}

func TestAgentDefinition(t *testing.T) {
	def := AgentDefinition{
		Description:     "Test agent",
		Tools:           []string{"Read", "Write"},
		DisallowedTools: []string{"Bash"},
		Prompt:          "You are a test agent",
		Model:           AgentModelSonnet,
	}
	assert.Equal(t, "Test agent", def.Description)
	assert.Equal(t, []string{"Read", "Write"}, def.Tools)
	assert.Equal(t, []string{"Bash"}, def.DisallowedTools)
	assert.Equal(t, "You are a test agent", def.Prompt)
	assert.Equal(t, AgentModelSonnet, def.Model)
}
