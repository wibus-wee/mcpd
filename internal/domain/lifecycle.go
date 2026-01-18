package domain

import "context"

// Lifecycle manages starting and stopping server instances.
type Lifecycle interface {
	StartInstance(ctx context.Context, specKey string, spec ServerSpec) (*Instance, error)
	StopInstance(ctx context.Context, instance *Instance, reason string) error
}
