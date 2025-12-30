package domain

import "context"

type Lifecycle interface {
	StartInstance(ctx context.Context, spec ServerSpec) (*Instance, error)
	StopInstance(ctx context.Context, instance *Instance, reason string) error
}
