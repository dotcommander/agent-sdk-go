package shared

// ModelUsage represents token usage and cost for a model.
type ModelUsage struct {
	InputTokens              int     `json:"inputTokens"`
	OutputTokens             int     `json:"outputTokens"`
	CacheReadInputTokens     int     `json:"cacheReadInputTokens"`
	CacheCreationInputTokens int     `json:"cacheCreationInputTokens"`
	WebSearchRequests        int     `json:"webSearchRequests"`
	CostUSD                  float64 `json:"costUSD"`
	ContextWindow            int     `json:"contextWindow"`
	MaxOutputTokens          int     `json:"maxOutputTokens"`
}

// ResultSubtype constants for result message discrimination.
const (
	ResultSubtypeSuccess                         = "success"
	ResultSubtypeErrorDuringExecution            = "error_during_execution"
	ResultSubtypeErrorMaxTurns                   = "error_max_turns"
	ResultSubtypeErrorMaxBudgetUSD               = "error_max_budget_usd"
	ResultSubtypeErrorMaxStructuredOutputRetries = "error_max_structured_output_retries"
)

// SDKPermissionDenial represents a denied permission request.
type SDKPermissionDenial struct {
	ToolName  string         `json:"tool_name"`
	ToolUseID string         `json:"tool_use_id"`
	ToolInput map[string]any `json:"tool_input"`
}
