package gateway

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"

	"mcpv/internal/infra/rpc"
)

type runtime interface {
	StartRuntime(ctx context.Context) error
	StopRuntime(ctx context.Context) error
	Server() *mcp.Server
}

type runtimeFactory func(sel Selector, caller string) runtime

type PoolOptions struct {
	IdleTimeout    time.Duration
	MaxInstances   int
	RuntimeFactory runtimeFactory
}

type gatewayPool struct {
	ctx        context.Context
	baseCaller string
	logger     *zap.Logger
	options    PoolOptions

	mu       sync.Mutex
	runtimes map[string]*pooledRuntime
	closed   bool
	stopCh   chan struct{}
}

type pooledRuntime struct {
	runtime  runtime
	server   *mcp.Server
	lastUsed time.Time
}

func newGatewayPool(ctx context.Context, cfg rpc.ClientConfig, baseCaller string, logger *zap.Logger, opts PoolOptions) *gatewayPool {
	if logger == nil {
		logger = zap.NewNop()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if opts.RuntimeFactory == nil {
		opts.RuntimeFactory = func(sel Selector, caller string) runtime {
			return NewGateway(cfg, caller, sel.Tags, sel.Server, logger)
		}
	}
	pool := &gatewayPool{
		ctx:        ctx,
		baseCaller: strings.TrimSpace(baseCaller),
		logger:     logger.Named("gateway_pool"),
		options:    opts,
		runtimes:   make(map[string]*pooledRuntime),
		stopCh:     make(chan struct{}),
	}
	if pool.options.IdleTimeout > 0 {
		pool.startSweeper()
	}
	return pool
}

func (p *gatewayPool) Get(ctx context.Context, sel Selector) (*mcp.Server, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	sel = sel.normalized()
	if sel.empty() || (sel.Server != "" && len(sel.Tags) > 0) {
		return nil, errors.New("invalid selector")
	}
	key := SelectorKey(sel)
	if key == "" {
		return nil, errors.New("invalid selector")
	}
	now := time.Now()

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, errors.New("gateway pool closed")
	}
	if runtime, ok := p.runtimes[key]; ok {
		runtime.lastUsed = now
		server := runtime.server
		p.mu.Unlock()
		if server == nil {
			return nil, errors.New("gateway server unavailable")
		}
		return server, nil
	}
	if p.options.MaxInstances > 0 && len(p.runtimes) >= p.options.MaxInstances {
		p.mu.Unlock()
		return nil, errors.New("gateway pool capacity reached")
	}
	p.mu.Unlock()

	caller := deriveSelectorCaller(p.baseCaller, key)
	runtime := p.options.RuntimeFactory(sel, caller)
	if runtime == nil {
		return nil, errors.New("gateway runtime factory returned nil")
	}
	if err := runtime.StartRuntime(p.ctx); err != nil {
		return nil, err
	}
	server := runtime.Server()
	if server == nil {
		_ = runtime.StopRuntime(ctx)
		return nil, errors.New("gateway server unavailable")
	}

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		_ = runtime.StopRuntime(ctx)
		return nil, errors.New("gateway pool closed")
	}
	if existing, ok := p.runtimes[key]; ok {
		existing.lastUsed = now
		p.mu.Unlock()
		_ = runtime.StopRuntime(ctx)
		if existing.server == nil {
			return nil, errors.New("gateway server unavailable")
		}
		return existing.server, nil
	}
	p.runtimes[key] = &pooledRuntime{
		runtime:  runtime,
		server:   server,
		lastUsed: now,
	}
	p.mu.Unlock()

	return server, nil
}

func (p *gatewayPool) Close(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	close(p.stopCh)
	all := make([]*pooledRuntime, 0, len(p.runtimes))
	for _, runtime := range p.runtimes {
		all = append(all, runtime)
	}
	p.runtimes = make(map[string]*pooledRuntime)
	p.mu.Unlock()

	var result error
	for _, runtime := range all {
		if err := runtime.runtime.StopRuntime(ctx); err != nil && result == nil {
			result = err
		}
	}
	return result
}

func (p *gatewayPool) startSweeper() {
	interval := p.options.IdleTimeout / 2
	if interval < time.Second {
		interval = time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-p.stopCh:
				return
			case <-ticker.C:
				p.evictIdle(time.Now())
			}
		}
	}()
}

func (p *gatewayPool) evictIdle(now time.Time) {
	if p.options.IdleTimeout <= 0 {
		return
	}
	var expired []*pooledRuntime
	p.mu.Lock()
	for key, runtime := range p.runtimes {
		if now.Sub(runtime.lastUsed) > p.options.IdleTimeout {
			delete(p.runtimes, key)
			expired = append(expired, runtime)
		}
	}
	p.mu.Unlock()

	for _, runtime := range expired {
		_ = runtime.runtime.StopRuntime(context.Background())
	}
}

func deriveSelectorCaller(baseCaller, key string) string {
	base := strings.TrimSpace(baseCaller)
	if base == "" {
		base = "mcpvmcp"
	}
	return base + ":" + key
}
