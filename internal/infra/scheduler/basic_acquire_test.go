package scheduler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"mcpv/internal/domain"
)

func TestBasicScheduler_StartsAndReusesInstance(t *testing.T) {
	lc := &fakeLifecycle{}
	spec := newTestSpec("svc")
	spec.MaxConcurrent = 2
	spec.IdleSeconds = 10

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})

	inst1, err := s.Acquire(context.Background(), "svc", "")
	require.NoError(t, err)
	require.Equal(t, 1, inst1.BusyCount())

	inst2, err := s.Acquire(context.Background(), "svc", "")
	require.NoError(t, err)
	require.Same(t, inst1, inst2)
	require.Equal(t, 2, inst1.BusyCount())

	require.NoError(t, s.Release(context.Background(), inst1))
	require.Equal(t, domain.InstanceStateBusy, inst1.State())
	require.NoError(t, s.Release(context.Background(), inst1))
	require.Equal(t, domain.InstanceStateReady, inst1.State())
}

func TestBasicScheduler_StickyBinding(t *testing.T) {
	lc := &fakeLifecycle{}
	spec := newTestSpec("svc")
	spec.MaxConcurrent = 1
	spec.Strategy = domain.StrategyStateful

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})

	inst1, err := s.Acquire(context.Background(), "svc", "userA")
	require.NoError(t, err)

	inst2, err := s.Acquire(context.Background(), "svc", "userA")
	require.ErrorIs(t, err, ErrStickyBusy)
	require.Nil(t, inst2)

	_ = s.Release(context.Background(), inst1)

	inst3, err := s.Acquire(context.Background(), "svc", "userA")
	require.NoError(t, err)
	require.Same(t, inst1, inst3)
}

func TestBasicScheduler_UnknownServer(t *testing.T) {
	s := newScheduler(t, &fakeLifecycle{}, map[string]domain.ServerSpec{}, Options{})

	_, err := s.Acquire(context.Background(), "missing", "")
	require.ErrorIs(t, err, ErrUnknownSpecKey)
}

func TestBasicScheduler_AcquireReadyDoesNotStart(t *testing.T) {
	lc := &fakeLifecycle{}
	spec := newTestSpec("svc")

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})

	_, err := s.AcquireReady(context.Background(), "svc", "")
	require.ErrorIs(t, err, domain.ErrNoReadyInstance)
	require.Equal(t, 0, lc.counter)
}

func TestBasicScheduler_StatelessSelection(t *testing.T) {
	t.Run("round_robin_cycle", func(t *testing.T) {
		lc := &countingLifecycle{}
		spec := newTestSpec("svc")
		spec.MaxConcurrent = 2
		spec.Strategy = domain.StrategyStateless

		s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})
		require.NoError(t, s.SetDesiredMinReady(context.Background(), "svc", 3))

		state := s.getPool("svc", spec)
		state.mu.Lock()
		require.Len(t, state.instances, 3)
		instA := state.instances[0].instance
		instB := state.instances[1].instance
		instC := state.instances[2].instance
		state.mu.Unlock()

		cases := []struct {
			name string
			want *domain.Instance
		}{
			{name: "first", want: instA},
			{name: "second", want: instB},
			{name: "third", want: instC},
			{name: "wrap", want: instA},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				selected, err := s.AcquireReady(context.Background(), "svc", "")
				require.NoError(t, err)
				require.Equal(t, tc.want.ID(), selected.ID())
				require.NoError(t, s.Release(context.Background(), selected))
			})
		}
	})

	t.Run("least_loaded_prefers_lower_busy", func(t *testing.T) {
		lc := &countingLifecycle{}
		spec := newTestSpec("svc")
		spec.MaxConcurrent = 10
		spec.Strategy = domain.StrategyStateless

		s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})
		require.NoError(t, s.SetDesiredMinReady(context.Background(), "svc", 3))

		state := s.getPool("svc", spec)
		state.mu.Lock()
		require.Len(t, state.instances, 3)
		instA := state.instances[0].instance
		instB := state.instances[1].instance
		instC := state.instances[2].instance

		instA.SetBusyCount(2)
		instA.SetState(domain.InstanceStateBusy)
		instB.SetBusyCount(1)
		instB.SetState(domain.InstanceStateBusy)
		instC.SetBusyCount(0)
		instC.SetState(domain.InstanceStateReady)
		state.mu.Unlock()

		selected, err := s.AcquireReady(context.Background(), "svc", "")
		require.NoError(t, err)
		require.Equal(t, instC.ID(), selected.ID())
	})

	t.Run("skip_unroutable_and_full", func(t *testing.T) {
		lc := &countingLifecycle{}
		spec := newTestSpec("svc")
		spec.MaxConcurrent = 2
		spec.Strategy = domain.StrategyStateless

		s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})
		require.NoError(t, s.SetDesiredMinReady(context.Background(), "svc", 3))

		state := s.getPool("svc", spec)
		state.mu.Lock()
		require.Len(t, state.instances, 3)
		instA := state.instances[0].instance
		instB := state.instances[1].instance
		instC := state.instances[2].instance

		instA.SetState(domain.InstanceStateFailed)
		instB.SetBusyCount(spec.MaxConcurrent)
		instB.SetState(domain.InstanceStateBusy)
		instC.SetBusyCount(0)
		instC.SetState(domain.InstanceStateReady)
		state.mu.Unlock()

		selected, err := s.AcquireReady(context.Background(), "svc", "")
		require.NoError(t, err)
		require.Equal(t, instC.ID(), selected.ID())
	})
}

func TestBasicScheduler_SharedPool(t *testing.T) {
	lc := &fakeLifecycle{}
	specA := domain.ServerSpec{
		Name:            "svc-a",
		Cmd:             []string{"./svc"},
		MaxConcurrent:   2,
		IdleSeconds:     10,
		ProtocolVersion: domain.DefaultProtocolVersion,
	}
	specB := specA
	specB.Name = "svc-b"

	specKeyA := domain.SpecFingerprint(specA)
	specKeyB := domain.SpecFingerprint(specB)
	require.Equal(t, specKeyA, specKeyB)

	s := newScheduler(t, lc, map[string]domain.ServerSpec{
		specKeyA: specA,
	}, Options{})

	instA, err := s.Acquire(context.Background(), specKeyA, "")
	require.NoError(t, err)
	instB, err := s.Acquire(context.Background(), specKeyB, "")
	require.NoError(t, err)
	require.Same(t, instA, instB)
}
