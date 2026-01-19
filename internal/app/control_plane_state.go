package app

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"mcpd/internal/domain"
	"mcpd/internal/infra/aggregator"
)

type controlPlaneState struct {
	mu sync.RWMutex

	info             domain.ControlPlaneInfo
	profiles         map[string]*profileRuntime
	callers          map[string]string
	specRegistry     map[string]domain.ServerSpec
	scheduler        domain.Scheduler
	initManager      *ServerInitializationManager
	bootstrapManager *BootstrapManager
	runtime          domain.RuntimeConfig
	logger           *zap.Logger
	ctx              context.Context
	profileStore     domain.ProfileStore
}

func newControlPlaneState(
	ctx context.Context,
	profiles map[string]*profileRuntime,
	scheduler domain.Scheduler,
	initManager *ServerInitializationManager,
	bootstrapManager *BootstrapManager,
	state *domain.CatalogState,
	logger *zap.Logger,
) *controlPlaneState {
	if logger == nil {
		logger = zap.NewNop()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	store := state.Store
	summary := state.Summary
	callers := copyCallers(store.Callers)
	if callers == nil {
		callers = map[string]string{}
	}
	specRegistry := copySpecRegistryMap(summary.SpecRegistry)

	return &controlPlaneState{
		info:             defaultControlPlaneInfo(),
		profiles:         profiles,
		callers:          callers,
		specRegistry:     specRegistry,
		scheduler:        scheduler,
		initManager:      initManager,
		bootstrapManager: bootstrapManager,
		runtime:          summary.DefaultRuntime,
		profileStore:     store,
		logger:           logger.Named("control_plane"),
		ctx:              ctx,
	}
}

type callerState struct {
	pid           int
	profile       string
	lastHeartbeat time.Time
}

type profileRuntime struct {
	name      string
	specKeys  []string
	tools     *aggregator.ToolIndex
	resources *aggregator.ResourceIndex
	prompts   *aggregator.PromptIndex

	mu     sync.RWMutex
	active bool
}

// Activate starts indexes for the profile runtime.
func (p *profileRuntime) Activate(ctx context.Context) {
	p.mu.Lock()
	if p.active {
		p.mu.Unlock()
		return
	}
	p.active = true
	p.mu.Unlock()

	if p.tools != nil {
		p.tools.Start(ctx)
	}
	if p.resources != nil {
		p.resources.Start(ctx)
	}
	if p.prompts != nil {
		p.prompts.Start(ctx)
	}
}

// Deactivate stops indexes for the profile runtime.
func (p *profileRuntime) Deactivate() {
	p.mu.Lock()
	if !p.active {
		p.mu.Unlock()
		return
	}
	p.active = false
	p.mu.Unlock()

	if p.tools != nil {
		p.tools.Stop()
	}
	if p.resources != nil {
		p.resources.Stop()
	}
	if p.prompts != nil {
		p.prompts.Stop()
	}
}

// SpecKeys returns a copy of the profile spec keys.
func (p *profileRuntime) SpecKeys() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.specKeys) == 0 {
		return nil
	}
	return append([]string(nil), p.specKeys...)
}

// UpdateCatalog refreshes profile runtime state from the catalog.
func (p *profileRuntime) UpdateCatalog(cfg domain.CatalogProfile) {
	p.mu.Lock()
	p.specKeys = collectSpecKeys(cfg.SpecKeys)
	p.mu.Unlock()

	if p.tools != nil {
		p.tools.UpdateSpecs(cfg.Profile.Catalog.Specs, cfg.SpecKeys, cfg.Profile.Catalog.Runtime)
	}
	if p.resources != nil {
		p.resources.UpdateSpecs(cfg.Profile.Catalog.Specs, cfg.SpecKeys, cfg.Profile.Catalog.Runtime)
	}
	if p.prompts != nil {
		p.prompts.UpdateSpecs(cfg.Profile.Catalog.Specs, cfg.SpecKeys, cfg.Profile.Catalog.Runtime)
	}
}

func defaultControlPlaneInfo() domain.ControlPlaneInfo {
	return domain.ControlPlaneInfo{
		Name:    "mcpd",
		Version: Version,
		Build:   Build,
	}
}

// UpdateCatalog replaces the control plane state with a new catalog.
func (s *controlPlaneState) UpdateCatalog(state *domain.CatalogState, profiles map[string]*profileRuntime) {
	store := state.Store
	callers := copyCallers(store.Callers)
	if callers == nil {
		callers = map[string]string{}
	}
	specRegistry := copySpecRegistryMap(state.Summary.SpecRegistry)

	s.mu.Lock()
	s.profileStore = store
	s.callers = callers
	s.specRegistry = specRegistry
	s.runtime = state.Summary.DefaultRuntime
	s.profiles = profiles
	s.mu.Unlock()
}

// ProfileStore returns the current profile store.
func (s *controlPlaneState) ProfileStore() domain.ProfileStore {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profileStore
}

// Callers returns a copy of the current caller mapping.
func (s *controlPlaneState) Callers() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return copyCallers(s.callers)
}

// Profiles returns a copy of the current profile runtimes.
func (s *controlPlaneState) Profiles() map[string]*profileRuntime {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return copyProfiles(s.profiles)
}

// Profile returns a profile runtime by name.
func (s *controlPlaneState) Profile(name string) (*profileRuntime, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	runtime, ok := s.profiles[name]
	return runtime, ok
}

// SpecRegistry returns a copy of the current spec registry.
func (s *controlPlaneState) SpecRegistry() map[string]domain.ServerSpec {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return copySpecRegistryMap(s.specRegistry)
}

// Runtime returns the current runtime config.
func (s *controlPlaneState) Runtime() domain.RuntimeConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.runtime
}

func copyCallers(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func copyProfiles(src map[string]*profileRuntime) map[string]*profileRuntime {
	if src == nil {
		return nil
	}
	dst := make(map[string]*profileRuntime, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func copySpecRegistryMap(src map[string]domain.ServerSpec) map[string]domain.ServerSpec {
	if src == nil {
		return nil
	}
	dst := make(map[string]domain.ServerSpec, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}
