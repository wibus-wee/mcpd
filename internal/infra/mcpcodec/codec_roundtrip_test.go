package mcpcodec

import (
	"encoding/json"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mcpv/internal/domain"
)

const toolDefinitionSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["name", "inputSchema"],
  "properties": {
    "_meta": { "$ref": "#/$defs/meta" },
    "annotations": { "$ref": "#/$defs/toolAnnotations" },
    "description": { "type": "string" },
    "inputSchema": { "type": "object" },
    "name": { "type": "string" },
    "outputSchema": { "type": "object" },
    "title": { "type": "string" },
    "icons": { "type": "array", "items": { "$ref": "#/$defs/icon" } }
  },
  "additionalProperties": true,
  "$defs": {
    "meta": {
      "type": "object"
    },
    "toolAnnotations": {
      "type": "object",
      "properties": {
        "idempotentHint": { "type": "boolean" },
        "readOnlyHint": { "type": "boolean" },
        "destructiveHint": { "type": ["boolean", "null"] },
        "openWorldHint": { "type": ["boolean", "null"] },
        "title": { "type": "string" }
      },
      "additionalProperties": true
    },
    "icon": {
      "type": "object",
      "required": ["src"],
      "properties": {
        "src": { "type": "string" },
        "mimeType": { "type": "string" },
        "sizes": { "type": "array", "items": { "type": "string" } },
        "theme": { "type": "string" }
      },
      "additionalProperties": true
    }
  }
}`

const resourceDefinitionSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["uri", "name"],
  "properties": {
    "_meta": { "$ref": "#/$defs/meta" },
    "annotations": { "$ref": "#/$defs/annotations" },
    "description": { "type": "string" },
    "mimeType": { "type": "string" },
    "name": { "type": "string" },
    "size": { "type": "integer" },
    "title": { "type": "string" },
    "uri": { "type": "string" },
    "icons": { "type": "array", "items": { "$ref": "#/$defs/icon" } }
  },
  "additionalProperties": true,
  "$defs": {
    "meta": {
      "type": "object"
    },
    "annotations": {
      "type": "object",
      "properties": {
        "audience": { "type": "array", "items": { "type": "string" } },
        "lastModified": { "type": "string" },
        "priority": { "type": "number" }
      },
      "additionalProperties": true
    },
    "icon": {
      "type": "object",
      "required": ["src"],
      "properties": {
        "src": { "type": "string" },
        "mimeType": { "type": "string" },
        "sizes": { "type": "array", "items": { "type": "string" } },
        "theme": { "type": "string" }
      },
      "additionalProperties": true
    }
  }
}`

const promptDefinitionSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["name"],
  "properties": {
    "_meta": { "$ref": "#/$defs/meta" },
    "arguments": { "type": "array", "items": { "$ref": "#/$defs/promptArgument" } },
    "description": { "type": "string" },
    "name": { "type": "string" },
    "title": { "type": "string" },
    "icons": { "type": "array", "items": { "$ref": "#/$defs/icon" } }
  },
  "additionalProperties": true,
  "$defs": {
    "meta": {
      "type": "object"
    },
    "promptArgument": {
      "type": "object",
      "required": ["name"],
      "properties": {
        "name": { "type": "string" },
        "title": { "type": "string" },
        "description": { "type": "string" },
        "required": { "type": "boolean" }
      },
      "additionalProperties": true
    },
    "icon": {
      "type": "object",
      "required": ["src"],
      "properties": {
        "src": { "type": "string" },
        "mimeType": { "type": "string" },
        "sizes": { "type": "array", "items": { "type": "string" } },
        "theme": { "type": "string" }
      },
      "additionalProperties": true
    }
  }
}`

func validateAgainstSchema(t *testing.T, schemaJSON string, payload []byte) {
	t.Helper()

	var schema jsonschema.Schema
	require.NoError(t, json.Unmarshal([]byte(schemaJSON), &schema))

	resolved, err := schema.Resolve(nil)
	require.NoError(t, err)

	var decoded any
	require.NoError(t, json.Unmarshal(payload, &decoded))
	require.NoError(t, resolved.Validate(decoded))
}

func TestRoundTrip_ToolDefinition(t *testing.T) {
	boolTrue := true
	boolFalse := false

	original := domain.ToolDefinition{
		Name:        "test_tool",
		Description: "A comprehensive test tool",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"param1": map[string]any{"type": "string"},
				"param2": map[string]any{"type": "number"},
			},
			"required": []any{"param1"},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"result": map[string]any{"type": "string"},
			},
		},
		Title: "Test Tool",
		Annotations: &domain.ToolAnnotations{
			IdempotentHint:  true,
			ReadOnlyHint:    false,
			DestructiveHint: &boolFalse,
			OpenWorldHint:   &boolTrue,
			Title:           "Annotated Tool",
		},
		Meta: domain.Meta{
			"version": "1.0.0",
			"author":  "test",
		},
	}

	data, err := MarshalToolDefinition(original)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	validateAgainstSchema(t, toolDefinitionSchema, data)

	var mcpTool mcp.Tool
	err = json.Unmarshal(data, &mcpTool)
	require.NoError(t, err)

	roundtripped := ToolFromMCP(&mcpTool)

	assert.Equal(t, original.Name, roundtripped.Name)
	assert.Equal(t, original.Description, roundtripped.Description)
	assert.Equal(t, original.Title, roundtripped.Title)

	originalInputJSON, _ := json.Marshal(original.InputSchema)
	roundtrippedInputJSON, _ := json.Marshal(roundtripped.InputSchema)
	assert.JSONEq(t, string(originalInputJSON), string(roundtrippedInputJSON))

	if original.OutputSchema != nil {
		originalOutputJSON, _ := json.Marshal(original.OutputSchema)
		roundtrippedOutputJSON, _ := json.Marshal(roundtripped.OutputSchema)
		assert.JSONEq(t, string(originalOutputJSON), string(roundtrippedOutputJSON))
	}

	if original.Annotations != nil {
		require.NotNil(t, roundtripped.Annotations)
		assert.Equal(t, original.Annotations.IdempotentHint, roundtripped.Annotations.IdempotentHint)
		assert.Equal(t, original.Annotations.ReadOnlyHint, roundtripped.Annotations.ReadOnlyHint)
		assert.Equal(t, *original.Annotations.DestructiveHint, *roundtripped.Annotations.DestructiveHint)
		assert.Equal(t, *original.Annotations.OpenWorldHint, *roundtripped.Annotations.OpenWorldHint)
		assert.Equal(t, original.Annotations.Title, roundtripped.Annotations.Title)
	}

	assert.Equal(t, original.Meta, roundtripped.Meta)

	hash1, err := HashToolDefinition(original)
	require.NoError(t, err)
	hash2, err := HashToolDefinition(roundtripped)
	require.NoError(t, err)
	assert.Equal(t, hash1, hash2, "Hash should be same after round-trip")
}

func TestRoundTrip_ResourceDefinition(t *testing.T) {
	original := domain.ResourceDefinition{
		URI:         "file:///test/resource.txt",
		Name:        "test_resource",
		Title:       "Test Resource",
		Description: "A test resource",
		MIMEType:    "text/plain",
		Size:        1024,
		Annotations: &domain.Annotations{
			Audience:     []domain.Role{"user", "assistant"},
			LastModified: "2024-01-01T00:00:00Z",
			Priority:     1.0,
		},
		Meta: domain.Meta{
			"version": "1.0.0",
		},
	}

	data, err := MarshalResourceDefinition(original)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	validateAgainstSchema(t, resourceDefinitionSchema, data)

	var mcpResource mcp.Resource
	err = json.Unmarshal(data, &mcpResource)
	require.NoError(t, err)

	roundtripped := ResourceFromMCP(&mcpResource)

	assert.Equal(t, original.URI, roundtripped.URI)
	assert.Equal(t, original.Name, roundtripped.Name)
	assert.Equal(t, original.Title, roundtripped.Title)
	assert.Equal(t, original.Description, roundtripped.Description)
	assert.Equal(t, original.MIMEType, roundtripped.MIMEType)
	assert.Equal(t, original.Size, roundtripped.Size)

	if original.Annotations != nil {
		require.NotNil(t, roundtripped.Annotations)
		assert.Equal(t, original.Annotations.Audience, roundtripped.Annotations.Audience)
		assert.Equal(t, original.Annotations.LastModified, roundtripped.Annotations.LastModified)
		assert.Equal(t, original.Annotations.Priority, roundtripped.Annotations.Priority)
	}

	assert.Equal(t, original.Meta, roundtripped.Meta)

	hash1, err := HashResourceDefinition(original)
	require.NoError(t, err)
	hash2, err := HashResourceDefinition(roundtripped)
	require.NoError(t, err)
	assert.Equal(t, hash1, hash2, "Hash should be same after round-trip")
}

func TestRoundTrip_PromptDefinition(t *testing.T) {
	original := domain.PromptDefinition{
		Name:        "test_prompt",
		Title:       "Test Prompt",
		Description: "A test prompt",
		Arguments: []domain.PromptArgument{
			{
				Name:        "arg1",
				Title:       "Argument 1",
				Description: "First argument",
				Required:    true,
			},
			{
				Name:        "arg2",
				Title:       "Argument 2",
				Description: "Second argument",
				Required:    false,
			},
		},
		Meta: domain.Meta{
			"version": "1.0.0",
		},
	}

	data, err := MarshalPromptDefinition(original)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	validateAgainstSchema(t, promptDefinitionSchema, data)

	var mcpPrompt mcp.Prompt
	err = json.Unmarshal(data, &mcpPrompt)
	require.NoError(t, err)

	roundtripped := PromptFromMCP(&mcpPrompt)

	assert.Equal(t, original.Name, roundtripped.Name)
	assert.Equal(t, original.Title, roundtripped.Title)
	assert.Equal(t, original.Description, roundtripped.Description)

	require.Equal(t, len(original.Arguments), len(roundtripped.Arguments))
	for i, arg := range original.Arguments {
		assert.Equal(t, arg.Name, roundtripped.Arguments[i].Name)
		assert.Equal(t, arg.Title, roundtripped.Arguments[i].Title)
		assert.Equal(t, arg.Description, roundtripped.Arguments[i].Description)
		assert.Equal(t, arg.Required, roundtripped.Arguments[i].Required)
	}

	assert.Equal(t, original.Meta, roundtripped.Meta)

	hash1, err := HashPromptDefinition(original)
	require.NoError(t, err)
	hash2, err := HashPromptDefinition(roundtripped)
	require.NoError(t, err)
	assert.Equal(t, hash1, hash2, "Hash should be same after round-trip")
}
