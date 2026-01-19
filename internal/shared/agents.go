package shared

// AgentModel represents the model to use for an agent.
type AgentModel string

const (
	AgentModelSonnet  AgentModel = "sonnet"
	AgentModelOpus    AgentModel = "opus"
	AgentModelHaiku   AgentModel = "haiku"
	AgentModelInherit AgentModel = "inherit"
)

// AgentDefinition defines a custom subagent.
type AgentDefinition struct {
	Description                        string               `json:"description"`
	Tools                              []string             `json:"tools,omitempty"`
	DisallowedTools                    []string             `json:"disallowedTools,omitempty"`
	Prompt                             string               `json:"prompt"`
	Model                              AgentModel           `json:"model,omitempty"`
	McpServers                         []AgentMcpServerSpec `json:"mcpServers,omitempty"`
	CriticalSystemReminderExperimental string               `json:"criticalSystemReminder_EXPERIMENTAL,omitempty"`
}

// AgentMcpServerSpec represents an MCP server specification for an agent.
// It can be either a simple string (server name) or a map of server configurations.
type AgentMcpServerSpec struct {
	Name    string                          `json:"name,omitempty"`
	Servers map[string]McpStdioServerConfig `json:"servers,omitempty"`
}
