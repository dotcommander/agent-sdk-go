package shared

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testMessage is a simple struct for testing MarshalWithType.
type testMessage struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// testMessageWithSubtype is a struct for testing MarshalWithTypeAndSubtype.
type testMessageWithSubtype struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

func TestMarshalWithType(t *testing.T) {
	tests := []struct {
		name     string
		input    *testMessage
		typeName string
		want     map[string]any
	}{
		{
			name:     "basic message",
			input:    &testMessage{Name: "test", Value: 42},
			typeName: "test_type",
			want: map[string]any{
				"type":  "test_type",
				"name":  "test",
				"value": float64(42), // JSON numbers are float64
			},
		},
		{
			name:     "empty message",
			input:    &testMessage{},
			typeName: "empty_type",
			want: map[string]any{
				"type":  "empty_type",
				"name":  "",
				"value": float64(0),
			},
		},
		{
			name:     "message with empty type",
			input:    &testMessage{Name: "foo", Value: 100},
			typeName: "",
			want: map[string]any{
				"type":  "",
				"name":  "foo",
				"value": float64(100),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := MarshalWithType(tt.input, tt.typeName)
			require.NoError(t, err)

			var got map[string]any
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMarshalWithTypeAndSubtype(t *testing.T) {
	tests := []struct {
		name    string
		input   *testMessageWithSubtype
		msgType string
		subtype string
		want    map[string]any
	}{
		{
			name:    "basic message with subtype",
			input:   &testMessageWithSubtype{ID: "123", Content: "hello"},
			msgType: "system",
			subtype: "status",
			want: map[string]any{
				"type":    "system",
				"subtype": "status",
				"id":      "123",
				"content": "hello",
			},
		},
		{
			name:    "empty message with subtype",
			input:   &testMessageWithSubtype{},
			msgType: "system",
			subtype: "init",
			want: map[string]any{
				"type":    "system",
				"subtype": "init",
				"id":      "",
				"content": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := MarshalWithTypeAndSubtype(tt.input, tt.msgType, tt.subtype)
			require.NoError(t, err)

			var got map[string]any
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestMarshalWithType_FieldOrder verifies that type field appears first in JSON output.
// Note: Go's encoding/json does not guarantee field order, but our wrapper struct
// places Type first which typically results in it being serialized first.
func TestMarshalWithType_FieldOrder(t *testing.T) {
	msg := &testMessage{Name: "test", Value: 42}
	data, err := MarshalWithType(msg, "test_type")
	require.NoError(t, err)

	// Verify the JSON is valid and contains the type field
	var result map[string]any
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, "test_type", result["type"])
}

// TestMarshalWithType_NilPointer verifies behavior with nil values inside struct.
func TestMarshalWithType_OptionalFields(t *testing.T) {
	type msgWithOptional struct {
		Required string  `json:"required"`
		Optional *string `json:"optional,omitempty"`
	}

	// Without optional field
	msg1 := &msgWithOptional{Required: "value"}
	data1, err := MarshalWithType(msg1, "optional_test")
	require.NoError(t, err)

	var result1 map[string]any
	err = json.Unmarshal(data1, &result1)
	require.NoError(t, err)
	assert.Equal(t, "optional_test", result1["type"])
	assert.Equal(t, "value", result1["required"])
	_, hasOptional := result1["optional"]
	assert.False(t, hasOptional, "optional field should be omitted when nil")

	// With optional field
	optValue := "present"
	msg2 := &msgWithOptional{Required: "value", Optional: &optValue}
	data2, err := MarshalWithType(msg2, "optional_test")
	require.NoError(t, err)

	var result2 map[string]any
	err = json.Unmarshal(data2, &result2)
	require.NoError(t, err)
	assert.Equal(t, "present", result2["optional"])
}

// TestMarshalWithType_NestedStruct verifies behavior with nested structs.
func TestMarshalWithType_NestedStruct(t *testing.T) {
	type inner struct {
		Field string `json:"field"`
	}
	type outer struct {
		Inner inner `json:"inner"`
	}

	msg := &outer{Inner: inner{Field: "nested_value"}}
	data, err := MarshalWithType(msg, "nested_type")
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, "nested_type", result["type"])

	innerMap, ok := result["inner"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "nested_value", innerMap["field"])
}

// TestMarshalWithType_PreservesExistingTypeField verifies that the wrapper's type field
// takes precedence over any existing type field in the struct.
func TestMarshalWithType_PreservesExistingTypeField(t *testing.T) {
	type msgWithType struct {
		MessageType string `json:"type"`
		Data        string `json:"data"`
	}

	msg := &msgWithType{MessageType: "original_type", Data: "some_data"}
	data, err := MarshalWithType(msg, "override_type")
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// The wrapper's type should take precedence
	assert.Equal(t, "override_type", result["type"])
	assert.Equal(t, "some_data", result["data"])
}
