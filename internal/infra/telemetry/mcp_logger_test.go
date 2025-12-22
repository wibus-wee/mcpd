package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestMCPLogSink_SendsNotificationsAfterSetLevel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sink := NewMCPLogSink("mcpd", zapcore.DebugLevel)
	logger := zap.New(sink.Core())

	server := mcp.NewServer(&mcp.Implementation{Name: "mcpd", Version: "0.1.0"}, nil)
	sink.SetServer(server)

	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	_, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)

	msgCh := make(chan *mcp.LoggingMessageParams, 1)
	client := mcp.NewClient(&mcp.Implementation{Name: "client", Version: "0.1.0"}, &mcp.ClientOptions{
		LoggingMessageHandler: func(_ context.Context, req *mcp.LoggingMessageRequest) {
			if req != nil {
				msgCh <- req.Params
			}
		},
	})
	cs, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer cs.Close()

	require.NoError(t, cs.SetLoggingLevel(ctx, &mcp.SetLoggingLevelParams{Level: "info"}))

	logger.Info("hello", zap.String("serverType", "test"))

	select {
	case msg := <-msgCh:
		require.Equal(t, mcp.LoggingLevel("info"), msg.Level)
		payload, ok := msg.Data.(map[string]any)
		require.True(t, ok)
		require.Equal(t, "hello", payload["message"])
	case <-time.After(1 * time.Second):
		t.Fatalf("expected logging notification")
	}
}
