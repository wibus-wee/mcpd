package governance

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcpv/internal/domain"
	"mcpv/internal/infra/pipeline"
)

type Executor struct {
	chain *Chain
}

func NewExecutor(pipe *pipeline.Engine) *Executor {
	if pipe == nil {
		return &Executor{}
	}
	return &Executor{chain: NewChain(NewPipelinePolicy(pipe))}
}

func NewExecutorWithPolicies(policies ...Policy) *Executor {
	return &Executor{chain: NewChain(policies...)}
}

func (e *Executor) Request(ctx context.Context, req domain.GovernanceRequest) (domain.GovernanceDecision, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if e.chain == nil {
		return domain.GovernanceDecision{Continue: true}, nil
	}
	return e.chain.Request(ctx, req)
}

func (e *Executor) Response(ctx context.Context, req domain.GovernanceRequest) (domain.GovernanceDecision, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if e.chain == nil {
		return domain.GovernanceDecision{Continue: true}, nil
	}
	return e.chain.Response(ctx, req)
}

func (e *Executor) Execute(ctx context.Context, req domain.GovernanceRequest, next func(context.Context, domain.GovernanceRequest) (json.RawMessage, error)) (json.RawMessage, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if e.chain == nil {
		return next(ctx, req)
	}
	return e.chain.Execute(ctx, req, next)
}

func handleRejection(req domain.GovernanceRequest, decision domain.GovernanceDecision) (json.RawMessage, error) {
	if req.Method == "tools/call" {
		return buildToolRejection(decision)
	}
	return nil, domain.GovernanceRejection{
		Category: decision.Category,
		Plugin:   decision.Plugin,
		Code:     decision.RejectCode,
		Message:  decision.RejectMessage,
	}
}

func buildToolRejection(decision domain.GovernanceDecision) (json.RawMessage, error) {
	message := decision.RejectMessage
	if message == "" {
		message = "request rejected"
	}

	structured := map[string]any{
		"code":    decision.RejectCode,
		"message": message,
	}
	result := mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: message},
		},
		StructuredContent: structured,
	}
	payload, err := json.Marshal(&result)
	if err != nil {
		return nil, fmt.Errorf("encode rejection: %w", err)
	}
	return payload, nil
}
