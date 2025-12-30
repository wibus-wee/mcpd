package app

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateToolArguments_Invalid(t *testing.T) {
	toolJSON := json.RawMessage(`{"name":"demo.echo","inputSchema":{"type":"object","required":["message"],"properties":{"message":{"type":"string"}}}}`)
	args := json.RawMessage(`{}`)

	err := validateToolArguments(toolJSON, args)
	require.Error(t, err)
}

func TestBuildAutomaticEvalSchemaError_EmbedsSchema(t *testing.T) {
	toolJSON := json.RawMessage(`{"name":"demo.echo","inputSchema":{"type":"object","required":["message"],"properties":{"message":{"type":"string"}}}}`)

	raw, err := buildAutomaticEvalSchemaError(toolJSON, errors.New("invalid tool arguments"))
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
