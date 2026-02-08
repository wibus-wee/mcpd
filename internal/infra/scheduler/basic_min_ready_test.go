package scheduler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"mcpv/internal/domain"
)

func TestBasicScheduler_SetDesiredMinReadyStartsInstance(t *testing.T) {
	lc := &fakeLifecycle{}
	spec := newTestSpec("svc")

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})

	require.NoError(t, s.SetDesiredMinReady(context.Background(), "svc", 1))

	state := s.getPool("svc", spec)
	state.mu.Lock()
	defer state.mu.Unlock()
	require.Len(t, state.instances, 1)
}

func TestBasicScheduler_SetDesiredMinReady_PanicDoesNotLeakStarting(t *testing.T) {
	lc := &panicLifecycle{}
	spec := newTestSpec("svc")

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})

	func() {
		defer func() {
			require.NotNil(t, recover())
		}()
		_ = s.SetDesiredMinReady(context.Background(), "svc", 1)
	}()

	require.NoError(t, s.SetDesiredMinReady(context.Background(), "svc", 1))
	require.Equal(t, 2, lc.starts())

	inst, err := s.AcquireReady(context.Background(), "svc", "")
	require.NoError(t, err)
	require.NotNil(t, inst)
}
