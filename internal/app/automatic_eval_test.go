package app

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"mcpd/internal/domain"
)

func TestValidateToolArguments_Invalid(t *testing.T) {
	tool := domain.ToolDefinition{
		Name: "demo.echo",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []any{"message"},
			"properties": map[string]any{
				"message": map[string]any{"type": "string"},
			},
		},
	}
	args := json.RawMessage(`{}`)

	err := validateToolArguments(tool, args)
	require.Error(t, err)
}

func TestBuildAutomaticEvalSchemaError_EmbedsSchema(t *testing.T) {
	tool := domain.ToolDefinition{
		Name: "demo.echo",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []any{"message"},
			"properties": map[string]any{
				"message": map[string]any{"type": "string"},
			},
		},
	}

	raw, err := buildAutomaticEvalSchemaError(tool, errors.New("invalid tool arguments"))
	require.NoError(t, err)

	var result struct {
		IsError bool `json:"isError"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	require.NoError(t, json.Unmarshal(raw, &result))
	require.True(t, result.IsError)
	require.Len(t, result.Content, 1)

	var payload struct {
		Error      string                 `json:"error"`
		ToolSchema map[string]interface{} `json:"toolSchema"`
	}
	require.NoError(t, json.Unmarshal([]byte(result.Content[0].Text), &payload))
	require.Equal(t, "invalid tool arguments", payload.Error)
	require.Equal(t, "demo.echo", payload.ToolSchema["name"])
}
