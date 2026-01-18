// Package mcp provides in-process MCP (Model Context Protocol) server support.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// TypedToolHandler is a handler for typed MCP tool execution.
// T is the input type (must be a struct with json tags).
type TypedToolHandler[T any] func(ctx context.Context, input T) (any, error)

// TypedTool represents a type-safe MCP tool definition with generics.
type TypedTool[T any] struct {
	name        string
	description string
	handler     TypedToolHandler[T]
	inputSchema map[string]any
}

// Name returns the tool name.
func (t *TypedTool[T]) Name() string {
	return t.name
}

// Description returns the tool description.
func (t *TypedTool[T]) Description() string {
	return t.description
}

// InputSchema returns the JSON schema for the tool input.
func (t *TypedTool[T]) InputSchema() map[string]any {
	return t.inputSchema
}

// Execute invokes the handler with typed input converted from map.
func (t *TypedTool[T]) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	// Convert map to typed struct
	var input T
	data, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("marshal args: %w", err)
	}
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("unmarshal to type: %w", err)
	}

	// Call handler with panic recovery
	var result any
	var handlerErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				handlerErr = fmt.Errorf("handler panic: %v", r)
			}
		}()
		result, handlerErr = t.handler(ctx, input)
	}()

	if handlerErr != nil {
		return ErrorContent(handlerErr.Error()), nil
	}

	// If result is already a map, return it
	if m, ok := result.(map[string]any); ok {
		return m, nil
	}

	// Otherwise wrap in text content
	return TextContent(fmt.Sprintf("%v", result)), nil
}

// ToSdkMcpTool converts to the generic SdkMcpTool type for use with CreateSdkMcpServer.
func (t *TypedTool[T]) ToSdkMcpTool() *SdkMcpTool {
	return &SdkMcpTool{
		Name:        t.name,
		Description: t.description,
		InputSchema: t.inputSchema,
		Handler:     t.Execute,
	}
}

// schemaCache caches generated schemas by type to avoid repeated reflection.
var schemaCache sync.Map // map[reflect.Type]map[string]any

// NewTypedTool creates a type-safe MCP tool definition.
// Uses reflection to generate JSON schema from T's struct tags.
//
// Example:
//
//	type GreetInput struct {
//	    Name string `json:"name" jsonschema:"required,description=User name to greet"`
//	    Age  int    `json:"age,omitempty" jsonschema:"description=User age"`
//	}
//
//	greetTool := mcp.NewTypedTool("greet", "Greet a user",
//	    func(ctx context.Context, input GreetInput) (any, error) {
//	        return mcp.TextContent(fmt.Sprintf("Hello, %s!", input.Name)), nil
//	    })
//
//	server := mcp.CreateSdkMcpServer("my-tools", "1.0.0", []*mcp.SdkMcpTool{
//	    greetTool.ToSdkMcpTool(),
//	})
func NewTypedTool[T any](name, description string, handler TypedToolHandler[T]) *TypedTool[T] {
	var zero T
	t := reflect.TypeOf(zero)

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check cache first
	if cached, ok := schemaCache.Load(t); ok {
		return &TypedTool[T]{
			name:        name,
			description: description,
			handler:     handler,
			inputSchema: cached.(map[string]any),
		}
	}

	// Generate schema
	schema := generateSchema(t)

	// Cache the schema
	schemaCache.Store(t, schema)

	return &TypedTool[T]{
		name:        name,
		description: description,
		handler:     handler,
		inputSchema: schema,
	}
}

// generateSchema generates a JSON schema from a reflect.Type.
func generateSchema(t reflect.Type) map[string]any {
	if t.Kind() != reflect.Struct {
		// Non-struct types get a simple string schema
		return map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		}
	}

	properties := make(map[string]any)
	required := make([]string, 0)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		name, omitempty := parseJSONTag(jsonTag)
		if name == "" {
			name = field.Name
		}

		// Generate property schema
		propSchema := typeToJSONSchema(field.Type)

		// Parse jsonschema tag for additional metadata
		jsTag := field.Tag.Get("jsonschema")
		isRequired := !omitempty
		if jsTag != "" {
			parseJSONSchemaTag(jsTag, propSchema, &isRequired)
		}

		properties[name] = propSchema

		if isRequired {
			required = append(required, name)
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// parseJSONTag parses the json tag and returns the field name and omitempty flag.
func parseJSONTag(tag string) (name string, omitempty bool) {
	if tag == "" {
		return "", false
	}

	parts := strings.Split(tag, ",")
	name = parts[0]

	for _, part := range parts[1:] {
		if part == "omitempty" {
			omitempty = true
		}
	}

	return name, omitempty
}

// parseJSONSchemaTag parses the jsonschema tag and updates the schema.
// Supported directives: required, description=..., enum=a|b|c, min=N, max=N
func parseJSONSchemaTag(tag string, schema map[string]any, isRequired *bool) {
	parts := strings.Split(tag, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if part == "required" {
			*isRequired = true
			continue
		}

		if strings.HasPrefix(part, "description=") {
			schema["description"] = strings.TrimPrefix(part, "description=")
			continue
		}

		if strings.HasPrefix(part, "enum=") {
			enumStr := strings.TrimPrefix(part, "enum=")
			schema["enum"] = strings.Split(enumStr, "|")
			continue
		}

		if strings.HasPrefix(part, "min=") {
			var min float64
			fmt.Sscanf(strings.TrimPrefix(part, "min="), "%f", &min)
			schema["minimum"] = min
			continue
		}

		if strings.HasPrefix(part, "max=") {
			var max float64
			fmt.Sscanf(strings.TrimPrefix(part, "max="), "%f", &max)
			schema["maximum"] = max
			continue
		}
	}
}

// typeToJSONSchema converts a Go type to a JSON schema type.
func typeToJSONSchema(t reflect.Type) map[string]any {
	switch t.Kind() {
	case reflect.String:
		return map[string]any{"type": "string"}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}

	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}

	case reflect.Bool:
		return map[string]any{"type": "boolean"}

	case reflect.Slice, reflect.Array:
		return map[string]any{
			"type":  "array",
			"items": typeToJSONSchema(t.Elem()),
		}

	case reflect.Map:
		return map[string]any{
			"type":                 "object",
			"additionalProperties": typeToJSONSchema(t.Elem()),
		}

	case reflect.Ptr:
		return typeToJSONSchema(t.Elem())

	case reflect.Struct:
		return generateSchema(t)

	case reflect.Interface:
		return map[string]any{} // Any type

	default:
		return map[string]any{"type": "string"}
	}
}
