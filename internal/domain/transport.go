package domain

import (
	"context"
	"encoding/json"
	"io"
)

type IOStreams struct {
	Reader io.ReadCloser
	Writer io.WriteCloser
}

type Conn interface {
	Call(ctx context.Context, payload json.RawMessage) (json.RawMessage, error)
	Close() error
}

type StopFn func(ctx context.Context) error

type Launcher interface {
	Start(ctx context.Context, specKey string, spec ServerSpec) (IOStreams, StopFn, error)
}

type Transport interface {
	Connect(ctx context.Context, specKey string, spec ServerSpec, streams IOStreams) (Conn, error)
}
