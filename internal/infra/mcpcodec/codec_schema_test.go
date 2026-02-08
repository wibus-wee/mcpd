package mcpcodec

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsObjectSchema verifies schema type detection.
func TestIsObjectSchema(t *testing.T) {
	tests := []struct {
		name     string
		schema   any
		isObject bool
	}{
		{
			name:     "object schema detected from map",
			schema:   map[string]any{"type": "object"},
			isObject: true,
		},
		{
			name:     "object schema detected case-insensitive",
			schema:   map[string]any{"type": "Object"},
			isObject: true,
		},
		{
			name:     "array schema not detected",
			schema:   map[string]any{"type": "array"},
			isObject: false,
		},
		{
			name:     "string schema not detected",
			schema:   map[string]any{"type": "string"},
			isObject: false,
		},
		{
			name:     "object in type array detected",
			schema:   map[string]any{"type": []any{"string", "object"}},
			isObject: true,
		},
		{
			name:     "object in string array detected",
			schema:   map[string]any{"type": []string{"string", "object"}},
			isObject: true,
		},
		{
			name:     "nil schema returns false",
			schema:   nil,
			isObject: false,
		},
		{
			name:     "empty map returns false",
			schema:   map[string]any{},
			isObject: false,
		},
		{
			name:     "JSON string with object type",
			schema:   `{"type": "object"}`,
			isObject: true,
		},
		{
			name:     "JSON bytes with object type",
			schema:   []byte(`{"type": "object"}`),
			isObject: true,
		},
		{
			name:     "JSON RawMessage with object type",
			schema:   json.RawMessage(`{"type": "object"}`),
			isObject: true,
		},
		{
			name:     "invalid JSON returns false",
			schema:   `{invalid json}`,
			isObject: false,
		},
		{
			name:     "empty JSON bytes returns false",
			schema:   []byte{},
			isObject: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsObjectSchema(tt.schema)
			assert.Equal(t, tt.isObject, result)
		})
	}
}
