package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"mcpv/internal/domain"
)

func TestBasicScheduler_IdleReapRespectsMinReady(t *testing.T) {
	lc := &fakeLifecycle{}
	spec := newTestSpec("svc")
	spec.IdleSeconds = 0
	spec.MinReady = 1
	spec.Strategy = domain.StrategyStateless

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})
	require.NoError(t, s.SetDesiredMinReady(context.Background(), "svc", 1))

	inst, err := s.Acquire(context.Background(), "svc", "")
	require.NoError(t, err)
	require.NotNil(t, inst)
	require.NoError(t, s.Release(context.Background(), inst))

	s.reapIdle()
	require.Equal(t, domain.InstanceStateReady, inst.State())
}

func TestBasicScheduler_IdleReapStopsWhenBelowMinReady(t *testing.T) {
	lc := &fakeLifecycle{}
	spec := newTestSpec("svc")
	spec.IdleSeconds = 0
	spec.MinReady = 0
	spec.Strategy = domain.StrategyStateless

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})

	inst, err := s.Acquire(context.Background(), "svc", "")
	require.NoError(t, err)
	require.NoError(t, s.Release(context.Background(), inst))

	s.reapIdle()
	require.Equal(t, domain.InstanceStateStopped, inst.State())
}

func TestBasicScheduler_StatefulWithBindingSkipsIdle(t *testing.T) {
	lc := &fakeLifecycle{}
	spec := newTestSpec("svc")
	spec.MaxConcurrent = 1
	spec.IdleSeconds = 0
	spec.Strategy = domain.StrategyStateful
	spec.SessionTTLSeconds = 3600

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})

	inst, err := s.Acquire(context.Background(), "svc", "rk")
	require.NoError(t, err)
	require.NoError(t, s.Release(context.Background(), inst))

	s.reapIdle()
	require.Equal(t, domain.InstanceStateReady, inst.State())
}

func TestBasicScheduler_IdleReapIgnoresIdleSecondsWhenMinReadyZero(t *testing.T) {
	lc := &fakeLifecycle{}
	spec := newTestSpec("svc")
	spec.IdleSeconds = 3600
	spec.MinReady = 0
	spec.Strategy = domain.StrategyStateless

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})

	inst, err := s.Acquire(context.Background(), "svc", "")
	require.NoError(t, err)
	require.NoError(t, s.Release(context.Background(), inst))

	s.reapIdle()
	require.Equal(t, domain.InstanceStateStopped, inst.State())
}

func TestBasicScheduler_StatefulSessionTTLLimitsBindings(t *testing.T) {
	lc := &fakeLifecycle{}
	spec := newTestSpec("svc")
	spec.MaxConcurrent = 1
	spec.IdleSeconds = 0
	spec.Strategy = domain.StrategyStateful
	spec.SessionTTLSeconds = 1

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{
		Logger: zap.NewNop(),
	})

	inst, err := s.Acquire(context.Background(), "svc", "rk")
	require.NoError(t, err)
	require.Equal(t, "rk", inst.StickyKey())
	require.NoError(t, s.Release(context.Background(), inst))

	state := s.getPool("svc", spec)
	state.mu.Lock()
	binding := state.sticky["rk"]
	require.NotNil(t, binding)
	binding.lastAccess = time.Now().Add(-2 * time.Second)
	state.mu.Unlock()

	s.reapStaleBindings()

	state.mu.Lock()
	_, exists := state.sticky["rk"]
	state.mu.Unlock()
	require.False(t, exists)
	require.Equal(t, "", inst.StickyKey())
}

func TestBasicScheduler_StatefulSessionTTLZeroKeepsBindings(t *testing.T) {
	lc := &fakeLifecycle{}
	spec := newTestSpec("svc")
	spec.MaxConcurrent = 1
	spec.IdleSeconds = 0
	spec.Strategy = domain.StrategyStateful
	spec.SessionTTLSeconds = 0

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})

	inst, err := s.Acquire(context.Background(), "svc", "rk")
	require.NoError(t, err)
	require.NoError(t, s.Release(context.Background(), inst))

	state := s.getPool("svc", spec)
	state.mu.Lock()
	binding := state.sticky["rk"]
	require.NotNil(t, binding)
	binding.lastAccess = time.Now().Add(-2 * time.Second)
	state.mu.Unlock()

	s.reapStaleBindings()

	state.mu.Lock()
	_, exists := state.sticky["rk"]
	state.mu.Unlock()
	require.True(t, exists)
	require.Equal(t, "rk", inst.StickyKey())
}

func TestBasicScheduler_IdleReapSkipsPersistentAndSingleton(t *testing.T) {
	cases := []struct {
		name     string
		strategy domain.InstanceStrategy
	}{
		{name: "persistent", strategy: domain.StrategyPersistent},
		{name: "singleton", strategy: domain.StrategySingleton},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lc := &fakeLifecycle{}
			spec := newTestSpec("svc")
			spec.MaxConcurrent = 1
			spec.IdleSeconds = 0
			spec.Strategy = tc.strategy

			s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})

			inst, err := s.Acquire(context.Background(), "svc", "")
			require.NoError(t, err)
			require.NoError(t, s.Release(context.Background(), inst))

			s.reapIdle()
			require.Equal(t, domain.InstanceStateReady, inst.State())

			state := s.getPool("svc", spec)
			state.mu.Lock()
			require.Len(t, state.instances, 1)
			state.mu.Unlock()
		})
	}
}

func TestBasicScheduler_PingFailureStopsInstance(t *testing.T) {
	lc := &fakeLifecycle{}
	spec := newTestSpec("svc")
	spec.MaxConcurrent = 1
	spec.IdleSeconds = 10

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{
		Probe:  &fakeProbe{err: errors.New("ping failed")},
		Logger: zap.NewNop(),
	})

	inst, err := s.Acquire(context.Background(), "svc", "")
	require.NoError(t, err)
	require.NoError(t, s.Release(context.Background(), inst))

	s.probeInstances()

	require.Equal(t, domain.InstanceStateStopped, inst.State())
	specKey := domain.SpecFingerprint(spec)
	state := s.getPool(specKey, spec)
	state.mu.Lock()
	defer state.mu.Unlock()
	require.Len(t, state.instances, 0)
}
