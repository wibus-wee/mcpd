package domain

import (
	"context"
	"time"
)

type Scheduler interface {
	Acquire(ctx context.Context, specKey, routingKey string) (*Instance, error)
	AcquireReady(ctx context.Context, specKey, routingKey string) (*Instance, error)
	Release(ctx context.Context, instance *Instance) error
	SetDesiredMinReady(ctx context.Context, specKey string, minReady int) error
	StopSpec(ctx context.Context, specKey, reason string) error
	StartIdleManager(interval time.Duration)
	StopIdleManager()
	StartPingManager(interval time.Duration)
	StopPingManager()
	StopAll(ctx context.Context)
	GetPoolStatus(ctx context.Context) ([]PoolInfo, error)
}
