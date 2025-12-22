package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"mcpd/internal/domain"
)

func TestBasicScheduler_StartsAndReusesInstance(t *testing.T) {
	lc := &fakeLifecycle{}
	spec := domain.ServerSpec{
		Name:            "svc",
		Cmd:             []string{"./svc"},
		MaxConcurrent:   2,
		IdleSeconds:     10,
		MinReady:        0,
		ProtocolVersion: domain.DefaultProtocolVersion,
	}
	s := NewBasicScheduler(lc, map[string]domain.ServerSpec{"svc": spec})

	inst1, err := s.Acquire(context.Background(), "svc", "")
	require.NoError(t, err)
	require.Equal(t, 1, inst1.BusyCount)

	inst2, err := s.Acquire(context.Background(), "svc", "")
	require.NoError(t, err)
	require.Same(t, inst1, inst2)
	require.Equal(t, 2, inst1.BusyCount)

	require.NoError(t, s.Release(context.Background(), inst1))
	require.Equal(t, domain.InstanceStateBusy, inst1.State)
	require.NoError(t, s.Release(context.Background(), inst1))
	require.Equal(t, domain.InstanceStateReady, inst1.State)
}

func TestBasicScheduler_StickyBinding(t *testing.T) {
	lc := &fakeLifecycle{}
	spec := domain.ServerSpec{
		Name:            "svc",
		Cmd:             []string{"./svc"},
		MaxConcurrent:   1,
		Sticky:          true,
		ProtocolVersion: domain.DefaultProtocolVersion,
	}
	s := NewBasicScheduler(lc, map[string]domain.ServerSpec{"svc": spec})

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
	s := NewBasicScheduler(&fakeLifecycle{}, map[string]domain.ServerSpec{})
	_, err := s.Acquire(context.Background(), "missing", "")
	require.ErrorIs(t, err, ErrUnknownServerType)
}

type fakeLifecycle struct {
	counter int
}

func (f *fakeLifecycle) StartInstance(ctx context.Context, spec domain.ServerSpec) (*domain.Instance, error) {
	f.counter++
	return &domain.Instance{
		ID:         spec.Name + "-inst",
		Spec:       spec,
		State:      domain.InstanceStateReady,
		LastActive: time.Now(),
	}, nil
}

func (f *fakeLifecycle) StopInstance(ctx context.Context, instance *domain.Instance, reason string) error {
	return nil
}
