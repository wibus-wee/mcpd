package gateway

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"mcpv/internal/buildinfo"
	"mcpv/internal/infra/rpc"
	"mcpv/internal/infra/subagent"
	controlv1 "mcpv/pkg/api/control/v1"
)

type Gateway struct {
	cfg               rpc.ClientConfig
	caller            string
	tags              []string
	serverName        string
	logger            *zap.Logger
	server            *mcp.Server
	clients           *clientManager
	registry          *toolRegistry
	resources         *resourceRegistry
	prompts           *promptRegistry
	callerPID         int64
	registered        atomic.Bool
	subAgentEnabled   atomic.Bool
	toolsReadyCh      chan struct{}
	toolsReadyOnce    sync.Once
	toolsReadyWarn    atomic.Bool
	toolsReadyWait    time.Duration
	serverReadyCh     chan struct{}
	registerReadyCh   chan struct{}
	registerReadyOnce sync.Once
	runtimeMu         sync.Mutex
	runtimeCancel     context.CancelFunc
	runtimeDone       chan error
	runtimeStarted    bool
}

const defaultHeartbeatInterval = 2 * time.Second
const defaultToolsReadyWait = 2 * time.Second

func NewGateway(cfg rpc.ClientConfig, caller string, tags []string, serverName string, logger *zap.Logger) *Gateway {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Gateway{
		cfg:            cfg,
		caller:         caller,
		tags:           tags,
		serverName:     serverName,
		logger:         logger.Named("gateway"),
		callerPID:      resolveCallerPID(),
		toolsReadyCh:   make(chan struct{}),
		toolsReadyWait: defaultToolsReadyWait,
	}
}

func (g *Gateway) Run(ctx context.Context) error {
	return g.run(ctx, func(runCtx context.Context) error {
		g.logger.Info("gateway starting (stdio transport)")
		return g.server.Run(runCtx, &mcp.StdioTransport{})
	})
}

func (g *Gateway) run(ctx context.Context, runner func(context.Context) error) error {
	if err := g.validateRuntimeConfig(); err != nil {
		return err
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	g.server = mcp.NewServer(&mcp.Implementation{
		Name:    "mcpv-mcp",
		Version: buildinfo.Version,
	}, &mcp.ServerOptions{
		HasTools:     true,
		HasResources: true,
		HasPrompts:   true,
	})
	if g.serverReadyCh != nil {
		close(g.serverReadyCh)
	}
	g.server.AddReceivingMiddleware(g.toolsReadyMiddleware())

	g.clients = newClientManager(g.cfg, g.logger)
	g.registry = newToolRegistry(g.server, g.toolHandler, g.logger)
	g.resources = newResourceRegistry(g.server, g.resourceHandler, g.logger)
	g.prompts = newPromptRegistry(g.server, g.promptHandler, g.logger)

	if err := g.registerCaller(runCtx); err != nil {
		return err
	}
	defer func() {
		cancel()
		_ = g.unregisterCaller(context.Background())
	}()

	if err := g.checkAndSetupSubAgent(runCtx); err != nil {
		g.logger.Warn("failed to check SubAgent status", zap.Error(err))
	}

	go g.heartbeat(runCtx)
	if !g.subAgentEnabled.Load() {
		go g.syncTools(runCtx)
	}
	go g.syncResources(runCtx)
	go g.syncPrompts(runCtx)
	go newLogBridge(g.server, g.clients, g.caller, g.tags, g.serverName, g.callerPID, g.logger).Run(runCtx)

	err := runner(runCtx)
	g.clients.close()
	return err
}

func (g *Gateway) StartRuntime(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	g.runtimeMu.Lock()
	if g.runtimeStarted {
		g.runtimeMu.Unlock()
		return nil
	}
	g.runtimeStarted = true
	g.serverReadyCh = make(chan struct{})
	g.registerReadyCh = make(chan struct{})
	runtimeCtx, cancel := context.WithCancel(ctx)
	g.runtimeCancel = cancel
	doneCh := make(chan error, 1)
	g.runtimeDone = doneCh
	g.runtimeMu.Unlock()

	go func() {
		doneCh <- g.run(runtimeCtx, func(runCtx context.Context) error {
			<-runCtx.Done()
			return runCtx.Err()
		})
	}()

	select {
	case <-g.registerReadyCh:
	case err := <-doneCh:
		g.runtimeMu.Lock()
		g.runtimeStarted = false
		g.runtimeCancel = nil
		g.runtimeDone = nil
		g.registerReadyCh = nil
		g.registerReadyOnce = sync.Once{}
		g.runtimeMu.Unlock()
		return err
	case <-ctx.Done():
		g.runtimeMu.Lock()
		g.runtimeStarted = false
		g.runtimeCancel = nil
		g.runtimeDone = nil
		g.registerReadyCh = nil
		g.registerReadyOnce = sync.Once{}
		g.runtimeMu.Unlock()
		return ctx.Err()
	}

	select {
	case err := <-doneCh:
		if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			return err
		}
	default:
	}

	return nil
}

