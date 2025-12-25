package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"mcpd/internal/infra/telemetry"

	"mcpd/internal/domain"
)

var (
	ErrUnknownServerType = errors.New("unknown server type")
	ErrNoCapacity        = errors.New("no capacity available")
	ErrStickyBusy        = errors.New("sticky instance at capacity")
	ErrNotImplemented    = errors.New("scheduler not implemented")
)

type SchedulerOptions struct {
	Probe        domain.HealthProbe
	PingInterval time.Duration
	Logger       *zap.Logger
	Metrics      domain.Metrics
	Health       *telemetry.HealthTracker
}

type BasicScheduler struct {
	lifecycle domain.Lifecycle
	specs     map[string]domain.ServerSpec

	mu        sync.Mutex
	instances map[string][]*trackedInstance
	sticky    map[string]map[string]*trackedInstance // serverType -> routingKey -> instance

	probe   domain.HealthProbe
	logger  *zap.Logger
	metrics domain.Metrics
	health  *telemetry.HealthTracker

	idleTicker *time.Ticker
	stopIdle   chan struct{}
	pingTicker *time.Ticker
	stopPing   chan struct{}

	idleBeat *telemetry.Heartbeat
	pingBeat *telemetry.Heartbeat
}

type trackedInstance struct {
	instance *domain.Instance
}

type stopCandidate struct {
	serverType string
	inst       *trackedInstance
	reason     string
}

func NewBasicScheduler(lifecycle domain.Lifecycle, specs map[string]domain.ServerSpec, opts SchedulerOptions) *BasicScheduler {
	logger := opts.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	return &BasicScheduler{
		lifecycle: lifecycle,
		specs:     specs,
		instances: make(map[string][]*trackedInstance),
		sticky:    make(map[string]map[string]*trackedInstance),
		probe:     opts.Probe,
		logger:    logger.Named("scheduler"),
		metrics:   opts.Metrics,
		health:    opts.Health,
		stopIdle:  make(chan struct{}),
		stopPing:  make(chan struct{}),
	}
}

func (s *BasicScheduler) Acquire(ctx context.Context, serverType, routingKey string) (*domain.Instance, error) {
	s.mu.Lock()
	spec, ok := s.specs[serverType]
	if !ok {
		s.mu.Unlock()
		return nil, ErrUnknownServerType
	}

	if spec.Sticky && routingKey != "" {
		if inst := s.lookupStickyLocked(serverType, routingKey); inst != nil {
			if !isRoutable(inst.instance.State) {
				s.unbindStickyLocked(serverType, routingKey)
			} else {
				if inst.instance.BusyCount >= spec.MaxConcurrent {
					s.mu.Unlock()
					return nil, ErrStickyBusy
				}
				instance := s.markBusyLocked(inst)
				s.mu.Unlock()
				return instance, nil
			}
		}
	}

	if inst := s.findReadyInstanceLocked(serverType, spec); inst != nil {
		instance := s.markBusyLocked(inst)
		s.mu.Unlock()
		return instance, nil
	}
	s.mu.Unlock()

	started := time.Now()
	newInst, err := s.lifecycle.StartInstance(ctx, spec)
	s.observeInstanceStart(spec.Name, started, err)
	if err != nil {
		return nil, fmt.Errorf("start instance: %w", err)
	}
	tracked := &trackedInstance{instance: newInst}

	s.mu.Lock()
	s.instances[serverType] = append(s.instances[serverType], tracked)
	if spec.Sticky && routingKey != "" {
		s.bindStickyLocked(serverType, routingKey, tracked)
	}
	instance := s.markBusyLocked(tracked)
	activeCount := len(s.instances[serverType])
	s.mu.Unlock()
	s.observeActiveInstances(serverType, activeCount)

	return instance, nil
}

