package shared

import (
	"encoding/json"
	"reflect"
)

// marshalWithoutMethod marshals a struct value without calling its MarshalJSON method.
// It does this by creating a shallow copy with the same field values but without methods.
func marshalWithoutMethod(v any) ([]byte, error) {
	val := reflect.ValueOf(v)

	// Handle pointer indirection
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return []byte("null"), nil
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return json.Marshal(v)
	}

	// Create a map to hold the fields
	result := make(map[string]any)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON tag info
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		// Parse JSON tag
		name, opts := parseJSONTag(jsonTag)
		if name == "" {
			name = field.Name
		}

		// Handle omitempty
		if opts == "omitempty" && isEmptyValue(fieldVal) {
			continue
		}

		// Get the value, handling interfaces properly
		var value any
		if fieldVal.CanInterface() {
			value = fieldVal.Interface()
		}

		result[name] = value
	}

	return json.Marshal(result)
}

// parseJSONTag parses a JSON struct tag into name and options.
func parseJSONTag(tag string) (string, string) {
	if tag == "" {
		return "", ""
	}
	for i := 0; i < len(tag); i++ {
		if tag[i] == ',' {
			return tag[:i], tag[i+1:]
		}
	}
	return tag, ""
}

// isEmptyValue reports whether v is the zero value for its type.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	}
	return false
}

// MarshalWithType marshals a value with an explicit "type" field.
// This is a helper for message types that need to include their type in JSON output.
//
// The function marshals the struct fields without calling the MarshalJSON method,
// then adds the type field. This avoids infinite recursion.
//
// Example usage:
//
//	func (m *UserMessage) MarshalJSON() ([]byte, error) {
//	    return MarshalWithType(m, MessageTypeUser)
//	}
func MarshalWithType(v any, typeName string) ([]byte, error) {
	// Marshal without calling the MarshalJSON method
	data, err := marshalWithoutMethod(v)
	if err != nil {
		return nil, err
	}

	// Unmarshal into a map to add the type field
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	// Add the type field (this will override any existing "type" field)
	typeBytes, err := json.Marshal(typeName)
	if err != nil {
		return nil, err
	}
	m["type"] = typeBytes

	return json.Marshal(m)
}

// MarshalWithTypeAndSubtype marshals a value with explicit "type" and "subtype" fields.
// This is a helper for system message types that need both fields in JSON output.
//
// Example usage:
//
//	func (m *StatusMessage) MarshalJSON() ([]byte, error) {
//	    return MarshalWithTypeAndSubtype(m, MessageTypeSystem, SystemSubtypeStatus)
//	}
func MarshalWithTypeAndSubtype(v any, typeName, subtype string) ([]byte, error) {
	// Marshal without calling the MarshalJSON method
	data, err := marshalWithoutMethod(v)
	if err != nil {
		return nil, err
	}

	// Unmarshal into a map to add the type and subtype fields
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	// Add the type field
	typeBytes, err := json.Marshal(typeName)
	if err != nil {
		return nil, err
	}
	m["type"] = typeBytes

	// Add the subtype field
	subtypeBytes, err := json.Marshal(subtype)
	if err != nil {
		return nil, err
	}
	m["subtype"] = subtypeBytes

	return json.Marshal(m)
}
