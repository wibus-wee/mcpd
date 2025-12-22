package router

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"go.uber.org/zap"

	"mcpd/internal/domain"
)

type BasicRouter struct {
	scheduler domain.Scheduler
	timeout   time.Duration
	logger    *zap.Logger
	metrics   domain.Metrics
}

type RouterOptions struct {
	Timeout time.Duration
	Logger  *zap.Logger
	Metrics domain.Metrics
}

func NewBasicRouter(scheduler domain.Scheduler, opts RouterOptions) *BasicRouter {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = time.Duration(domain.DefaultRouteTimeoutSeconds) * time.Second
	}
	logger := opts.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	return &BasicRouter{
		scheduler: scheduler,
		timeout:   timeout,
		logger:    logger.Named("router"),
		metrics:   opts.Metrics,
	}
}

func (r *BasicRouter) Route(ctx context.Context, serverType, routingKey string, payload json.RawMessage) (json.RawMessage, error) {
	method, isCall, err := extractMethod(payload)
	if err != nil {
		return nil, fmt.Errorf("decode request: %w", err)
	}
	if method == "" || !isCall {
		return nil, domain.ErrInvalidRequest
	}

	start := time.Now()

	inst, err := r.scheduler.Acquire(ctx, serverType, routingKey)
	if err != nil {
		r.observeRoute(serverType, start, err)
		return nil, err
	}
	defer func() { _ = r.scheduler.Release(ctx, inst) }()

	if inst.Conn == nil {
		err := fmt.Errorf("instance has no connection: %s", inst.ID)
		r.observeRoute(serverType, start, err)
		return nil, err
	}

	if !domain.MethodAllowed(inst.Capabilities, method) {
		r.logger.Warn("method not allowed", zap.String("serverType", serverType), zap.String("method", method))
		r.observeRoute(serverType, start, domain.ErrMethodNotAllowed)
		return nil, domain.ErrMethodNotAllowed
	}

	callCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	if err := inst.Conn.Send(callCtx, payload); err != nil {
		sendErr := fmt.Errorf("send request: %w", err)
		r.observeRoute(serverType, start, sendErr)
		return nil, sendErr
	}

	resp, err := inst.Conn.Recv(callCtx)
	if err != nil {
		recvErr := fmt.Errorf("receive response: %w", err)
		r.observeRoute(serverType, start, recvErr)
		return nil, recvErr
	}

	r.observeRoute(serverType, start, nil)
	return resp, nil
}

func (r *BasicRouter) observeRoute(serverType string, start time.Time, err error) {
	if r.metrics == nil {
		return
	}
	r.metrics.ObserveRoute(serverType, time.Since(start), err)
}

func extractMethod(payload json.RawMessage) (string, bool, error) {
	msg, err := jsonrpc.DecodeMessage(payload)
	if err != nil {
		return "", false, err
	}
	if req, ok := msg.(*jsonrpc.Request); ok {
		return req.Method, req.ID.IsValid(), nil
	}
	return "", false, nil
}
