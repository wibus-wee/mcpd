package domain

import (
	"context"
	"encoding/json"
)

type Router interface {
	Route(ctx context.Context, serverType, specKey, routingKey string, payload json.RawMessage) (json.RawMessage, error)
	RouteWithOptions(ctx context.Context, serverType, specKey, routingKey string, payload json.RawMessage, opts RouteOptions) (json.RawMessage, error)
}

type RouteOptions struct {
	AllowStart bool
}
