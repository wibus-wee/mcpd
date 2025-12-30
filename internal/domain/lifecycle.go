package domain

import "context"

type Lifecycle interface {
	StartInstance(ctx context.Context, specKey string, spec ServerSpec) (*Instance, error)
	StopInstance(ctx context.Context, instance *Instance, reason string) error
}
