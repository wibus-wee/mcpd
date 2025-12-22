package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mcpd/internal/domain"
)

var (
	ErrUnknownServerType = errors.New("unknown server type")
	ErrNoCapacity        = errors.New("no capacity available")
	ErrStickyBusy        = errors.New("sticky instance at capacity")
	ErrNotImplemented    = errors.New("scheduler not implemented")
)

type BasicScheduler struct {
	lifecycle domain.Lifecycle
	specs     map[string]domain.ServerSpec

	instances map[string][]*trackedInstance
	sticky    map[string]map[string]*trackedInstance // serverType -> routingKey -> instance
}

type trackedInstance struct {
	instance *domain.Instance
}

func NewBasicScheduler(lifecycle domain.Lifecycle, specs map[string]domain.ServerSpec) *BasicScheduler {
	return &BasicScheduler{
		lifecycle: lifecycle,
		specs:     specs,
		instances: make(map[string][]*trackedInstance),
		sticky:    make(map[string]map[string]*trackedInstance),
	}
}

func (s *BasicScheduler) Acquire(ctx context.Context, serverType, routingKey string) (*domain.Instance, error) {
	spec, ok := s.specs[serverType]
	if !ok {
		return nil, ErrUnknownServerType
	}

	if spec.Sticky && routingKey != "" {
		if inst := s.lookupSticky(serverType, routingKey); inst != nil {
			if inst.instance.BusyCount >= spec.MaxConcurrent {
				return nil, ErrStickyBusy
			}
			return s.markBusy(inst, spec), nil
		}
	}

	if inst := s.findReadyInstance(serverType, spec); inst != nil {
		return s.markBusy(inst, spec), nil
	}

	newInst, err := s.lifecycle.StartInstance(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("start instance: %w", err)
	}
	tracked := &trackedInstance{instance: newInst}
	s.instances[serverType] = append(s.instances[serverType], tracked)

	if spec.Sticky && routingKey != "" {
		s.bindSticky(serverType, routingKey, tracked)
	}

	return s.markBusy(tracked, spec), nil
}

func (s *BasicScheduler) Release(ctx context.Context, instance *domain.Instance) error {
	if instance == nil {
		return errors.New("instance is nil")
	}

	instance.BusyCount--
	if instance.BusyCount < 0 {
		instance.BusyCount = 0
	}
	instance.LastActive = time.Now()
	if instance.BusyCount == 0 {
		instance.State = domain.InstanceStateReady
	}
	return nil
}

func (s *BasicScheduler) lookupSticky(serverType, routingKey string) *trackedInstance {
	if m := s.sticky[serverType]; m != nil {
		return m[routingKey]
	}
	return nil
}

func (s *BasicScheduler) bindSticky(serverType, routingKey string, inst *trackedInstance) {
	if s.sticky[serverType] == nil {
		s.sticky[serverType] = make(map[string]*trackedInstance)
	}
	s.sticky[serverType][routingKey] = inst
}

func (s *BasicScheduler) findReadyInstance(serverType string, spec domain.ServerSpec) *trackedInstance {
	for _, inst := range s.instances[serverType] {
		if inst.instance.BusyCount < spec.MaxConcurrent {
			return inst
		}
	}
	return nil
}

func (s *BasicScheduler) markBusy(inst *trackedInstance, spec domain.ServerSpec) *domain.Instance {
	inst.instance.BusyCount++
	if inst.instance.BusyCount > 0 {
		inst.instance.State = domain.InstanceStateBusy
	}
	inst.instance.LastActive = time.Now()
	return inst.instance
}
