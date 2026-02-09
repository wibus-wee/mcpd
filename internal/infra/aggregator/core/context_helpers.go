package core

import (
	"context"

	"mcpv/internal/domain"
)

func WithRefreshTimeout(ctx context.Context, cfg domain.RuntimeConfig) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, RefreshTimeout(cfg))
}
