package governance

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mcpv/internal/domain"
)

// TestPipelinePolicy_NilHandling verifies nil pipeline handling.
func TestPipelinePolicy_NilHandling(t *testing.T) {
	t.Run("nil policy returns Continue", func(t *testing.T) {
		var policy *PipelinePolicy
		req := domain.GovernanceRequest{Method: "test"}

		decision, err := policy.Request(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, decision.Continue)

		decision, err = policy.Response(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, decision.Continue)
	})

	t.Run("policy with nil pipeline returns Continue", func(t *testing.T) {
		policy := &PipelinePolicy{}
		req := domain.GovernanceRequest{Method: "test"}

		decision, err := policy.Request(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, decision.Continue)

		decision, err = policy.Response(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, decision.Continue)
	})

	t.Run("nil context uses Background", func(t *testing.T) {
		policy := &PipelinePolicy{}
		req := domain.GovernanceRequest{Method: "test"}

		_, err := policy.Request(nil, req)
		require.NoError(t, err)

		_, err = policy.Response(context.Background(), req)
		require.NoError(t, err)
	})
}
