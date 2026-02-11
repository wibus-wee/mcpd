package gateway

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"mcpv/internal/infra/rpc"
)

type fakeRuntime struct {
	server     *mcp.Server
	startCount atomic.Int32
	stopCount  atomic.Int32
}

func (f *fakeRuntime) StartRuntime(_ context.Context) error {
	f.startCount.Add(1)
	return nil
}

func (f *fakeRuntime) StopRuntime(_ context.Context) error {
	f.stopCount.Add(1)
	return nil
}

func (f *fakeRuntime) Server() *mcp.Server {
	return f.server
}

func TestGatewayPool_ReusesRuntime(t *testing.T) {
	created := atomic.Int32{}
	factory := func(_ Selector, _ string) runtime {
		created.Add(1)
		return &fakeRuntime{server: mcp.NewServer(&mcp.Implementation{Name: "fake", Version: "test"}, nil)}
	}

	pool := newGatewayPool(context.Background(), rpc.ClientConfig{}, "base", zap.NewNop(), PoolOptions{RuntimeFactory: factory})
	sel := Selector{Server: "context7"}

	serverA, err := pool.Get(context.Background(), sel)
	require.NoError(t, err)
	serverB, err := pool.Get(context.Background(), sel)
	require.NoError(t, err)
	require.Same(t, serverA, serverB)
	require.Equal(t, int32(1), created.Load())
}

func TestGatewayPool_EvictsIdle(t *testing.T) {
	created := atomic.Int32{}
	fake := &fakeRuntime{server: mcp.NewServer(&mcp.Implementation{Name: "fake", Version: "test"}, nil)}
	factory := func(_ Selector, _ string) runtime {
		created.Add(1)
		return fake
	}

	pool := newGatewayPool(context.Background(), rpc.ClientConfig{}, "base", zap.NewNop(), PoolOptions{
		IdleTimeout:    10 * time.Millisecond,
		RuntimeFactory: factory,
	})

	sel := Selector{Server: "context7"}
	_, err := pool.Get(context.Background(), sel)
	require.NoError(t, err)

	key := SelectorKey(sel)
	pool.mu.Lock()
	if runtime, ok := pool.runtimes[key]; ok {
		runtime.lastUsed = time.Now().Add(-time.Minute)
	}
	pool.mu.Unlock()

	pool.evictIdle(time.Now())
	pool.mu.Lock()
	_, ok := pool.runtimes[key]
	pool.mu.Unlock()
	require.False(t, ok)
	require.Equal(t, int32(1), fake.stopCount.Load())
}

func TestGatewayPool_MaxInstances(t *testing.T) {
	factory := func(_ Selector, _ string) runtime {
		return &fakeRuntime{server: mcp.NewServer(&mcp.Implementation{Name: "fake", Version: "test"}, nil)}
	}

	pool := newGatewayPool(context.Background(), rpc.ClientConfig{}, "base", zap.NewNop(), PoolOptions{
		MaxInstances:   1,
		RuntimeFactory: factory,
	})

	_, err := pool.Get(context.Background(), Selector{Server: "a"})
	require.NoError(t, err)
	_, err = pool.Get(context.Background(), Selector{Server: "b"})
	require.Error(t, err)
}
