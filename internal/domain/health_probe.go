package domain

import "context"

// HealthProbe checks instance liveness.
type HealthProbe interface {
	Ping(ctx context.Context, conn Conn) error
}
