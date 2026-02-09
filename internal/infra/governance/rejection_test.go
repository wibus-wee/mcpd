package governance

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mcpv/internal/domain"
)

// TestHandleRejection_ToolCall verifies tool call rejection formatting.
func TestHandleRejection_ToolCall(t *testing.T) {
	req := domain.GovernanceRequest{
		Method: "tools/call",
	}

	decision := domain.GovernanceDecision{
		Continue:      false,
		RejectCode:    "FORBIDDEN",
		RejectMessage: "Tool access denied",
	}

	result, err := handleRejection(req, decision)
	require.NoError(t, err)
	assert.NotEmpty(t, result)

	var parsed map[string]any
	err = json.Unmarshal(result, &parsed)
	require.NoError(t, err)
	assert.True(t, parsed["isError"].(bool))
	assert.NotNil(t, parsed["content"])
	assert.NotNil(t, parsed["structuredContent"])
}

// TestHandleRejection_OtherMethods verifies non-tool rejection formatting.
func TestHandleRejection_OtherMethods(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{"resources/read", "resources/read"},
		{"prompts/get", "prompts/get"},
		{"custom/method", "custom/method"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := domain.GovernanceRequest{
				Method: tt.method,
			}

			decision := domain.GovernanceDecision{
				Continue:      false,
				RejectCode:    "FORBIDDEN",
				RejectMessage: "Access denied",
			}

			result, err := handleRejection(req, decision)
			assert.Nil(t, result)
			assert.Error(t, err)

			var govErr domain.GovernanceRejection
			assert.ErrorAs(t, err, &govErr)
			assert.Equal(t, "FORBIDDEN", govErr.Code)
			assert.Equal(t, "Access denied", govErr.Message)
		})
	}
}

// TestBuildToolRejection_EmptyMessage verifies default message.
func TestBuildToolRejection_EmptyMessage(t *testing.T) {
	decision := domain.GovernanceDecision{
		Continue:      false,
		RejectCode:    "ERROR",
		RejectMessage: "",
	}

	result, err := buildToolRejection(decision)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(result, &parsed)
	require.NoError(t, err)

	content := parsed["content"].([]any)[0].(map[string]any)
	assert.Equal(t, "request rejected", content["text"])
}
