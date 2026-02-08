package mcpcodec

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mcpv/internal/domain"
)

// TestToolFromMCP_Conversion verifies MCP to domain conversion.
func TestToolFromMCP_Conversion(t *testing.T) {
	boolTrue := true
	boolFalse := false

	mcpTool := &mcp.Tool{
		Name:        "test_tool",
		Description: "Test description",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"param": map[string]any{"type": "string"},
			},
		},
		OutputSchema: map[string]any{"type": "string"},
		Title:        "Test Tool",
		Annotations: &mcp.ToolAnnotations{
			IdempotentHint:  true,
			ReadOnlyHint:    false,
			DestructiveHint: &boolFalse,
			OpenWorldHint:   &boolTrue,
			Title:           "Annotated",
		},
		Meta: mcp.Meta{
			"version": "1.0",
		},
	}

	result := ToolFromMCP(mcpTool)

	assert.Equal(t, "test_tool", result.Name)
	assert.Equal(t, "Test description", result.Description)
	assert.Equal(t, "Test Tool", result.Title)
	assert.NotNil(t, result.InputSchema)
	assert.NotNil(t, result.OutputSchema)
	assert.NotNil(t, result.Annotations)
	assert.True(t, result.Annotations.IdempotentHint)
	assert.False(t, result.Annotations.ReadOnlyHint)
	assert.NotNil(t, result.Annotations.DestructiveHint)
	assert.False(t, *result.Annotations.DestructiveHint)
	assert.NotNil(t, result.Annotations.OpenWorldHint)
	assert.True(t, *result.Annotations.OpenWorldHint)
	assert.NotNil(t, result.Meta)
}

// TestResourceFromMCP_Conversion verifies MCP to domain conversion.
func TestResourceFromMCP_Conversion(t *testing.T) {
	mcpResource := &mcp.Resource{
		URI:         "file:///test.txt",
		Name:        "test_resource",
		Title:       "Test Resource",
		Description: "Test description",
		MIMEType:    "text/plain",
		Size:        1024,
		Annotations: &mcp.Annotations{
			Audience:     []mcp.Role{"user"},
			LastModified: "2024-01-01T00:00:00Z",
			Priority:     1.0,
		},
		Meta: mcp.Meta{
			"version": "1.0",
		},
	}

	result := ResourceFromMCP(mcpResource)

	assert.Equal(t, "file:///test.txt", result.URI)
	assert.Equal(t, "test_resource", result.Name)
	assert.Equal(t, "Test Resource", result.Title)
	assert.Equal(t, "Test description", result.Description)
	assert.Equal(t, "text/plain", result.MIMEType)
	assert.Equal(t, int64(1024), result.Size)
	assert.NotNil(t, result.Annotations)
	assert.Len(t, result.Annotations.Audience, 1)
	assert.Equal(t, domain.Role("user"), result.Annotations.Audience[0])
	assert.NotNil(t, result.Meta)
}

