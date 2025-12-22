package transport

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"mcpd/internal/domain"
)

func TestStdioTransport_StartAndRoundTrip(t *testing.T) {
	transport := NewStdioTransport()
	spec := domain.ServerSpec{
		Name:            "cat",
		Cmd:             []string{"/bin/sh", "-c", "cat"},
		MaxConcurrent:   1,
		IdleSeconds:     0,
		MinReady:        0,
		ProtocolVersion: domain.DefaultProtocolVersion,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, stop, err := transport.Start(ctx, spec)
	require.NoError(t, err)
	defer stop(context.Background())

	msg := json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"ping"}`)
	require.NoError(t, conn.Send(ctx, msg))

	got, err := conn.Recv(ctx)
	require.NoError(t, err)
	require.JSONEq(t, string(msg), string(got))
}

func TestStdioTransport_InvalidCmd(t *testing.T) {
	transport := NewStdioTransport()
	spec := domain.ServerSpec{
		Name:            "bad",
		Cmd:             []string{},
		ProtocolVersion: domain.DefaultProtocolVersion,
	}

	_, _, err := transport.Start(context.Background(), spec)
	require.Error(t, err)
}

func TestStdioTransport_StopKillsProcess(t *testing.T) {
	transport := NewStdioTransport()
	spec := domain.ServerSpec{
		Name:            "sleep",
		Cmd:             []string{"/bin/sh", "-c", "sleep 10"},
		ProtocolVersion: domain.DefaultProtocolVersion,
		MaxConcurrent:   1,
	}

	conn, stop, err := transport.Start(context.Background(), spec)
	require.NoError(t, err)
	_ = conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = stop(ctx)
	require.NoError(t, err)
}
