package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"mcpv/internal/domain"
)

func newTestSpec(name string) domain.ServerSpec {
	return domain.ServerSpec{
		Name:            name,
		Cmd:             []string{"./" + name},
		MaxConcurrent:   1,
		ProtocolVersion: domain.DefaultProtocolVersion,
	}
}

func newScheduler(t *testing.T, lc domain.Lifecycle, specs map[string]domain.ServerSpec, opts Options) *BasicScheduler {
	t.Helper()

	s, err := NewBasicScheduler(lc, specs, opts)
	require.NoError(t, err)
	return s
}

type fakeLifecycle struct {
	counter int
}

func (f *fakeLifecycle) StartInstance(_ context.Context, specKey string, spec domain.ServerSpec) (*domain.Instance, error) {
	f.counter++
	return domain.NewInstance(domain.InstanceOptions{
		ID:         spec.Name + "-inst",
		Spec:       spec,
		SpecKey:    specKey,
		State:      domain.InstanceStateReady,
		LastActive: time.Now(),
	}), nil
}

func (f *fakeLifecycle) StopInstance(_ context.Context, instance *domain.Instance, _ string) error {
	if instance != nil {
		instance.SetState(domain.InstanceStateStopped)
	}
	return nil
}

type countingLifecycle struct {
	mu    sync.Mutex
	count int
}

func (c *countingLifecycle) StartInstance(_ context.Context, specKey string, spec domain.ServerSpec) (*domain.Instance, error) {
	c.mu.Lock()
	c.count++
	id := fmt.Sprintf("%s-%d", spec.Name, c.count)
	c.mu.Unlock()
	return domain.NewInstance(domain.InstanceOptions{
		ID:         id,
		Spec:       spec,
		SpecKey:    specKey,
		State:      domain.InstanceStateReady,
		LastActive: time.Now(),
	}), nil
}

func (c *countingLifecycle) StopInstance(_ context.Context, instance *domain.Instance, _ string) error {
	if instance != nil {
		instance.SetState(domain.InstanceStateStopped)
	}
	return nil
}

type fakeProbe struct {
	err error
}

func (f *fakeProbe) Ping(_ context.Context, _ domain.Conn) error {
	return f.err
}

type blockingLifecycle struct {
	mu      sync.Mutex
	started chan struct{}
	release chan struct{}
	count   int
	stopMu  sync.Mutex
	stopped int
}

func (b *blockingLifecycle) StartInstance(_ context.Context, specKey string, spec domain.ServerSpec) (*domain.Instance, error) {
	b.mu.Lock()
	b.count++
	b.mu.Unlock()
	if b.started != nil {
		b.started <- struct{}{}
	}
	if b.release != nil {
		<-b.release
	}
	return domain.NewInstance(domain.InstanceOptions{
		ID:         spec.Name + "-inst",
		Spec:       spec,
		SpecKey:    specKey,
		State:      domain.InstanceStateReady,
		LastActive: time.Now(),
	}), nil
}

func (b *blockingLifecycle) StopInstance(_ context.Context, instance *domain.Instance, _ string) error {
	if instance != nil {
		instance.SetState(domain.InstanceStateStopped)
	}
	b.stopMu.Lock()
	b.stopped++
	b.stopMu.Unlock()
	return nil
}

func (b *blockingLifecycle) starts() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.count
}

func (b *blockingLifecycle) stops() int {
	b.stopMu.Lock()
	defer b.stopMu.Unlock()
	return b.stopped
}

type ctxBlockingLifecycle struct {
	started  chan struct{}
	release  chan struct{}
	startCtx context.Context
}

func (c *ctxBlockingLifecycle) StartInstance(ctx context.Context, specKey string, spec domain.ServerSpec) (*domain.Instance, error) {
	c.startCtx = ctx
	if c.started != nil {
		c.started <- struct{}{}
	}
	if c.release != nil {
		<-c.release
	}
	return domain.NewInstance(domain.InstanceOptions{
		ID:         spec.Name + "-inst",
		Spec:       spec,
		SpecKey:    specKey,
		State:      domain.InstanceStateReady,
		LastActive: time.Now(),
	}), nil
}

