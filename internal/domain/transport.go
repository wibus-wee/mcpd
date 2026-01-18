package domain

import (
	"context"
	"encoding/json"
	"io"
)

// IOStreams holds the reader and writer used by transports.
type IOStreams struct {
	Reader io.ReadCloser
	Writer io.WriteCloser
}

// Conn represents an active transport connection.
type Conn interface {
	Call(ctx context.Context, payload json.RawMessage) (json.RawMessage, error)
	Close() error
}

// StopFn stops a running instance.
type StopFn func(ctx context.Context) error

// Launcher starts a server process and returns its IO streams.
type Launcher interface {
	Start(ctx context.Context, specKey string, spec ServerSpec) (IOStreams, StopFn, error)
}

// Transport connects to a server using the provided streams.
type Transport interface {
	Connect(ctx context.Context, specKey string, spec ServerSpec, streams IOStreams) (Conn, error)
}
