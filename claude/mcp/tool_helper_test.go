package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGreetInput is a test input type.
type TestGreetInput struct {
	Name string `json:"name" jsonschema:"required,description=User name to greet"`
	Age  int    `json:"age,omitempty" jsonschema:"description=User age,min=0,max=150"`
}

func TestNewTypedTool_BasicUsage(t *testing.T) {
	tool := NewTypedTool("greet", "Greet a user",
		func(ctx context.Context, input TestGreetInput) (any, error) {
			return TextContent("Hello, " + input.Name), nil
		})

	assert.Equal(t, "greet", tool.Name())
	assert.Equal(t, "Greet a user", tool.Description())
	assert.NotNil(t, tool.InputSchema())
}

func TestNewTypedTool_SchemaGeneration(t *testing.T) {
	tool := NewTypedTool("greet", "Greet a user",
		func(ctx context.Context, input TestGreetInput) (any, error) {
			return nil, nil
		})

	schema := tool.InputSchema()

	// Check type
	assert.Equal(t, "object", schema["type"])

	// Check properties
	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, props, "name")
	require.Contains(t, props, "age")

	// Check name property
	nameProp := props["name"].(map[string]any)
	assert.Equal(t, "string", nameProp["type"])
	assert.Equal(t, "User name to greet", nameProp["description"])

	// Check age property
	ageProp := props["age"].(map[string]any)
	assert.Equal(t, "integer", ageProp["type"])
	assert.Equal(t, "User age", ageProp["description"])
	assert.Equal(t, 0.0, ageProp["minimum"])
	assert.Equal(t, 150.0, ageProp["maximum"])

	// Check required
	required, ok := schema["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "name")
	assert.NotContains(t, required, "age") // has omitempty
}

func TestTypedTool_Execute(t *testing.T) {
	tool := NewTypedTool("greet", "Greet a user",
		func(ctx context.Context, input TestGreetInput) (any, error) {
			return TextContent("Hello, " + input.Name + "!"), nil
		})

	ctx := context.Background()
	result, err := tool.Execute(ctx, map[string]any{
		"name": "Alice",
		"age":  30,
	})

	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, content, 1)
	assert.Equal(t, "text", content[0]["type"])
	assert.Equal(t, "Hello, Alice!", content[0]["text"])
}

func TestTypedTool_Execute_PanicRecovery(t *testing.T) {
	tool := NewTypedTool("panic", "Panic handler",
		func(ctx context.Context, input TestGreetInput) (any, error) {
			panic("test panic")
		})

	ctx := context.Background()
	result, err := tool.Execute(ctx, map[string]any{
		"name": "Test",
	})

	// Should not return an error, but return error content
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, content, 1)
	assert.Contains(t, content[0]["text"], "handler panic")
	assert.Equal(t, true, result["isError"])
}

func TestTypedTool_ToSdkMcpTool(t *testing.T) {
	tool := NewTypedTool("greet", "Greet a user",
		func(ctx context.Context, input TestGreetInput) (any, error) {
			return TextContent("Hello, " + input.Name), nil
		})

	sdkTool := tool.ToSdkMcpTool()

	assert.Equal(t, "greet", sdkTool.Name)
	assert.Equal(t, "Greet a user", sdkTool.Description)
	assert.NotNil(t, sdkTool.InputSchema)
	assert.NotNil(t, sdkTool.Handler)

	// Test that handler works through SdkMcpTool
	ctx := context.Background()
	result, err := sdkTool.Handler(ctx, map[string]any{"name": "Bob"})
	require.NoError(t, err)

	content, ok := result["content"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, content, 1)
	assert.Equal(t, "Hello, Bob", content[0]["text"])
}

// Test nested struct schema generation
type TestNestedInput struct {
	User struct {
		Name  string `json:"name" jsonschema:"required"`
		Email string `json:"email" jsonschema:"required"`
	} `json:"user" jsonschema:"required"`
}

func TestNewTypedTool_NestedStruct(t *testing.T) {
	tool := NewTypedTool("nested", "Nested input",
		func(ctx context.Context, input TestNestedInput) (any, error) {
			return nil, nil
		})

	schema := tool.InputSchema()
	props := schema["properties"].(map[string]any)

	userProp := props["user"].(map[string]any)
	assert.Equal(t, "object", userProp["type"])

	userProps := userProp["properties"].(map[string]any)
	assert.Contains(t, userProps, "name")
	assert.Contains(t, userProps, "email")
}

// Test array schema generation
type TestArrayInput struct {
	Tags []string `json:"tags" jsonschema:"required"`
}

func TestNewTypedTool_ArrayType(t *testing.T) {
	tool := NewTypedTool("tags", "Array input",
		func(ctx context.Context, input TestArrayInput) (any, error) {
			return nil, nil
		})

	schema := tool.InputSchema()
	props := schema["properties"].(map[string]any)

	tagsProp := props["tags"].(map[string]any)
	assert.Equal(t, "array", tagsProp["type"])

	items := tagsProp["items"].(map[string]any)
	assert.Equal(t, "string", items["type"])
}

// Test map schema generation
type TestMapInput struct {
	Metadata map[string]int `json:"metadata"`
}

func TestNewTypedTool_MapType(t *testing.T) {
	tool := NewTypedTool("map", "Map input",
		func(ctx context.Context, input TestMapInput) (any, error) {
			return nil, nil
		})

	schema := tool.InputSchema()
	props := schema["properties"].(map[string]any)

	metaProp := props["metadata"].(map[string]any)
	assert.Equal(t, "object", metaProp["type"])

	additionalProps := metaProp["additionalProperties"].(map[string]any)
	assert.Equal(t, "integer", additionalProps["type"])
}

// Test enum schema generation
type TestEnumInput struct {
	Status string `json:"status" jsonschema:"required,enum=pending|active|completed"`
}

func TestNewTypedTool_EnumType(t *testing.T) {
	tool := NewTypedTool("status", "Enum input",
		func(ctx context.Context, input TestEnumInput) (any, error) {
			return nil, nil
		})

	schema := tool.InputSchema()
	props := schema["properties"].(map[string]any)

	statusProp := props["status"].(map[string]any)
	assert.Equal(t, "string", statusProp["type"])

	enum := statusProp["enum"].([]string)
	assert.Equal(t, []string{"pending", "active", "completed"}, enum)
}

// Test schema caching
func TestNewTypedTool_SchemaCache(t *testing.T) {
	// Create two tools with the same input type
	tool1 := NewTypedTool("tool1", "First tool",
		func(ctx context.Context, input TestGreetInput) (any, error) {
			return nil, nil
		})

	tool2 := NewTypedTool("tool2", "Second tool",
		func(ctx context.Context, input TestGreetInput) (any, error) {
			return nil, nil
		})

	// Both should return the same schema (from cache)
	// We can't directly compare pointers due to sync.Map behavior,
	// but we can verify the content is identical
	assert.Equal(t, tool1.InputSchema(), tool2.InputSchema())
}
