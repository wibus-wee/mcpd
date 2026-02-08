package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"mcpv/internal/domain"
)

func TestBasicScheduler_StartGateSingleflight(t *testing.T) {
	started := make(chan struct{}, 3)
	release := make(chan struct{})
	lc := &blockingLifecycle{
		started: started,
		release: release,
	}
	spec := newTestSpec("svc")
	spec.MaxConcurrent = 3

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})

	results := make(chan *domain.Instance, 3)
	errorsCh := make(chan error, 3)
	for i := 0; i < 3; i++ {
		go func() {
			inst, err := s.Acquire(context.Background(), "svc", "")
			results <- inst
			errorsCh <- err
		}()
	}

	<-started
	close(release)

	var instances []*domain.Instance
	for i := 0; i < 3; i++ {
		require.NoError(t, <-errorsCh)
		instances = append(instances, <-results)
	}
	require.Equal(t, 1, lc.starts())
	require.Same(t, instances[0], instances[1])
	require.Same(t, instances[0], instances[2])
}

func TestBasicScheduler_SingletonBusyWaitsInsteadOfStarting(t *testing.T) {
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	lc := &blockingLifecycle{
		started: started,
		release: release,
	}
	spec := newTestSpec("svc")
	spec.MaxConcurrent = 1
	spec.Strategy = domain.StrategySingleton

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})

	instACh := make(chan *domain.Instance, 1)
	errACh := make(chan error, 1)
	go func() {
		inst, err := s.Acquire(context.Background(), "svc", "")
		instACh <- inst
		errACh <- err
	}()

	<-started
	close(release)

	require.NoError(t, <-errACh)
	instA := <-instACh

	ctxB, cancelB := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancelB()
	instBCh := make(chan *domain.Instance, 1)
	errBCh := make(chan error, 1)
	go func() {
		inst, err := s.Acquire(ctxB, "svc", "")
		instBCh <- inst
		errBCh <- err
	}()

	select {
	case <-started:
		t.Fatal("unexpected start while singleton is busy")
	case <-time.After(50 * time.Millisecond):
	}

	require.NoError(t, s.Release(context.Background(), instA))
	require.NoError(t, <-errBCh)
	instB := <-instBCh
	require.Same(t, instA, instB)
	require.Equal(t, 1, lc.starts())
}

func TestBasicScheduler_WaitersWakeAfterStart(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	lc := &blockingLifecycle{
		started: started,
		release: release,
	}
	spec := newTestSpec("svc")
	spec.MaxConcurrent = 3

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})

	results := make(chan *domain.Instance, 3)
	errorsCh := make(chan error, 3)
	for i := 0; i < 3; i++ {
		go func() {
			inst, err := s.Acquire(context.Background(), "svc", "")
			results <- inst
			errorsCh <- err
		}()
	}

	<-started
	time.Sleep(20 * time.Millisecond)
	close(release)

	for i := 0; i < 3; i++ {
		require.NoError(t, <-errorsCh)
		require.NotNil(t, <-results)
	}
	require.Equal(t, 1, lc.starts())
}

func TestBasicScheduler_WaitersWakeAfterStartFailure(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	lc := &failOnceBlockingLifecycle{
		started: started,
		release: release,
	}
	spec := newTestSpec("svc")
	spec.MaxConcurrent = 3

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	results := make(chan *domain.Instance, 3)
	errorsCh := make(chan error, 3)
	for i := 0; i < 3; i++ {
		go func() {
			inst, err := s.Acquire(ctx, "svc", "")
			results <- inst
			errorsCh <- err
		}()
	}

	<-started
	close(release)

	successes := 0
	for i := 0; i < 3; i++ {
		err := <-errorsCh
		inst := <-results
		if err != nil {
			require.Nil(t, inst)
			continue
		}
		require.NotNil(t, inst)
		successes++
	}
	require.Equal(t, 2, lc.starts())
	require.Equal(t, 2, successes)
}

func TestBasicScheduler_Acquire_DetachesStartFromCallerCancel(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	lc := &ctxBlockingLifecycle{
		started: started,
		release: release,
	}
	spec := newTestSpec("svc")

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})

	ctx, cancel := context.WithCancel(context.Background())
	results := make(chan *domain.Instance, 1)
	errorsCh := make(chan error, 1)
	go func() {
		inst, err := s.Acquire(ctx, "svc", "")
		results <- inst
		errorsCh <- err
	}()

	<-started
	cancel()

	select {
	case <-lc.startCtx.Done():
		t.Fatal("start context should not be canceled when caller context is canceled")
	case <-time.After(50 * time.Millisecond):
	}

	close(release)
	require.NoError(t, <-errorsCh)
	require.NotNil(t, <-results)
}

func TestBasicScheduler_AcquireSupersededByStopSpec(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	lc := &blockingLifecycle{
		started: started,
		release: release,
	}
	spec := newTestSpec("svc")

	s := newScheduler(t, lc, map[string]domain.ServerSpec{"svc": spec}, Options{})

	results := make(chan *domain.Instance, 1)
	errorsCh := make(chan error, 1)
	go func() {
		inst, err := s.Acquire(context.Background(), "svc", "")
		results <- inst
		errorsCh <- err
	}()

	<-started
	require.NoError(t, s.StopSpec(context.Background(), "svc", "test supersede"))
	close(release)

	require.Nil(t, <-results)
	require.ErrorIs(t, <-errorsCh, ErrNoCapacity)
	require.Equal(t, 1, lc.stops())
}