func (c *ctxBlockingLifecycle) StopInstance(_ context.Context, instance *domain.Instance, _ string) error {
	if instance != nil {
		instance.SetState(domain.InstanceStateStopped)
	}
	return nil
}

type failOnceBlockingLifecycle struct {
	mu      sync.Mutex
	count   int
	started chan struct{}
	release chan struct{}
}

func (f *failOnceBlockingLifecycle) StartInstance(_ context.Context, specKey string, spec domain.ServerSpec) (*domain.Instance, error) {
	f.mu.Lock()
	f.count++
	count := f.count
	f.mu.Unlock()

	if f.started != nil {
		f.started <- struct{}{}
	}
	if f.release != nil {
		<-f.release
	}

	if count == 1 {
		return nil, errors.New("start failed")
	}

	return domain.NewInstance(domain.InstanceOptions{
		ID:         spec.Name + "-inst",
		Spec:       spec,
		SpecKey:    specKey,
		State:      domain.InstanceStateReady,
		LastActive: time.Now(),
	}), nil
}

func (f *failOnceBlockingLifecycle) StopInstance(_ context.Context, instance *domain.Instance, _ string) error {
	if instance != nil {
		instance.SetState(domain.InstanceStateStopped)
	}
	return nil
}

func (f *failOnceBlockingLifecycle) starts() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.count
}

type panicLifecycle struct {
	mu    sync.Mutex
	count int
}

func (p *panicLifecycle) StartInstance(_ context.Context, specKey string, spec domain.ServerSpec) (*domain.Instance, error) {
	p.mu.Lock()
	p.count++
	count := p.count
	p.mu.Unlock()
	if count == 1 {
		panic("boom")
	}
	return domain.NewInstance(domain.InstanceOptions{
		ID:         spec.Name + "-inst",
		Spec:       spec,
		SpecKey:    specKey,
		State:      domain.InstanceStateReady,
		LastActive: time.Now(),
	}), nil
}

func (p *panicLifecycle) StopInstance(_ context.Context, instance *domain.Instance, _ string) error {
	if instance != nil {
		instance.SetState(domain.InstanceStateStopped)
	}
	return nil
}

func (p *panicLifecycle) starts() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.count
}

type trackingLifecycle struct {
	stopCh   chan struct{}
	stopOnce sync.Once
}

func (t *trackingLifecycle) StartInstance(_ context.Context, specKey string, spec domain.ServerSpec) (*domain.Instance, error) {
	return domain.NewInstance(domain.InstanceOptions{
		ID:         spec.Name + "-inst",
		Spec:       spec,
		SpecKey:    specKey,
		State:      domain.InstanceStateReady,
		LastActive: time.Now(),
	}), nil
}

func (t *trackingLifecycle) StopInstance(_ context.Context, instance *domain.Instance, _ string) error {
	if instance != nil {
		instance.SetState(domain.InstanceStateStopped)
	}
	t.stopOnce.Do(func() {
		close(t.stopCh)
	})
	return nil
}

// slowStopLifecycle simulates a slow process exit for testing stop performance.
type slowStopLifecycle struct {
	stopDelay time.Duration
	mu        sync.Mutex
	stopCount int
}

func (s *slowStopLifecycle) StartInstance(_ context.Context, specKey string, spec domain.ServerSpec) (*domain.Instance, error) {
	return domain.NewInstance(domain.InstanceOptions{
		ID:         fmt.Sprintf("%s-inst-%d", spec.Name, time.Now().UnixNano()),
		Spec:       spec,
		SpecKey:    specKey,
		State:      domain.InstanceStateReady,
		LastActive: time.Now(),
	}), nil
}

func (s *slowStopLifecycle) StopInstance(_ context.Context, instance *domain.Instance, _ string) error {
	time.Sleep(s.stopDelay)
	s.mu.Lock()
	s.stopCount++
	s.mu.Unlock()
	if instance != nil {
		instance.SetState(domain.InstanceStateStopped)
	}
	return nil
}

func (s *slowStopLifecycle) stops() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopCount
}
