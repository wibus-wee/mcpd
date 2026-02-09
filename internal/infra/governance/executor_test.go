package governance

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mcpv/internal/domain"
)

// TestExecutor_NilChain verifies executor with nil chain.
func TestExecutor_NilChain(t *testing.T) {
	executor := &Executor{}
	req := domain.GovernanceRequest{Method: "test"}

	t.Run("request returns Continue", func(t *testing.T) {
		decision, err := executor.Request(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, decision.Continue)
	})

	t.Run("response returns Continue", func(t *testing.T) {
		decision, err := executor.Response(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, decision.Continue)
	})

	t.Run("execute calls next", func(t *testing.T) {
		nextCalled := false
		next := func(_ context.Context, _ domain.GovernanceRequest) (json.RawMessage, error) {
			nextCalled = true
			return json.RawMessage(`{"ok":true}`), nil
		}

		result, err := executor.Execute(context.Background(), req, next)
		require.NoError(t, err)
		assert.True(t, nextCalled)
		assert.Equal(t, json.RawMessage(`{"ok":true}`), result)
	})
}

// TestNewExecutor verifies executor construction.
func TestNewExecutor(t *testing.T) {
	t.Run("nil pipeline creates empty executor", func(t *testing.T) {
		executor := NewExecutor(nil)
		assert.NotNil(t, executor)
		assert.Nil(t, executor.chain)
	})

	t.Run("with policies creates executor with chain", func(t *testing.T) {
		policy := &mockPolicy{}
		executor := NewExecutorWithPolicies(policy)
		assert.NotNil(t, executor)
		assert.NotNil(t, executor.chain)
	})
}

// TestExecute_FullFlow verifies complete execute flow with mutations.
func TestExecute_FullFlow(t *testing.T) {
	policy1 := &mockPolicy{
		requestFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			return domain.GovernanceDecision{
				Continue:    true,
				RequestJSON: json.RawMessage(`{"modified":"request"}`),
			}, nil
		},
		responseFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			return domain.GovernanceDecision{
				Continue:     true,
				ResponseJSON: json.RawMessage(`{"modified":"response"}`),
			}, nil
		},
	}

	chain := NewChain(policy1)
	req := domain.GovernanceRequest{
		Method:      "test",
		RequestJSON: json.RawMessage(`{"original":"request"}`),
	}

	nextCalled := false
	var capturedRequest domain.GovernanceRequest
	next := func(_ context.Context, req domain.GovernanceRequest) (json.RawMessage, error) {
		nextCalled = true
		capturedRequest = req
		return json.RawMessage(`{"original":"response"}`), nil
	}

	result, err := chain.Execute(context.Background(), req, next)
	require.NoError(t, err)
	assert.True(t, nextCalled)
	assert.Equal(t, json.RawMessage(`{"modified":"request"}`), capturedRequest.RequestJSON)
	assert.Equal(t, json.RawMessage(`{"modified":"response"}`), result)
}

// TestExecute_RequestRejection verifies rejection during request phase.
func TestExecute_RequestRejection(t *testing.T) {
	policy := &mockPolicy{
		requestFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			return domain.GovernanceDecision{
				Continue:      false,
				RejectCode:    "BLOCKED",
				RejectMessage: "Request blocked",
			}, nil
		},
	}

	chain := NewChain(policy)
	req := domain.GovernanceRequest{Method: "resources/read"}

	next := func(_ context.Context, _ domain.GovernanceRequest) (json.RawMessage, error) {
		t.Fatal("next should not be called")
		return nil, nil
	}

	result, err := chain.Execute(context.Background(), req, next)
	assert.Nil(t, result)
	assert.Error(t, err)

	var govErr domain.GovernanceRejection
	assert.ErrorAs(t, err, &govErr)
	assert.Equal(t, "BLOCKED", govErr.Code)
}

// TestExecute_ResponseRejection verifies rejection during response phase.
func TestExecute_ResponseRejection(t *testing.T) {
	policy := &mockPolicy{
		requestFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			return domain.GovernanceDecision{Continue: true}, nil
		},
		responseFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			return domain.GovernanceDecision{
				Continue:      false,
				RejectCode:    "BLOCKED",
				RejectMessage: "Response blocked",
			}, nil
		},
	}

	chain := NewChain(policy)
	req := domain.GovernanceRequest{Method: "tools/call"}

	nextCalled := false
	next := func(_ context.Context, _ domain.GovernanceRequest) (json.RawMessage, error) {
		nextCalled = true
		return json.RawMessage(`{"result":"ok"}`), nil
	}

	result, err := chain.Execute(context.Background(), req, next)
	assert.True(t, nextCalled)
	assert.NotNil(t, result)
	assert.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(result, &parsed)
	require.NoError(t, err)
	assert.True(t, parsed["isError"].(bool))
}

// TestExecute_NextError verifies error from next function.
func TestExecute_NextError(t *testing.T) {
	policy := &mockPolicy{
		requestFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			return domain.GovernanceDecision{Continue: true}, nil
		},
	}

	chain := NewChain(policy)
	req := domain.GovernanceRequest{Method: "test"}

	testErr := errors.New("next error")
	next := func(_ context.Context, _ domain.GovernanceRequest) (json.RawMessage, error) {
		return nil, testErr
	}

	result, err := chain.Execute(context.Background(), req, next)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, testErr)
}

// TestExecute_NilChain verifies execute with nil chain.
func TestExecute_NilChain(t *testing.T) {
	var chain *Chain
	req := domain.GovernanceRequest{Method: "test"}

	nextCalled := false
	next := func(_ context.Context, _ domain.GovernanceRequest) (json.RawMessage, error) {
		nextCalled = true
		return json.RawMessage(`{"ok":true}`), nil
	}

	result, err := chain.Execute(context.Background(), req, next)
	require.NoError(t, err)
	assert.True(t, nextCalled)
	assert.Equal(t, json.RawMessage(`{"ok":true}`), result)
}

// TestExecutor_NilContext verifies nil context handling.
func TestExecutor_NilContext(t *testing.T) {
	policy := &mockPolicy{
		requestFunc: func(ctx context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			assert.NotNil(t, ctx)
			return domain.GovernanceDecision{Continue: true}, nil
		},
	}

	executor := NewExecutorWithPolicies(policy)
	req := domain.GovernanceRequest{Method: "test"}

	_, err := executor.Request(context.Background(), req)
	require.NoError(t, err)

	_, err = executor.Response(context.Background(), req)
	require.NoError(t, err)

	next := func(ctx context.Context, _ domain.GovernanceRequest) (json.RawMessage, error) {
		assert.NotNil(t, ctx)
		return json.RawMessage(`{"ok":true}`), nil
	}
	_, err = executor.Execute(context.Background(), req, next)
	require.NoError(t, err)
}
