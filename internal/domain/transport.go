package domain

import (
	"context"
	"encoding/json"
)

type Conn interface {
	Send(ctx context.Context, msg json.RawMessage) error
	Recv(ctx context.Context) (json.RawMessage, error)
	Close() error
}

type StopFn func(ctx context.Context) error

type Transport interface {
	Start(ctx context.Context, spec ServerSpec) (Conn, StopFn, error)
}
