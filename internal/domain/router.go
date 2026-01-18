package domain

import (
	"context"
	"encoding/json"
)

// Router routes requests to server instances.
type Router interface {
	Route(ctx context.Context, serverType, specKey, routingKey string, payload json.RawMessage) (json.RawMessage, error)
	RouteWithOptions(ctx context.Context, serverType, specKey, routingKey string, payload json.RawMessage, opts RouteOptions) (json.RawMessage, error)
}

// RouteOptions customizes routing behavior.
type RouteOptions struct {
	AllowStart bool
}
