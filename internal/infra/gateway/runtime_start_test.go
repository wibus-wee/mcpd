package gateway

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"mcpv/internal/infra/rpc"
)

func TestGatewayStartRuntime_FailsWhenRegisterCallerFails(t *testing.T) {
	g := NewGateway(rpc.ClientConfig{
		Address:        "127.0.0.1:1",
		MaxRecvMsgSize: 1024,
		MaxSendMsgSize: 1024,
	}, "caller", nil, "", zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := g.StartRuntime(ctx)
	require.Error(t, err)
	require.False(t, g.runtimeStarted)
}
