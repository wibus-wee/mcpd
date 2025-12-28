package transport

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"mcpd/internal/domain"
	"mcpd/internal/infra/telemetry"
)

func TestStdioTransport_StartAndRoundTrip(t *testing.T) {
	transport := NewStdioTransport(StdioTransportOptions{})
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
	transport := NewStdioTransport(StdioTransportOptions{})
	spec := domain.ServerSpec{
		Name:            "bad",
		Cmd:             []string{},
		ProtocolVersion: domain.DefaultProtocolVersion,
	}

	_, _, err := transport.Start(context.Background(), spec)
	require.Error(t, err)
}

func TestStdioTransport_StopKillsProcess(t *testing.T) {
	transport := NewStdioTransport(StdioTransportOptions{})
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

func TestStdioTransport_MirrorsStderr(t *testing.T) {
	logs := telemetry.NewLogBroadcaster(zapcore.InfoLevel)
	logger := zap.New(zapcore.NewTee(zapcore.NewNopCore(), logs.Core()))
	logger = logger.With(zap.String(telemetry.FieldLogSource, telemetry.LogSourceCore))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	logCh := logs.Subscribe(ctx)

	transport := NewStdioTransport(StdioTransportOptions{Logger: logger})
	spec := domain.ServerSpec{
		Name:            "stderr",
		Cmd:             []string{"/bin/sh", "-c", "echo \"stderr line\" 1>&2; cat"},
		MaxConcurrent:   1,
		IdleSeconds:     0,
		MinReady:        0,
		ProtocolVersion: domain.DefaultProtocolVersion,
	}

	conn, stop, err := transport.Start(ctx, spec)
	require.NoError(t, err)
	defer stop(context.Background())
	_ = conn.Close()

	entry := waitForDownstreamLog(t, logCh)
	require.NotEmpty(t, entry.DataJSON)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(entry.DataJSON, &payload))

	fields, ok := payload["fields"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, telemetry.LogSourceDownstream, fields[telemetry.FieldLogSource])
	require.Equal(t, "stderr", fields[telemetry.FieldLogStream])
	require.Equal(t, spec.Name, fields[telemetry.FieldServerType])
	require.Equal(t, "stderr line", payload["message"])
}

func waitForDownstreamLog(t *testing.T, logCh <-chan domain.LogEntry) domain.LogEntry {
	t.Helper()

	deadline := time.After(2 * time.Second)
	for {
		select {
		case entry := <-logCh:
			if len(entry.DataJSON) == 0 {
				continue
			}
			var payload map[string]any
			if err := json.Unmarshal(entry.DataJSON, &payload); err != nil {
				continue
			}
			fields, ok := payload["fields"].(map[string]any)
			if !ok {
				continue
			}
			if fields[telemetry.FieldLogSource] == telemetry.LogSourceDownstream {
				return entry
			}
		case <-deadline:
			t.Fatal("timed out waiting for downstream stderr log")
		}
	}
}
