package agent

import "context"

// Executor is the interface for executing agent tasks.
// Implementations provide the core execution logic for agents, handling
// request processing and response generation.
type Executor interface {
	// Execute executes a task with the given request and returns a response.
	Execute(ctx context.Context, request *Request) (*Response, error)
}

// Request represents an execution request.
type Request struct {
	// Role is the role name to execute (e.g., "writer", "editor", "analyst").
	Role string

	// Prompt is the input prompt for the role execution.
	Prompt string

	// Temperature is the sampling temperature (optional).
	// If nil, the role's default temperature is used.
	Temperature *float32

	// MaxTokens is the maximum number of tokens to generate (optional).
	// If nil, the role's default max tokens is used.
	MaxTokens *int

	// Metadata contains additional execution context (optional).
	Metadata map[string]string
}

// Response represents an execution response.
type Response struct {
	// Content is the generated content from the role execution.
	Content string

	// Metadata contains additional response information (optional).
	Metadata map[string]string
}
