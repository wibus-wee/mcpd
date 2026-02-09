package governance

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mcpv/internal/domain"
)

// TestChain_ExecutionOrder verifies chain execution order.
func TestChain_ExecutionOrder(t *testing.T) {
	var executionOrder []string
	var mu sync.Mutex

	policy1 := &mockPolicy{
		name: "policy1",
		requestFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			mu.Lock()
			executionOrder = append(executionOrder, "p1-request")
			mu.Unlock()
			return domain.GovernanceDecision{Continue: true}, nil
		},
		responseFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			mu.Lock()
			executionOrder = append(executionOrder, "p1-response")
			mu.Unlock()
			return domain.GovernanceDecision{Continue: true}, nil
		},
	}

	policy2 := &mockPolicy{
		name: "policy2",
		requestFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			mu.Lock()
			executionOrder = append(executionOrder, "p2-request")
			mu.Unlock()
			return domain.GovernanceDecision{Continue: true}, nil
		},
		responseFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			mu.Lock()
			executionOrder = append(executionOrder, "p2-response")
			mu.Unlock()
			return domain.GovernanceDecision{Continue: true}, nil
		},
	}

	policy3 := &mockPolicy{
		name: "policy3",
		requestFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			mu.Lock()
			executionOrder = append(executionOrder, "p3-request")
			mu.Unlock()
			return domain.GovernanceDecision{Continue: true}, nil
		},
		responseFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			mu.Lock()
			executionOrder = append(executionOrder, "p3-response")
			mu.Unlock()
			return domain.GovernanceDecision{Continue: true}, nil
		},
	}

	chain := NewChain(policy1, policy2, policy3)
	req := domain.GovernanceRequest{Method: "test"}

	t.Run("request policies execute forward", func(t *testing.T) {
		executionOrder = nil
		_, err := chain.Request(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, []string{"p1-request", "p2-request", "p3-request"}, executionOrder)
	})

	t.Run("response policies execute backward", func(t *testing.T) {
		executionOrder = nil
		_, err := chain.Response(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, []string{"p3-response", "p2-response", "p1-response"}, executionOrder)
	})

	t.Run("execute runs request forward then response backward", func(t *testing.T) {
		executionOrder = nil
		nextCalled := false
		next := func(_ context.Context, _ domain.GovernanceRequest) (json.RawMessage, error) {
			mu.Lock()
			executionOrder = append(executionOrder, "next")
			nextCalled = true
			mu.Unlock()
			return json.RawMessage(`{"result":"ok"}`), nil
		}

		_, err := chain.Execute(context.Background(), req, next)
		require.NoError(t, err)
		assert.True(t, nextCalled)
		assert.Equal(t, []string{
			"p1-request", "p2-request", "p3-request",
			"next",
			"p3-response", "p2-response", "p1-response",
		}, executionOrder)
	})
}

// TestChain_EarlyRejection verifies early rejection stops chain.
func TestChain_EarlyRejection(t *testing.T) {
	tests := []struct {
		name           string
		rejectAt       int
		totalPolicies  int
		calledPolicies []string
	}{
		{
			name:           "first policy rejects, others not called",
			rejectAt:       0,
			totalPolicies:  3,
			calledPolicies: []string{"p0"},
		},
		{
			name:           "middle policy rejects, later not called",
			rejectAt:       1,
			totalPolicies:  3,
			calledPolicies: []string{"p0", "p1"},
		},
		{
			name:           "last policy rejects",
			rejectAt:       2,
			totalPolicies:  3,
			calledPolicies: []string{"p0", "p1", "p2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calledPolicies []string
			var mu sync.Mutex

			policies := make([]Policy, tt.totalPolicies)
			for i := 0; i < tt.totalPolicies; i++ {
				idx := i
				policies[i] = &mockPolicy{
					name: "p" + string(rune('0'+i)),
					requestFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
						mu.Lock()
						calledPolicies = append(calledPolicies, "p"+string(rune('0'+idx)))
						mu.Unlock()

						if idx == tt.rejectAt {
							return domain.GovernanceDecision{
								Continue:      false,
								RejectCode:    "REJECTED",
								RejectMessage: "Policy rejected",
							}, nil
						}
						return domain.GovernanceDecision{Continue: true}, nil
					},
				}
			}

			chain := NewChain(policies...)
			req := domain.GovernanceRequest{Method: "test"}

			decision, err := chain.Request(context.Background(), req)
			require.NoError(t, err)
			assert.False(t, decision.Continue)
			assert.Equal(t, "REJECTED", decision.RejectCode)
			assert.Equal(t, tt.calledPolicies, calledPolicies)
		})
	}
}

// TestChain_RequestMutation verifies request mutation propagation.
func TestChain_RequestMutation(t *testing.T) {
	policy1 := &mockPolicy{
		requestFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			return domain.GovernanceDecision{
				Continue:    true,
				RequestJSON: json.RawMessage(`{"step":1}`),
			}, nil
		},
	}

	policy2 := &mockPolicy{
		requestFunc: func(_ context.Context, req domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			assert.Equal(t, json.RawMessage(`{"step":1}`), req.RequestJSON)
			return domain.GovernanceDecision{
				Continue:    true,
				RequestJSON: json.RawMessage(`{"step":2}`),
			}, nil
		},
	}

	policy3 := &mockPolicy{
		requestFunc: func(_ context.Context, req domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			assert.Equal(t, json.RawMessage(`{"step":2}`), req.RequestJSON)
			return domain.GovernanceDecision{
				Continue:    true,
				RequestJSON: json.RawMessage(`{"step":3}`),
			}, nil
		},
	}

	chain := NewChain(policy1, policy2, policy3)
	req := domain.GovernanceRequest{
		Method:      "test",
		RequestJSON: json.RawMessage(`{"step":0}`),
	}

	decision, err := chain.Request(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, decision.Continue)
	assert.Equal(t, json.RawMessage(`{"step":3}`), decision.RequestJSON)
}

