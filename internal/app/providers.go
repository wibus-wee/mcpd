package app

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"mcpd/internal/domain"
	"mcpd/internal/infra/aggregator"
	"mcpd/internal/infra/lifecycle"
	"mcpd/internal/infra/notifications"
	"mcpd/internal/infra/probe"
	"mcpd/internal/infra/router"
	"mcpd/internal/infra/rpc"
	"mcpd/internal/infra/scheduler"
	"mcpd/internal/infra/telemetry"
	"mcpd/internal/infra/transport"
)

// NewMetricsRegistry creates a Prometheus registry.
func NewMetricsRegistry() *prometheus.Registry {
	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	registry.MustRegister(prometheus.NewGoCollector())
	return registry
}

// NewMetrics constructs metrics backed by Prometheus.
func NewMetrics(registry *prometheus.Registry) domain.Metrics {
	return telemetry.NewPrometheusMetrics(registry)
}

// NewHealthTracker constructs a health tracker.
func NewHealthTracker() *telemetry.HealthTracker {
	return telemetry.NewHealthTracker()
}

// NewListChangeHub constructs a list change hub.
func NewListChangeHub() *notifications.ListChangeHub {
	return notifications.NewListChangeHub()
}

// NewCommandLauncher constructs a launcher for stdio servers.
func NewCommandLauncher(logger *zap.Logger) domain.Launcher {
	return transport.NewCommandLauncher(transport.CommandLauncherOptions{Logger: logger})
}

// NewMCPTransport constructs an MCP transport for stdio servers.
func NewMCPTransport(logger *zap.Logger, listChanges *notifications.ListChangeHub) domain.Transport {
	stdioTransport := transport.NewMCPTransport(transport.MCPTransportOptions{
		Logger:            logger,
		ListChangeEmitter: listChanges,
	})
	httpTransport := transport.NewStreamableHTTPTransport(transport.StreamableHTTPTransportOptions{
		Logger:            logger,
		ListChangeEmitter: listChanges,
	})
	return transport.NewCompositeTransport(transport.CompositeTransportOptions{
		Stdio:          stdioTransport,
		StreamableHTTP: httpTransport,
	})
}

// NewLifecycleManager constructs the lifecycle manager.
func NewLifecycleManager(ctx context.Context, launcher domain.Launcher, transport domain.Transport, logger *zap.Logger) domain.Lifecycle {
	return lifecycle.NewManager(ctx, launcher, transport, logger)
}

// NewPingProbe constructs a ping-based health probe.
func NewPingProbe() *probe.PingProbe {
	return &probe.PingProbe{Timeout: defaultPingProbeTimeout}
}

// NewScheduler constructs the scheduler.
func NewScheduler(
	lifecycle domain.Lifecycle,
	state *domain.CatalogState,
	pingProbe *probe.PingProbe,
	metrics domain.Metrics,
	health *telemetry.HealthTracker,
	logger *zap.Logger,
) (domain.Scheduler, error) {
	summary := state.Summary
	return scheduler.NewBasicScheduler(lifecycle, summary.SpecRegistry, scheduler.SchedulerOptions{
		Probe:   pingProbe,
		Logger:  logger,
		Metrics: metrics,
		Health:  health,
	})
}

// NewBootstrapManagerProvider constructs the bootstrap manager.
func NewBootstrapManagerProvider(
	lifecycle domain.Lifecycle,
	scheduler domain.Scheduler,
	state *domain.CatalogState,
	cache *domain.MetadataCache,
	logger *zap.Logger,
) *BootstrapManager {
	summary := state.Summary
	runtime := summary.DefaultRuntime

	mode := runtime.BootstrapMode
	if mode == "" {
		mode = domain.DefaultBootstrapMode
	}
	if mode == domain.BootstrapModeDisabled {
		return nil
	}

	concurrency := runtime.BootstrapConcurrency
	if concurrency <= 0 {
		concurrency = domain.DefaultBootstrapConcurrency
	}

	timeout := time.Duration(runtime.BootstrapTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = time.Duration(domain.DefaultBootstrapTimeoutSeconds) * time.Second
	}

	// Collect all specKeys from all profiles
	allSpecKeys := make(map[string]string)
	for _, profile := range summary.Profiles {
		for k, v := range profile.SpecKeys {
			allSpecKeys[k] = v
		}
	}

	return NewBootstrapManager(BootstrapManagerOptions{
		Scheduler:   scheduler,
		Lifecycle:   lifecycle,
		Specs:       summary.SpecRegistry,
		SpecKeys:    allSpecKeys,
		Runtime:     runtime,
		Cache:       cache,
		Logger:      logger,
		Concurrency: concurrency,
		Timeout:     timeout,
		Mode:        mode,
	})
}