// TestPromptFromMCP_Conversion verifies MCP to domain conversion.
func TestPromptFromMCP_Conversion(t *testing.T) {
	mcpPrompt := &mcp.Prompt{
		Name:        "test_prompt",
		Title:       "Test Prompt",
		Description: "Test description",
		Arguments: []*mcp.PromptArgument{
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
		Meta: mcp.Meta{
			"version": "1.0",
		},
	}

	result := PromptFromMCP(mcpPrompt)

	assert.Equal(t, "test_prompt", result.Name)
	assert.Equal(t, "Test Prompt", result.Title)
	assert.Equal(t, "Test description", result.Description)
	assert.Len(t, result.Arguments, 2)
	assert.Equal(t, "arg1", result.Arguments[0].Name)
	assert.True(t, result.Arguments[0].Required)
	assert.Equal(t, "arg2", result.Arguments[1].Name)
	assert.False(t, result.Arguments[1].Required)
	assert.NotNil(t, result.Meta)
}

// TestPromptFromMCP_NilArguments verifies nil argument handling.
func TestPromptFromMCP_NilArguments(t *testing.T) {
	mcpPrompt := &mcp.Prompt{
		Name:      "test_prompt",
		Arguments: []*mcp.PromptArgument{nil, {Name: "valid"}},
	}

	result := PromptFromMCP(mcpPrompt)

	assert.Len(t, result.Arguments, 1)
	assert.Equal(t, "valid", result.Arguments[0].Name)
}

// TestMetaConversion verifies meta field conversion.
func TestMetaConversion(t *testing.T) {
	t.Run("nil meta returns nil", func(t *testing.T) {
		tool := domain.ToolDefinition{
			Name: "test",
			Meta: nil,
		}
		data, err := MarshalToolDefinition(tool)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		var mcpTool mcp.Tool
		require.NoError(t, json.Unmarshal(data, &mcpTool))
		assert.Nil(t, mcpTool.Meta)

		roundtripped := ToolFromMCP(&mcpTool)
		assert.Nil(t, roundtripped.Meta)
	})

	t.Run("meta with nested objects", func(t *testing.T) {
		expected := domain.Meta{
			"nested": map[string]any{
				"key": "value",
			},
		}
		tool := domain.ToolDefinition{
			Name: "test",
			Meta: expected,
		}
		data, err := MarshalToolDefinition(tool)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		var mcpTool mcp.Tool
		require.NoError(t, json.Unmarshal(data, &mcpTool))

		roundtripped := ToolFromMCP(&mcpTool)
		require.NotNil(t, roundtripped.Meta)
		assert.Equal(t, expected, roundtripped.Meta)
	})
}

// TestAnnotationsConversion verifies annotations conversion.
func TestAnnotationsConversion(t *testing.T) {
	t.Run("nil annotations handled", func(t *testing.T) {
		resource := domain.ResourceDefinition{
			URI:         "file:///test",
			Annotations: nil,
		}
		data, err := MarshalResourceDefinition(resource)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		var mcpResource mcp.Resource
		require.NoError(t, json.Unmarshal(data, &mcpResource))
		assert.Nil(t, mcpResource.Annotations)

		roundtripped := ResourceFromMCP(&mcpResource)
		assert.Nil(t, roundtripped.Annotations)
	})

	t.Run("empty audience handled", func(t *testing.T) {
		resource := domain.ResourceDefinition{
			URI: "file:///test",
			Annotations: &domain.Annotations{
				Audience: []domain.Role{},
			},
		}
		data, err := MarshalResourceDefinition(resource)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		var mcpResource mcp.Resource
		require.NoError(t, json.Unmarshal(data, &mcpResource))
		require.NotNil(t, mcpResource.Annotations)

		roundtripped := ResourceFromMCP(&mcpResource)
		require.NotNil(t, roundtripped.Annotations)
		assert.Empty(t, roundtripped.Annotations.Audience)
	})
}

// TestToolAnnotationsConversion verifies tool annotations conversion.
func TestToolAnnotationsConversion(t *testing.T) {
	t.Run("nil tool annotations handled", func(t *testing.T) {
		tool := domain.ToolDefinition{
			Name:        "test",
			Annotations: nil,
		}
		data, err := MarshalToolDefinition(tool)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		var mcpTool mcp.Tool
		require.NoError(t, json.Unmarshal(data, &mcpTool))
		assert.Nil(t, mcpTool.Annotations)

		roundtripped := ToolFromMCP(&mcpTool)
		assert.Nil(t, roundtripped.Annotations)
	})

	t.Run("tool annotations with nil hints", func(t *testing.T) {
		tool := domain.ToolDefinition{
			Name: "test",
			Annotations: &domain.ToolAnnotations{
				IdempotentHint:  true,
				DestructiveHint: nil,
				OpenWorldHint:   nil,
			},
		}
		data, err := MarshalToolDefinition(tool)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		var mcpTool mcp.Tool
		require.NoError(t, json.Unmarshal(data, &mcpTool))

		roundtripped := ToolFromMCP(&mcpTool)
		require.NotNil(t, roundtripped.Annotations)
		assert.True(t, roundtripped.Annotations.IdempotentHint)
		assert.Nil(t, roundtripped.Annotations.DestructiveHint)
		assert.Nil(t, roundtripped.Annotations.OpenWorldHint)
	})
}

// TestPromptArgumentsConversion verifies prompt arguments conversion.
func TestPromptArgumentsConversion(t *testing.T) {
	t.Run("empty arguments handled", func(t *testing.T) {
		prompt := domain.PromptDefinition{
			Name:      "test",
			Arguments: []domain.PromptArgument{},
		}
		data, err := MarshalPromptDefinition(prompt)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		var mcpPrompt mcp.Prompt
		require.NoError(t, json.Unmarshal(data, &mcpPrompt))
		assert.Nil(t, mcpPrompt.Arguments)

		roundtripped := PromptFromMCP(&mcpPrompt)
		assert.Nil(t, roundtripped.Arguments)
	})

	t.Run("nil arguments handled", func(t *testing.T) {
		prompt := domain.PromptDefinition{
			Name:      "test",
			Arguments: nil,
		}
		data, err := MarshalPromptDefinition(prompt)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		var mcpPrompt mcp.Prompt
		require.NoError(t, json.Unmarshal(data, &mcpPrompt))
		assert.Nil(t, mcpPrompt.Arguments)

		roundtripped := PromptFromMCP(&mcpPrompt)
		assert.Nil(t, roundtripped.Arguments)
	})
}
