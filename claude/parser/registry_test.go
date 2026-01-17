package parser

import (
	"sync"
	"testing"

	"github.com/dotcommander/agent-sdk-go/claude/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMessageParserRegistry tests registry creation and defaults.
func TestNewMessageParserRegistry(t *testing.T) {
	t.Parallel()

	registry := NewMessageParserRegistry()
	require.NotNil(t, registry)

	// Verify default parsers are registered
	expectedTypes := []string{
		shared.MessageTypeUser,
		shared.MessageTypeAssistant,
		shared.MessageTypeSystem,
		shared.MessageTypeResult,
		shared.MessageTypeStreamEvent,
		shared.MessageTypeControlRequest,
		shared.MessageTypeControlResponse,
		shared.MessageTypeToolProgress,
		shared.MessageTypeAuthStatus,
	}

	registeredTypes := registry.RegisteredTypes()
	assert.Equal(t, len(expectedTypes), len(registeredTypes), "should have all default parsers")

	// Check each expected type is registered
	for _, expectedType := range expectedTypes {
		assert.True(t, registry.HasParser(expectedType),
			"should have parser for type: %s", expectedType)
	}
}

// TestRegistryRegister tests registering custom parsers.
func TestRegistryRegister(t *testing.T) {
	t.Parallel()

	registry := NewMessageParserRegistry()

	// Create a custom parser
	customParser := func(jsonStr string, lineNumber int) (shared.Message, error) {
		return &shared.UserMessage{
			MessageType: shared.MessageTypeUser,
			Content:     "custom",
		}, nil
	}

	// Register it
	registry.Register("custom_type", customParser)

	// Verify it's registered
	assert.True(t, registry.HasParser("custom_type"))

	// Parse using it
	msg, err := registry.Parse("custom_type", `{"type":"custom"}`, 1)
	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.IsType(t, (*shared.UserMessage)(nil), msg)
}

// TestRegistryUnregister tests unregistering parsers.
func TestRegistryUnregister(t *testing.T) {
	t.Parallel()

	registry := NewMessageParserRegistry()

	// User type should exist
	assert.True(t, registry.HasParser(shared.MessageTypeUser))

	// Unregister it
	registry.Unregister(shared.MessageTypeUser)

	// Verify it's gone
	assert.False(t, registry.HasParser(shared.MessageTypeUser))
}

// TestRegistryParseUnregisteredType tests parsing with missing parser.
func TestRegistryParseUnregisteredType(t *testing.T) {
	t.Parallel()

	registry := NewMessageParserRegistry()

	// Remove user parser
	registry.Unregister(shared.MessageTypeUser)

	// Try to parse - should fail
	msg, err := registry.Parse(shared.MessageTypeUser, `{"type":"user"}`, 1)
	assert.Error(t, err)
	assert.Nil(t, msg)
	assert.Contains(t, err.Error(), "unknown message type")
}

// TestRegistryParserReplacement tests replacing existing parser.
func TestRegistryParserReplacement(t *testing.T) {
	t.Parallel()

	registry := NewMessageParserRegistry()

	// Create a custom parser that always fails
	customParser := func(jsonStr string, lineNumber int) (shared.Message, error) {
		return nil, shared.NewParserError(lineNumber, 0, jsonStr, "custom error")
	}

	// Replace user parser
	registry.Register(shared.MessageTypeUser, customParser)

	// Try to parse - should get custom error
	msg, err := registry.Parse(shared.MessageTypeUser, `{"type":"user","content":"test"}`, 1)
	assert.Error(t, err)
	assert.Nil(t, msg)
	assert.Contains(t, err.Error(), "custom error")
}

// TestRegistryConcurrentRegister tests thread-safe registration.
func TestRegistryConcurrentRegister(t *testing.T) {
	t.Parallel()

	registry := NewMessageParserRegistry()
	var wg sync.WaitGroup

	// Register 10 custom types concurrently
	for i := range 10 {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			parser := func(jsonStr string, lineNumber int) (shared.Message, error) {
				return &shared.UserMessage{}, nil
			}

			typeID := "" + string(rune(48+index%10))
			registry.Register("concurrent_"+typeID, parser)
		}(i)
	}

	wg.Wait()

	// Verify at least some were registered
	types := registry.RegisteredTypes()
	assert.Greater(t, len(types), 7, "should have original + concurrent types")
}

// TestRegistryConcurrentParse tests thread-safe parsing.
func TestRegistryConcurrentParse(t *testing.T) {
	t.Parallel()

	registry := NewMessageParserRegistry()
	var wg sync.WaitGroup

	results := make([]error, 10)
	for i := range 10 {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			jsonStr := `{"type":"user","content":"Hello"}`
			msg, err := registry.Parse(shared.MessageTypeUser, jsonStr, index)

			results[index] = err
			if msg == nil {
				results[index] = shared.NewParserError(index, 0, jsonStr, "got nil")
			}
		}(i)
	}

	wg.Wait()

	// All should succeed
	for i, err := range results {
		assert.NoError(t, err, "parse %d should succeed", i)
	}
}

// TestRegistryDefaultRegistry tests the global default registry.
func TestRegistryDefaultRegistry(t *testing.T) {
	t.Parallel()

	registry := DefaultRegistry()
	require.NotNil(t, registry)

	// Should have all default parsers
	assert.True(t, registry.HasParser(shared.MessageTypeUser))
	assert.True(t, registry.HasParser(shared.MessageTypeAssistant))
	assert.True(t, registry.HasParser(shared.MessageTypeSystem))
	assert.True(t, registry.HasParser(shared.MessageTypeResult))
}

// TestRegistryIntegrationWithParser tests registry integration with Parser.
func TestRegistryIntegrationWithParser(t *testing.T) {
	t.Parallel()

	registry := NewMessageParserRegistry()

	// Test user message parsing
	jsonStr := `{
		"type": "user",
		"content": "Hello, Claude!"
	}`

	msg, err := registry.Parse(shared.MessageTypeUser, jsonStr, 1)
	require.NoError(t, err)
	require.NotNil(t, msg)

	userMsg, ok := msg.(*shared.UserMessage)
	require.True(t, ok)
	assert.Equal(t, "Hello, Claude!", userMsg.Content)

	// Test assistant message parsing
	jsonStr = `{
		"type": "assistant",
		"content": [{"type":"text","text":"Hi!"}],
		"model": "claude-3-5-sonnet-20241022"
	}`

	msg, err = registry.Parse(shared.MessageTypeAssistant, jsonStr, 2)
	require.NoError(t, err)
	require.NotNil(t, msg)

	assistantMsg, ok := msg.(*shared.AssistantMessage)
	require.True(t, ok)
	assert.Equal(t, 1, len(assistantMsg.Content))
}

// BenchmarkRegistryParse benchmarks registry parsing.
func BenchmarkRegistryParse(b *testing.B) {
	registry := NewMessageParserRegistry()
	jsonStr := `{
		"type": "user",
		"content": "Hello, Claude!"
	}`

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := registry.Parse(shared.MessageTypeUser, jsonStr, 1)
			if err != nil {
				b.Error(err)
			}
		}
	})
}