// NewProfileRuntimes constructs runtime state for profiles.
func NewProfileRuntimes(
	state *domain.CatalogState,
	scheduler domain.Scheduler,
	metrics domain.Metrics,
	health *telemetry.HealthTracker,
	metadataCache *domain.MetadataCache,
	listChanges *notifications.ListChangeHub,
	logger *zap.Logger,
) map[string]*profileRuntime {
	summary := state.Summary
	profiles := make(map[string]*profileRuntime, len(summary.Profiles))
	for name, cfg := range summary.Profiles {
		profiles[name] = buildProfileRuntime(name, cfg, scheduler, metrics, health, metadataCache, listChanges, logger)
	}
	return profiles
}

// NewControlPlaneState constructs a control plane state container.
func NewControlPlaneState(
	ctx context.Context,
	profiles map[string]*profileRuntime,
	state *domain.CatalogState,
	scheduler domain.Scheduler,
	initManager *ServerInitializationManager,
	bootstrapManager *BootstrapManager,
	logger *zap.Logger,
) *controlPlaneState {
	controlState := newControlPlaneState(ctx, profiles, scheduler, initManager, bootstrapManager, state, logger)

	// Configure bootstrap waiters for all profile indexes
	if bootstrapManager != nil {
		waiter := func(ctx context.Context) error {
			return bootstrapManager.WaitForCompletion(ctx)
		}
		for _, profile := range profiles {
			if profile.tools != nil {
				profile.tools.SetBootstrapWaiter(waiter)
			}
			if profile.resources != nil {
				profile.resources.SetBootstrapWaiter(waiter)
			}
			if profile.prompts != nil {
				profile.prompts.SetBootstrapWaiter(waiter)
			}
		}
	}

	return controlState
}

// NewRPCServer constructs the RPC server.
func NewRPCServer(control domain.ControlPlane, state *domain.CatalogState, logger *zap.Logger) *rpc.Server {
	return rpc.NewServer(control, state.Summary.DefaultRuntime.RPC, logger)
}

func buildProfileRuntime(
	name string,
	cfg domain.CatalogProfile,
	scheduler domain.Scheduler,
	metrics domain.Metrics,
	health *telemetry.HealthTracker,
	metadataCache *domain.MetadataCache,
	listChanges *notifications.ListChangeHub,
	logger *zap.Logger,
) *profileRuntime {
	profileLogger := logger.With(zap.String("profile", name))
	refreshGate := aggregator.NewRefreshGate()
	baseRouter := router.NewBasicRouter(scheduler, router.RouterOptions{
		Timeout: time.Duration(cfg.Profile.Catalog.Runtime.RouteTimeoutSeconds) * time.Second,
		Logger:  profileLogger,
	})
	rt := router.NewMetricRouter(baseRouter, metrics)
	toolIndex := aggregator.NewToolIndex(rt, cfg.Profile.Catalog.Specs, cfg.SpecKeys, cfg.Profile.Catalog.Runtime, metadataCache, profileLogger, health, refreshGate, listChanges)
	resourceIndex := aggregator.NewResourceIndex(rt, cfg.Profile.Catalog.Specs, cfg.SpecKeys, cfg.Profile.Catalog.Runtime, metadataCache, profileLogger, health, refreshGate, listChanges)
	promptIndex := aggregator.NewPromptIndex(rt, cfg.Profile.Catalog.Specs, cfg.SpecKeys, cfg.Profile.Catalog.Runtime, metadataCache, profileLogger, health, refreshGate, listChanges)
	return &profileRuntime{
		name:      name,
		specKeys:  collectSpecKeys(cfg.SpecKeys),
		tools:     toolIndex,
		resources: resourceIndex,
		prompts:   promptIndex,
	}
}
