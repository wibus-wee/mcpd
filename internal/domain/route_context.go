package domain

import (
	"context"
	"time"
)

type RouteContext struct {
	Caller  string
	Profile string
}

type routeContextKey struct{}

type StartCauseReason string

const (
	StartCauseBootstrap      StartCauseReason = "bootstrap"
	StartCauseToolCall       StartCauseReason = "tool_call"
	StartCauseCallerActivate StartCauseReason = "caller_activate"
	StartCausePolicyAlwaysOn StartCauseReason = "policy_always_on"
	StartCausePolicyMinReady StartCauseReason = "policy_min_ready"
)

type StartCausePolicy struct {
	ActivationMode ActivationMode `json:"activationMode"`
	MinReady       int            `json:"minReady"`
}

type StartCause struct {
	Reason    StartCauseReason  `json:"reason"`
	Caller    string            `json:"caller,omitempty"`
	ToolName  string            `json:"toolName,omitempty"`
	Profile   string            `json:"profile,omitempty"`
	Policy    *StartCausePolicy `json:"policy,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

type startCauseKey struct{}

func WithRouteContext(ctx context.Context, meta RouteContext) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, routeContextKey{}, meta)
}

func RouteContextFrom(ctx context.Context) (RouteContext, bool) {
	if ctx == nil {
		return RouteContext{}, false
	}
	meta, ok := ctx.Value(routeContextKey{}).(RouteContext)
	return meta, ok
}

func WithStartCause(ctx context.Context, cause StartCause) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, startCauseKey{}, cause)
}

func StartCauseFromContext(ctx context.Context) (StartCause, bool) {
	if ctx == nil {
		return StartCause{}, false
	}
	cause, ok := ctx.Value(startCauseKey{}).(StartCause)
	return cause, ok
}

func CloneStartCause(cause *StartCause) *StartCause {
	if cause == nil {
		return nil
	}
	copyCause := *cause
	if cause.Policy != nil {
		policyCopy := *cause.Policy
		copyCause.Policy = &policyCopy
	}
	return &copyCause
}