// TestChain_ResponseMutation verifies response mutation propagation.
func TestChain_ResponseMutation(t *testing.T) {
	policy1 := &mockPolicy{
		responseFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			return domain.GovernanceDecision{
				Continue:     true,
				ResponseJSON: json.RawMessage(`{"step":1}`),
			}, nil
		},
	}

	policy2 := &mockPolicy{
		responseFunc: func(_ context.Context, req domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			assert.Equal(t, json.RawMessage(`{"step":1}`), req.ResponseJSON)
			return domain.GovernanceDecision{
				Continue:     true,
				ResponseJSON: json.RawMessage(`{"step":2}`),
			}, nil
		},
	}

	chain := NewChain(policy2, policy1)
	req := domain.GovernanceRequest{
		Method:       "test",
		ResponseJSON: json.RawMessage(`{"step":0}`),
	}

	decision, err := chain.Response(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, decision.Continue)
	assert.Equal(t, json.RawMessage(`{"step":2}`), decision.ResponseJSON)
}

// TestChain_NilHandling verifies nil policy/chain handling.
func TestChain_NilHandling(t *testing.T) {
	t.Run("nil chain returns Continue", func(t *testing.T) {
		var chain *Chain
		req := domain.GovernanceRequest{Method: "test"}

		decision, err := chain.Request(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, decision.Continue)

		decision, err = chain.Response(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, decision.Continue)
	})

	t.Run("empty chain returns Continue", func(t *testing.T) {
		chain := NewChain()
		req := domain.GovernanceRequest{Method: "test"}

		decision, err := chain.Request(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, decision.Continue)

		decision, err = chain.Response(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, decision.Continue)
	})

	t.Run("chain with nil policies filters them out", func(t *testing.T) {
		policy := &mockPolicy{
			requestFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
				return domain.GovernanceDecision{
					Continue:    true,
					RequestJSON: json.RawMessage(`{"called":true}`),
				}, nil
			},
		}

		chain := NewChain(nil, policy, nil)
		req := domain.GovernanceRequest{Method: "test"}

		decision, err := chain.Request(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, decision.Continue)
		assert.Equal(t, json.RawMessage(`{"called":true}`), decision.RequestJSON)
	})

	t.Run("nil context uses Background", func(t *testing.T) {
		policy := &mockPolicy{
			requestFunc: func(ctx context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
				assert.NotNil(t, ctx)
				return domain.GovernanceDecision{Continue: true}, nil
			},
		}

		chain := NewChain(policy)
		req := domain.GovernanceRequest{Method: "test"}

		_, err := chain.Request(context.Background(), req)
		require.NoError(t, err)
	})
}

// TestChain_ErrorHandling verifies error propagation.
func TestChain_ErrorHandling(t *testing.T) {
	testErr := errors.New("policy error")

	t.Run("request error stops chain", func(t *testing.T) {
		policy1 := &mockPolicy{
			requestFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
				return domain.GovernanceDecision{}, testErr
			},
		}

		policy2 := &mockPolicy{
			requestFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
				t.Fatal("policy2 should not be called")
				return domain.GovernanceDecision{Continue: true}, nil
			},
		}

		chain := NewChain(policy1, policy2)
		req := domain.GovernanceRequest{Method: "test"}

		_, err := chain.Request(context.Background(), req)
		assert.ErrorIs(t, err, testErr)
	})

	t.Run("response error stops chain", func(t *testing.T) {
		policy1 := &mockPolicy{
			responseFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
				return domain.GovernanceDecision{}, testErr
			},
		}

		policy2 := &mockPolicy{
			responseFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
				t.Fatal("policy2 should not be called")
				return domain.GovernanceDecision{Continue: true}, nil
			},
		}

		chain := NewChain(policy2, policy1)
		req := domain.GovernanceRequest{Method: "test"}

		_, err := chain.Response(context.Background(), req)
		assert.ErrorIs(t, err, testErr)
	})
}

// TestChain_ConcurrentExecute verifies thread-safe execution.
func TestChain_ConcurrentExecute(t *testing.T) {
	const goroutines = 100

	policy := &mockPolicy{
		requestFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			return domain.GovernanceDecision{Continue: true}, nil
		},
		responseFunc: func(_ context.Context, _ domain.GovernanceRequest) (domain.GovernanceDecision, error) {
			return domain.GovernanceDecision{Continue: true}, nil
		},
	}

	chain := NewChain(policy)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	errors := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			req := domain.GovernanceRequest{Method: "test"}
			next := func(_ context.Context, _ domain.GovernanceRequest) (json.RawMessage, error) {
				return json.RawMessage(`{"ok":true}`), nil
			}
			_, errors[idx] = chain.Execute(context.Background(), req, next)
		}(i)
	}

	wg.Wait()

	for i, err := range errors {
		assert.NoError(t, err, "Execution %d failed", i)
	}
}
