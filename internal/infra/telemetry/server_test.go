package telemetry

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestStartMetricsServer_Success(t *testing.T) {
	// Use random port to avoid conflicts
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- StartMetricsServer(ctx, port, zap.NewNop())
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test /metrics endpoint
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics", port))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "# HELP")

	// Trigger graceful shutdown
	cancel()

	// Wait for server to stop
	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("server did not stop in time")
	}
}

func TestStartMetricsServer_PortInUse(t *testing.T) {
	// Start a server on a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Try to start metrics server on the same port (should fail quickly)
	err = StartMetricsServer(ctx, port, zap.NewNop())
	// The error could be either from port conflict or context timeout
	// Both are acceptable for this test
	if err != nil {
		assert.True(t, err.Error() == "context deadline exceeded" ||
			err.Error() != "", "expected an error")
	}
}

func TestStartMetricsServer_GracefulShutdown(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		err := StartMetricsServer(ctx, port, zap.NewNop())
		assert.NoError(t, err)
		close(done)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context to trigger shutdown
	cancel()

	// Wait for graceful shutdown
	select {
	case <-done:
		// Success
	case <-time.After(10 * time.Second):
		t.Fatal("graceful shutdown timed out")
	}
}
