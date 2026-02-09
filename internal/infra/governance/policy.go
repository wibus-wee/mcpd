package governance

import (
	"context"
	"encoding/json"

	"mcpv/internal/domain"
	"mcpv/internal/infra/pipeline"
)

type Policy interface {
	Request(ctx context.Context, req domain.GovernanceRequest) (domain.GovernanceDecision, error)
	Response(ctx context.Context, req domain.GovernanceRequest) (domain.GovernanceDecision, error)
}

type Chain struct {
	policies []Policy
}

func NewChain(policies ...Policy) *Chain {
	filtered := make([]Policy, 0, len(policies))
	for _, policy := range policies {
		if policy == nil {
			continue
		}
		filtered = append(filtered, policy)
	}
	return &Chain{policies: filtered}
}

func (c *Chain) Request(ctx context.Context, req domain.GovernanceRequest) (domain.GovernanceDecision, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c == nil || len(c.policies) == 0 {
		return domain.GovernanceDecision{Continue: true}, nil
	}
	decision := domain.GovernanceDecision{Continue: true}
	working := req
	for _, policy := range c.policies {
		nextDecision, err := policy.Request(ctx, working)
		if err != nil {
			return domain.GovernanceDecision{}, err
		}
		decision = nextDecision
		if !nextDecision.Continue {
			return nextDecision, nil
		}
		if len(nextDecision.RequestJSON) > 0 {
			working.RequestJSON = nextDecision.RequestJSON
		}
	}
	return decision, nil
}

func (c *Chain) Response(ctx context.Context, req domain.GovernanceRequest) (domain.GovernanceDecision, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c == nil || len(c.policies) == 0 {
		return domain.GovernanceDecision{Continue: true}, nil
	}
	decision := domain.GovernanceDecision{Continue: true}
	working := req
	for i := len(c.policies) - 1; i >= 0; i-- {
		nextDecision, err := c.policies[i].Response(ctx, working)
		if err != nil {
			return domain.GovernanceDecision{}, err
		}
		decision = nextDecision
		if !nextDecision.Continue {
			return nextDecision, nil
		}
		if len(nextDecision.ResponseJSON) > 0 {
			working.ResponseJSON = nextDecision.ResponseJSON
		}
	}
	return decision, nil
}

func (c *Chain) Execute(ctx context.Context, req domain.GovernanceRequest, next func(context.Context, domain.GovernanceRequest) (json.RawMessage, error)) (json.RawMessage, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c == nil || len(c.policies) == 0 {
		return next(ctx, req)
	}

	working := req
	for _, policy := range c.policies {
		decision, err := policy.Request(ctx, working)
		if err != nil {
			return nil, err
		}
		if !decision.Continue {
			return handleRejection(req, decision)
		}
		if len(decision.RequestJSON) > 0 {
			working.RequestJSON = decision.RequestJSON
		}
	}

	resp, err := next(ctx, working)
	if err != nil {
		return nil, err
	}

	working.ResponseJSON = resp
	for i := len(c.policies) - 1; i >= 0; i-- {
		decision, err := c.policies[i].Response(ctx, working)
		if err != nil {
			return nil, err
		}
		if !decision.Continue {
			return handleRejection(req, decision)
		}
		if len(decision.ResponseJSON) > 0 {
			resp = decision.ResponseJSON
			working.ResponseJSON = decision.ResponseJSON
		}
	}

	return resp, nil
}

type PipelinePolicy struct {
	pipeline *pipeline.Engine
}

func NewPipelinePolicy(pipe *pipeline.Engine) *PipelinePolicy {
	return &PipelinePolicy{pipeline: pipe}
}

func (p *PipelinePolicy) Request(ctx context.Context, req domain.GovernanceRequest) (domain.GovernanceDecision, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if p == nil || p.pipeline == nil {
		return domain.GovernanceDecision{Continue: true}, nil
	}
	req.Flow = domain.PluginFlowRequest
	return p.pipeline.Handle(ctx, req)
}

func (p *PipelinePolicy) Response(ctx context.Context, req domain.GovernanceRequest) (domain.GovernanceDecision, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if p == nil || p.pipeline == nil {
		return domain.GovernanceDecision{Continue: true}, nil
	}
	req.Flow = domain.PluginFlowResponse
	return p.pipeline.Handle(ctx, req)
}
