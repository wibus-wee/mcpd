package app

import (
	"context"
	"time"

	"go.uber.org/zap"

	"mcpd/internal/infra/catalog"
	"mcpd/internal/infra/lifecycle"
	"mcpd/internal/infra/router"
	"mcpd/internal/infra/scheduler"
	"mcpd/internal/infra/server"
	"mcpd/internal/infra/transport"
)

type App struct {
	logger *zap.Logger
}

type ServeConfig struct {
	ConfigPath string
}

type ValidateConfig struct {
	ConfigPath string
}

func New(logger *zap.Logger) *App {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &App{
		logger: logger.Named("app"),
	}
}

func (a *App) Serve(ctx context.Context, cfg ServeConfig) error {
	loader := catalog.NewLoader(a.logger)

	specs, err := loader.Load(ctx, cfg.ConfigPath)
	if err != nil {
		return err
	}

	a.logger.Info("configuration loaded", zap.String("config", cfg.ConfigPath), zap.Int("servers", len(specs)))

	stdioTransport := transport.NewStdioTransport()
	lc := lifecycle.NewManager(stdioTransport, a.logger)
	sched := scheduler.NewBasicScheduler(lc, specs)
	rt := router.NewBasicRouter(sched)

	sched.StartIdleManager(time.Second)
	defer func() {
		sched.StopIdleManager()
		sched.StopAll(context.Background())
	}()

	return server.Run(ctx, rt, a.logger)
}

func (a *App) ValidateConfig(ctx context.Context, cfg ValidateConfig) error {
	loader := catalog.NewLoader(a.logger)

	specs, err := loader.Load(ctx, cfg.ConfigPath)
	if err != nil {
		return err
	}

	a.logger.Info("configuration validated", zap.String("config", cfg.ConfigPath), zap.Int("servers", len(specs)))
	return nil
}
