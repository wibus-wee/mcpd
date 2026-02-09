package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"mcpv/internal/domain"
)

func TestApplyCatalogDiff_PoolsMapCleaned(t *testing.T) {
	lc := &fakeLifecycle{}
	initialSpecs := map[string]domain.ServerSpec{
		"svc-a": {Name: "svc-a", Cmd: []string{"./a"}, MaxConcurrent: 1, ProtocolVersion: domain.DefaultProtocolVersion},
		"svc-b": {Name: "svc-b", Cmd: []string{"./b"}, MaxConcurrent: 1, ProtocolVersion: domain.DefaultProtocolVersion},
	}
	s := newScheduler(t, lc, initialSpecs, Options{})

	instA, err := s.Acquire(context.Background(), "svc-a", "")
	require.NoError(t, err)
	require.NoError(t, s.Release(context.Background(), instA))

	instB, err := s.Acquire(context.Background(), "svc-b", "")
	require.NoError(t, err)
	require.NoError(t, s.Release(context.Background(), instB))

	s.poolsMu.RLock()
	poolCountBefore := len(s.pools)
	s.poolsMu.RUnlock()
	require.Equal(t, 2, poolCountBefore)

	diff := domain.CatalogDiff{
		RemovedSpecKeys: []string{"svc-a"},
	}
	newRegistry := map[string]domain.ServerSpec{
		"svc-b": initialSpecs["svc-b"],
	}
	require.NoError(t, s.ApplyCatalogDiff(context.Background(), diff, newRegistry))

	require.Equal(t, domain.InstanceStateStopped, instA.State())

	s.poolsMu.RLock()
	poolCountAfter := len(s.pools)
	_, svcAPoolExists := s.pools["svc-a"]
	s.poolsMu.RUnlock()
	require.False(t, svcAPoolExists)
	require.Equal(t, poolCountBefore-1, poolCountAfter)
}

func TestApplyCatalogDiff_ContextTimeoutLeavesPartialState(t *testing.T) {
	const stopDelay = 100 * time.Millisecond

	lc := &slowStopLifecycle{stopDelay: stopDelay}
	specs := map[string]domain.ServerSpec{
		"svc-a": {Name: "svc-a", Cmd: []string{"./a"}, MaxConcurrent: 1, ProtocolVersion: domain.DefaultProtocolVersion},
		"svc-b": {Name: "svc-b", Cmd: []string{"./b"}, MaxConcurrent: 1, ProtocolVersion: domain.DefaultProtocolVersion},
		"svc-c": {Name: "svc-c", Cmd: []string{"./c"}, MaxConcurrent: 1, ProtocolVersion: domain.DefaultProtocolVersion},
	}
	s := newScheduler(t, lc, specs, Options{})

	instances := make(map[string]*domain.Instance)
	for key := range specs {
		inst, err := s.Acquire(context.Background(), key, "")
		require.NoError(t, err)
		require.NoError(t, s.Release(context.Background(), inst))
		instances[key] = inst
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	diff := domain.CatalogDiff{
		RemovedSpecKeys: []string{"svc-a", "svc-b", "svc-c"},
	}
	err := s.ApplyCatalogDiff(ctx, diff, map[string]domain.ServerSpec{})

	stoppedCount := 0
	for _, inst := range instances {
		if inst.State() == domain.InstanceStateStopped {
			stoppedCount++
		}
	}

	t.Logf("stopped %d/%d instances before timeout/completion", stoppedCount, len(instances))
	t.Logf("ApplyCatalogDiff returned: %v", err)
}