func (s *BasicScheduler) Release(ctx context.Context, instance *domain.Instance) error {
	if instance == nil {
		return errors.New("instance is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if instance.BusyCount > 0 {
		instance.BusyCount--
	}
	instance.LastActive = time.Now()
	if instance.BusyCount == 0 && instance.State == domain.InstanceStateBusy {
		instance.State = domain.InstanceStateReady
	}
	return nil
}

func (s *BasicScheduler) lookupStickyLocked(serverType, routingKey string) *trackedInstance {
	if m := s.sticky[serverType]; m != nil {
		return m[routingKey]
	}
	return nil
}

func (s *BasicScheduler) bindStickyLocked(serverType, routingKey string, inst *trackedInstance) {
	if s.sticky[serverType] == nil {
		s.sticky[serverType] = make(map[string]*trackedInstance)
	}
	s.sticky[serverType][routingKey] = inst
}

func (s *BasicScheduler) unbindStickyLocked(serverType, routingKey string) {
	if s.sticky[serverType] == nil {
		return
	}
	delete(s.sticky[serverType], routingKey)
	if len(s.sticky[serverType]) == 0 {
		delete(s.sticky, serverType)
	}
}

func (s *BasicScheduler) findReadyInstanceLocked(serverType string, spec domain.ServerSpec) *trackedInstance {
	for _, inst := range s.instances[serverType] {
		if inst.instance.BusyCount >= spec.MaxConcurrent {
			continue
		}
		if !isRoutable(inst.instance.State) {
			continue
		}
		return inst
	}
	return nil
}

func (s *BasicScheduler) markBusyLocked(inst *trackedInstance) *domain.Instance {
	inst.instance.BusyCount++
	inst.instance.State = domain.InstanceStateBusy
	inst.instance.LastActive = time.Now()
	return inst.instance
}

// StartIdleManager begins periodic idle reap respecting idleSeconds/persistent/sticky/minReady.
func (s *BasicScheduler) StartIdleManager(interval time.Duration) {
	if interval <= 0 {
		interval = time.Second
	}
	s.mu.Lock()
	if s.idleTicker != nil {
		s.mu.Unlock()
		return
	}
	if s.stopIdle == nil {
		s.stopIdle = make(chan struct{})
	}
	if s.health != nil && s.idleBeat == nil {
		s.idleBeat = s.health.Register("scheduler.idle", interval*3)
	}
	s.idleTicker = time.NewTicker(interval)
	s.mu.Unlock()

	go func() {
		for {
			select {
			case <-s.idleTicker.C:
				if s.idleBeat != nil {
					s.idleBeat.Beat()
				}
				s.reapIdle()
			case <-s.stopIdle:
				return
			}
		}
	}()
}

func (s *BasicScheduler) StopIdleManager() {
	s.mu.Lock()
	if s.idleTicker != nil {
		s.idleTicker.Stop()
		s.idleTicker = nil
	}
	if s.idleBeat != nil {
		s.idleBeat.Stop()
		s.idleBeat = nil
	}
	if s.stopIdle != nil {
		close(s.stopIdle)
		s.stopIdle = nil
	}
	s.mu.Unlock()
}

func (s *BasicScheduler) StartPingManager(interval time.Duration) {
	if interval <= 0 || s.probe == nil {
		return
	}
	s.mu.Lock()
	if s.pingTicker != nil {
		s.mu.Unlock()
		return
	}
	if s.stopPing == nil {
		s.stopPing = make(chan struct{})
	}
	if s.health != nil && s.pingBeat == nil {
		s.pingBeat = s.health.Register("scheduler.ping", interval*3)
	}
	s.pingTicker = time.NewTicker(interval)
	s.mu.Unlock()

	go func() {
		for {
			select {
			case <-s.pingTicker.C:
				if s.pingBeat != nil {
					s.pingBeat.Beat()
				}
				s.probeInstances()
			case <-s.stopPing:
				return
			}
		}
	}()
}

func (s *BasicScheduler) StopPingManager() {
	s.mu.Lock()
	if s.pingTicker != nil {
		s.pingTicker.Stop()
		s.pingTicker = nil
	}
	if s.pingBeat != nil {
		s.pingBeat.Stop()
		s.pingBeat = nil
	}
	if s.stopPing != nil {
		close(s.stopPing)
		s.stopPing = nil
	}
	s.mu.Unlock()
}

func (s *BasicScheduler) reapIdle() {
	now := time.Now()
	var candidates []stopCandidate

	s.mu.Lock()
	for serverType, list := range s.instances {
		spec := s.specs[serverType]
		readyCount := s.countReadyLocked(serverType)

		for _, inst := range list {
			if inst.instance.State != domain.InstanceStateReady {
				continue
			}
			if spec.Persistent || spec.Sticky {
				continue
			}
			if readyCount <= spec.MinReady {
				continue
			}
			idleFor := now.Sub(inst.instance.LastActive)
			if idleFor >= time.Duration(spec.IdleSeconds)*time.Second {
				inst.instance.State = domain.InstanceStateDraining
				s.logger.Info("idle reap",
					telemetry.EventField(telemetry.EventIdleReap),
					telemetry.ServerTypeField(serverType),
					telemetry.InstanceIDField(inst.instance.ID),
					telemetry.StateField(string(inst.instance.State)),
					telemetry.DurationField(idleFor),
				)
				candidates = append(candidates, stopCandidate{serverType: serverType, inst: inst, reason: "idle timeout"})
				readyCount--
			}
		}
	}
	s.mu.Unlock()

	for _, candidate := range candidates {
		err := s.lifecycle.StopInstance(context.Background(), candidate.inst.instance, candidate.reason)
		s.observeInstanceStop(candidate.serverType, err)
		s.mu.Lock()
		activeCount := s.removeInstanceLocked(candidate.serverType, candidate.inst)
		s.mu.Unlock()
		s.observeActiveInstances(candidate.serverType, activeCount)
	}
}

func (s *BasicScheduler) probeInstances() {
	if s.probe == nil {
		return
	}

	var candidates []stopCandidate
	var checks []stopCandidate

	s.mu.Lock()
	for serverType, list := range s.instances {
		for _, inst := range list {
			if !isRoutable(inst.instance.State) {
				continue
			}
			checks = append(checks, stopCandidate{serverType: serverType, inst: inst, reason: "ping failure"})
		}
	}
	s.mu.Unlock()

	for _, candidate := range checks {
		if err := s.probe.Ping(context.Background(), candidate.inst.instance.Conn); err != nil {
			s.logger.Warn("ping failed",
				telemetry.EventField(telemetry.EventPingFailure),
				telemetry.ServerTypeField(candidate.serverType),
				telemetry.InstanceIDField(candidate.inst.instance.ID),
				telemetry.StateField(string(candidate.inst.instance.State)),
				zap.Error(err),
			)
			candidates = append(candidates, candidate)
		}
	}

	for _, candidate := range candidates {
		s.mu.Lock()
		candidate.inst.instance.State = domain.InstanceStateFailed
		s.mu.Unlock()

		err := s.lifecycle.StopInstance(context.Background(), candidate.inst.instance, candidate.reason)
		s.observeInstanceStop(candidate.serverType, err)
		s.mu.Lock()
		activeCount := s.removeInstanceLocked(candidate.serverType, candidate.inst)
		s.mu.Unlock()
		s.observeActiveInstances(candidate.serverType, activeCount)
	}
}

// StopAll terminates all known instances for graceful shutdown.
func (s *BasicScheduler) StopAll(ctx context.Context) {
	var candidates []stopCandidate

	s.mu.Lock()
	for serverType, list := range s.instances {
		for _, inst := range list {
			candidates = append(candidates, stopCandidate{serverType: serverType, inst: inst, reason: "shutdown"})
		}
	}
	s.mu.Unlock()

	for _, candidate := range candidates {
		err := s.lifecycle.StopInstance(ctx, candidate.inst.instance, candidate.reason)
		s.observeInstanceStop(candidate.serverType, err)
	}

	s.mu.Lock()
	for serverType := range s.instances {
		s.observeActiveInstances(serverType, 0)
	}
	s.instances = make(map[string][]*trackedInstance)
	s.sticky = make(map[string]map[string]*trackedInstance)
	s.mu.Unlock()
}

func (s *BasicScheduler) removeInstanceLocked(serverType string, inst *trackedInstance) int {
	list := s.instances[serverType]
	if len(list) == 0 {
		return 0
	}

	out := list[:0]
	for _, candidate := range list {
		if candidate != inst {
			out = append(out, candidate)
		}
	}
	if len(out) == 0 {
		delete(s.instances, serverType)
	} else {
		s.instances[serverType] = out
	}

	if m := s.sticky[serverType]; m != nil {
		for key, bound := range m {
			if bound == inst {
				delete(m, key)
			}
		}
		if len(m) == 0 {
			delete(s.sticky, serverType)
		}
	}
	if s.instances[serverType] == nil {
		return 0
	}
	return len(s.instances[serverType])
}

func (s *BasicScheduler) countReadyLocked(serverType string) int {
	count := 0
	for _, inst := range s.instances[serverType] {
		if inst.instance.State == domain.InstanceStateReady {
			count++
		}
	}
	return count
}

func (s *BasicScheduler) observeInstanceStart(serverType string, start time.Time, err error) {
	if s.metrics == nil {
		return
	}
	s.metrics.ObserveInstanceStart(serverType, time.Since(start), err)
}

func (s *BasicScheduler) observeInstanceStop(serverType string, err error) {
	if s.metrics == nil {
		return
	}
	s.metrics.ObserveInstanceStop(serverType, err)
}

func (s *BasicScheduler) observeActiveInstances(serverType string, count int) {
	if s.metrics == nil {
		return
	}
	s.metrics.SetActiveInstances(serverType, count)
}

func isRoutable(state domain.InstanceState) bool {
	return state == domain.InstanceStateReady || state == domain.InstanceStateBusy
}
