package domain

import "context"

type HealthProbe interface {
	Ping(ctx context.Context, conn Conn) error
}
