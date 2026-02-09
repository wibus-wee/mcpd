package gateway

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"mcpv/internal/infra/rpc"
)

func TestGatewayWaitForToolsReady_UnblocksAfterReady(t *testing.T) {
	g := NewGateway(rpc.ClientConfig{}, "caller", nil, "", zap.NewNop())
	g.toolsReadyWait = 500 * time.Millisecond

	done := make(chan struct{})
	go func() {
		g.waitForToolsReady(context.Background())
		close(done)
	}()

	time.Sleep(10 * time.Millisecond)
	g.markToolsReady()

	select {
	case <-done:
		return
	case <-time.After(200 * time.Millisecond):
		t.Fatal("waitForToolsReady did not return after tools became ready")
	}
}
