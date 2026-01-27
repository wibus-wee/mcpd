package mcpcodec

import (
	"encoding/json"
	"strings"
)

// IsObjectSchema returns true when the schema type includes "object".
func IsObjectSchema(schema any) bool {
	if schema == nil {
		return false
	}

	switch value := schema.(type) {
	case map[string]any:
		return hasObjectType(value["type"])
	case json.RawMessage:
		return hasObjectTypeJSON(value)
	case []byte:
		return hasObjectTypeJSON(value)
	case string:
		return hasObjectTypeJSON([]byte(value))
	default:
		raw, err := json.Marshal(value)
		if err != nil {
			return false
		}
		return hasObjectTypeJSON(raw)
	}
}

type schemaTypeField struct {
	Type any `json:"type"`
}

func hasObjectTypeJSON(raw []byte) bool {
	if len(raw) == 0 {
		return false
	}
	var schema schemaTypeField
	if err := json.Unmarshal(raw, &schema); err != nil {
		return false
	}
	return hasObjectType(schema.Type)
}

func hasObjectType(value any) bool {
	switch typed := value.(type) {
	case string:
		return strings.EqualFold(typed, "object")
	case []any:
		for _, item := range typed {
			if str, ok := item.(string); ok && strings.EqualFold(str, "object") {
				return true
			}
		}
	case []string:
		for _, item := range typed {
			if strings.EqualFold(item, "object") {
				return true
			}
		}
	}
	return false
}