func (g *Gateway) StopRuntime(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	g.runtimeMu.Lock()
	if !g.runtimeStarted {
		g.runtimeMu.Unlock()
		return nil
	}
	cancel := g.runtimeCancel
	done := g.runtimeDone
	g.runtimeCancel = nil
	g.runtimeDone = nil
	g.runtimeStarted = false
	g.registerReadyCh = nil
	g.registerReadyOnce = sync.Once{}
	g.runtimeMu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done == nil {
		return nil
	}
	select {
	case err := <-done:
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (g *Gateway) Server() *mcp.Server {
	return g.server
}

func (g *Gateway) validateRuntimeConfig() error {
	if g.cfg.Address == "" {
		return errors.New("rpc address is required")
	}
	if g.cfg.MaxRecvMsgSize <= 0 {
		return errors.New("rpc max recv message size must be > 0")
	}
	if g.cfg.MaxSendMsgSize <= 0 {
		return errors.New("rpc max send message size must be > 0")
	}
	if g.caller == "" {
		return errors.New("caller is required")
	}
	return nil
}

func (g *Gateway) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(defaultHeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := g.registerCaller(ctx); err != nil {
				g.logger.Warn("caller heartbeat failed", zap.Error(err))
			}
		}
	}
}

func (g *Gateway) registerCaller(ctx context.Context) error {
	client, err := g.clients.get(ctx)
	if err != nil {
		return err
	}
	resp, err := client.Control().RegisterCaller(ctx, &controlv1.RegisterCallerRequest{
		Caller: g.caller,
		Pid:    g.callerPID,
		Tags:   append([]string(nil), g.tags...),
		Server: g.serverName,
	})
	if err != nil {
		if status.Code(err) == codes.Unavailable {
			g.clients.reset()
		}
		return err
	}
	if !g.registered.Swap(true) && resp != nil && resp.GetProfile() != "" {
		g.logger.Info("caller registered", zap.String("profile", resp.GetProfile()))
	}
	g.markRegisterReady()
	return nil
}

func (g *Gateway) markRegisterReady() {
	g.registerReadyOnce.Do(func() {
		if g.registerReadyCh != nil {
			close(g.registerReadyCh)
		}
	})
}

func (g *Gateway) unregisterCaller(ctx context.Context) error {
	client, err := g.clients.get(ctx)
	if err != nil {
		return err
	}
	_, err = client.Control().UnregisterCaller(ctx, &controlv1.UnregisterCallerRequest{
		Caller: g.caller,
	})
	if err != nil {
		if status.Code(err) == codes.Unavailable {
			g.clients.reset()
		}
		return err
	}
	return nil
}

func (g *Gateway) toolsReadyMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if method == "tools/list" {
				g.waitForToolsReady(ctx)
			}
			return next(ctx, method, req)
		}
	}
}

func (g *Gateway) waitForToolsReady(ctx context.Context) {
	if g.toolsReadyCh == nil {
		return
	}
	select {
	case <-g.toolsReadyCh:
		return
	default:
	}
	wait := g.toolsReadyWait
	if wait <= 0 {
		return
	}
	waitCtx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()
	select {
	case <-g.toolsReadyCh:
		return
	case <-waitCtx.Done():
		if errors.Is(waitCtx.Err(), context.DeadlineExceeded) && !g.toolsReadyWarn.Swap(true) {
			g.logger.Warn("tools snapshot not ready before tools/list", zap.Duration("wait", wait))
		}
		return
	}
}

func (g *Gateway) markToolsReady() {
	g.toolsReadyOnce.Do(func() {
		if g.toolsReadyCh != nil {
			close(g.toolsReadyCh)
		}
	})
}

// checkAndSetupSubAgent checks if SubAgent is enabled and registers builtin tools.
func (g *Gateway) checkAndSetupSubAgent(ctx context.Context) error {
	client, err := g.clients.get(ctx)
	if err != nil {
		return err
	}

	resp, err := client.Control().IsSubAgentEnabled(ctx, &controlv1.IsSubAgentEnabledRequest{
		Caller: g.caller,
	})
	if err != nil {
		if status.Code(err) == codes.Unimplemented {
			g.subAgentEnabled.Store(false)
			return nil
		}
		return err
	}

	if resp != nil && resp.GetEnabled() {
		g.subAgentEnabled.Store(true)
		g.registerSubAgentTools()
		g.logger.Info("SubAgent enabled, registered automatic_mcp and automatic_eval tools")
	} else {
		g.subAgentEnabled.Store(false)
	}

	return nil
}

// registerSubAgentTools registers the mcpv.automatic_mcp and mcpv.automatic_eval tools.
func (g *Gateway) registerSubAgentTools() {
	automaticMCPTool := subagent.AutomaticMCPTool()
	g.server.AddTool(&automaticMCPTool, g.automaticMCPHandler())

	automaticEvalTool := subagent.AutomaticEvalTool()
	g.server.AddTool(&automaticEvalTool, g.automaticEvalHandler())

	g.markToolsReady()
}
