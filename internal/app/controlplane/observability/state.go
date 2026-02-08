package observability

import (
	"context"

	"go.uber.org/zap"

	"mcpv/internal/app/bootstrap"
	"mcpv/internal/domain"
)

type State interface {
	Scheduler() domain.Scheduler
	Startup() *bootstrap.ServerStartupOrchestrator
	Context() context.Context
	Logger() *zap.Logger
}
