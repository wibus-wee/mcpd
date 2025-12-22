package router

import (
	"context"
	"encoding/json"
	"fmt"

	"mcpd/internal/domain"
)

type BasicRouter struct {
	scheduler domain.Scheduler
}

func NewBasicRouter(scheduler domain.Scheduler) *BasicRouter {
	return &BasicRouter{scheduler: scheduler}
}

func (r *BasicRouter) Route(ctx context.Context, serverType, routingKey string, payload json.RawMessage) (json.RawMessage, error) {
	inst, err := r.scheduler.Acquire(ctx, serverType, routingKey)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.scheduler.Release(ctx, inst) }()

	if inst.Conn == nil {
		return nil, fmt.Errorf("instance has no connection: %s", inst.ID)
	}

	if err := inst.Conn.Send(ctx, payload); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	resp, err := inst.Conn.Recv(ctx)
	if err != nil {
		return nil, fmt.Errorf("receive response: %w", err)
	}

	return resp, nil
}
