package governance

import (
	"context"

	"mcpv/internal/domain"
)

// mockPolicy implements Policy for testing.
type mockPolicy struct {
	name         string
	requestFunc  func(ctx context.Context, req domain.GovernanceRequest) (domain.GovernanceDecision, error)
	responseFunc func(ctx context.Context, req domain.GovernanceRequest) (domain.GovernanceDecision, error)
}

func (m *mockPolicy) Request(ctx context.Context, req domain.GovernanceRequest) (domain.GovernanceDecision, error) {
	if m.requestFunc != nil {
		return m.requestFunc(ctx, req)
	}
	return domain.GovernanceDecision{Continue: true}, nil
}

func (m *mockPolicy) Response(ctx context.Context, req domain.GovernanceRequest) (domain.GovernanceDecision, error) {
	if m.responseFunc != nil {
		return m.responseFunc(ctx, req)
	}
	return domain.GovernanceDecision{Continue: true}, nil
}
