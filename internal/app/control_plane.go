package app

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"mcpd/internal/domain"
	"mcpd/internal/infra/aggregator"
)

// ControlPlane aggregates control plane services behind a facade.
type ControlPlane struct {
	state         *controlPlaneState
	registry      *callerRegistry
	discovery     *discoveryService
	observability *observabilityService
	automation    *automationService
}

// NewControlPlane constructs a control plane facade from services.
func NewControlPlane(
	state *controlPlaneState,
	registry *callerRegistry,
	discovery *discoveryService,
	observability *observabilityService,
	automation *automationService,
) *ControlPlane {
	return &ControlPlane{
		state:         state,
		registry:      registry,
		discovery:     discovery,
		observability: observability,
		automation:    automation,
	}
}

// StartCallerMonitor begins caller monitoring and profile change handling.
func (c *ControlPlane) StartCallerMonitor(ctx context.Context) {
	c.discovery.StartProfileChangeListener(ctx)
	c.registry.StartMonitor(ctx)
}

// Info returns control plane metadata.
func (c *ControlPlane) Info(ctx context.Context) (domain.ControlPlaneInfo, error) {
	return c.state.info, nil
}

// RegisterCaller registers a caller with the control plane.
func (c *ControlPlane) RegisterCaller(ctx context.Context, caller string, pid int) (string, error) {
	return c.registry.RegisterCaller(ctx, caller, pid)
}

// UnregisterCaller unregisters a caller.
func (c *ControlPlane) UnregisterCaller(ctx context.Context, caller string) error {
	return c.registry.UnregisterCaller(ctx, caller)
}

// ListActiveCallers lists active callers.
func (c *ControlPlane) ListActiveCallers(ctx context.Context) ([]domain.ActiveCaller, error) {
	return c.registry.ListActiveCallers(ctx)
}

// WatchActiveCallers streams active caller updates.
func (c *ControlPlane) WatchActiveCallers(ctx context.Context) (<-chan domain.ActiveCallerSnapshot, error) {
	return c.registry.WatchActiveCallers(ctx)
}

// ListTools lists tools visible to a caller.
func (c *ControlPlane) ListTools(ctx context.Context, caller string) (domain.ToolSnapshot, error) {
	return c.discovery.ListTools(ctx, caller)
}

// ListToolsAllProfiles lists tools across all profiles.
func (c *ControlPlane) ListToolsAllProfiles(ctx context.Context) (domain.ToolSnapshot, error) {
	return c.discovery.ListToolsAllProfiles(ctx)
}

// ListToolCatalog returns the full tool catalog snapshot.
func (c *ControlPlane) ListToolCatalog(ctx context.Context) (domain.ToolCatalogSnapshot, error) {
	return c.discovery.ListToolCatalog(ctx)
}

// WatchTools streams tool snapshots for a caller.
func (c *ControlPlane) WatchTools(ctx context.Context, caller string) (<-chan domain.ToolSnapshot, error) {
	return c.discovery.WatchTools(ctx, caller)
}

// CallTool executes a tool on behalf of a caller.
func (c *ControlPlane) CallTool(ctx context.Context, caller, name string, args json.RawMessage, routingKey string) (json.RawMessage, error) {
	return c.discovery.CallTool(ctx, caller, name, args, routingKey)
}

// CallToolAllProfiles executes a tool across all profiles.
func (c *ControlPlane) CallToolAllProfiles(ctx context.Context, name string, args json.RawMessage, routingKey, specKey string) (json.RawMessage, error) {
	return c.discovery.CallToolAllProfiles(ctx, name, args, routingKey, specKey)
}

// ListResources lists resources visible to a caller.
func (c *ControlPlane) ListResources(ctx context.Context, caller string, cursor string) (domain.ResourcePage, error) {
	return c.discovery.ListResources(ctx, caller, cursor)
}

// ListResourcesAllProfiles lists resources across all profiles.
func (c *ControlPlane) ListResourcesAllProfiles(ctx context.Context, cursor string) (domain.ResourcePage, error) {
	return c.discovery.ListResourcesAllProfiles(ctx, cursor)
}

// WatchResources streams resource snapshots for a caller.
func (c *ControlPlane) WatchResources(ctx context.Context, caller string) (<-chan domain.ResourceSnapshot, error) {
	return c.discovery.WatchResources(ctx, caller)
}

// ReadResource reads a resource on behalf of a caller.
func (c *ControlPlane) ReadResource(ctx context.Context, caller, uri string) (json.RawMessage, error) {
	return c.discovery.ReadResource(ctx, caller, uri)
}

// ReadResourceAllProfiles reads a resource across all profiles.
func (c *ControlPlane) ReadResourceAllProfiles(ctx context.Context, uri, specKey string) (json.RawMessage, error) {
	return c.discovery.ReadResourceAllProfiles(ctx, uri, specKey)
}

// ListPrompts lists prompts visible to a caller.
func (c *ControlPlane) ListPrompts(ctx context.Context, caller string, cursor string) (domain.PromptPage, error) {
	return c.discovery.ListPrompts(ctx, caller, cursor)
}

// ListPromptsAllProfiles lists prompts across all profiles.
func (c *ControlPlane) ListPromptsAllProfiles(ctx context.Context, cursor string) (domain.PromptPage, error) {
	return c.discovery.ListPromptsAllProfiles(ctx, cursor)
}

// WatchPrompts streams prompt snapshots for a caller.
func (c *ControlPlane) WatchPrompts(ctx context.Context, caller string) (<-chan domain.PromptSnapshot, error) {
	return c.discovery.WatchPrompts(ctx, caller)
}

// GetPrompt resolves a prompt for a caller.
func (c *ControlPlane) GetPrompt(ctx context.Context, caller, name string, args json.RawMessage) (json.RawMessage, error) {
	return c.discovery.GetPrompt(ctx, caller, name, args)
}

// GetPromptAllProfiles resolves a prompt across all profiles.
func (c *ControlPlane) GetPromptAllProfiles(ctx context.Context, name string, args json.RawMessage, specKey string) (json.RawMessage, error) {
	return c.discovery.GetPromptAllProfiles(ctx, name, args, specKey)
}

// StreamLogs streams logs for a caller.
func (c *ControlPlane) StreamLogs(ctx context.Context, caller string, minLevel domain.LogLevel) (<-chan domain.LogEntry, error) {
	return c.observability.StreamLogs(ctx, caller, minLevel)
}

// StreamLogsAllProfiles streams logs across all profiles.
func (c *ControlPlane) StreamLogsAllProfiles(ctx context.Context, minLevel domain.LogLevel) (<-chan domain.LogEntry, error) {
	return c.observability.StreamLogsAllProfiles(ctx, minLevel)
}

// GetProfileStore returns the profile store.
func (c *ControlPlane) GetProfileStore() domain.ProfileStore {
	return c.state.ProfileStore()
}

// GetPoolStatus returns the current pool status snapshot.
func (c *ControlPlane) GetPoolStatus(ctx context.Context) ([]domain.PoolInfo, error) {
	return c.observability.GetPoolStatus(ctx)
}

// GetServerInitStatus returns current server init statuses.
func (c *ControlPlane) GetServerInitStatus(ctx context.Context) ([]domain.ServerInitStatus, error) {
	return c.observability.GetServerInitStatus(ctx)
}

// RetryServerInit requests a retry for server initialization.
func (c *ControlPlane) RetryServerInit(ctx context.Context, specKey string) error {
	if c.state.initManager == nil {
		return errors.New("server init manager not configured")
	}
	return c.state.initManager.RetrySpec(specKey)
}

// WatchRuntimeStatus streams runtime status snapshots for a caller.
func (c *ControlPlane) WatchRuntimeStatus(ctx context.Context, caller string) (<-chan domain.RuntimeStatusSnapshot, error) {
	return c.observability.WatchRuntimeStatus(ctx, caller)
}

// WatchRuntimeStatusAllProfiles streams runtime status snapshots across profiles.
func (c *ControlPlane) WatchRuntimeStatusAllProfiles(ctx context.Context) (<-chan domain.RuntimeStatusSnapshot, error) {
	return c.observability.WatchRuntimeStatusAllProfiles(ctx)
}

// WatchServerInitStatus streams server init status snapshots for a caller.
func (c *ControlPlane) WatchServerInitStatus(ctx context.Context, caller string) (<-chan domain.ServerInitStatusSnapshot, error) {
	return c.observability.WatchServerInitStatus(ctx, caller)
}

// WatchServerInitStatusAllProfiles streams server init status snapshots across profiles.
func (c *ControlPlane) WatchServerInitStatusAllProfiles(ctx context.Context) (<-chan domain.ServerInitStatusSnapshot, error) {
	return c.observability.WatchServerInitStatusAllProfiles(ctx)
}

// SetRuntimeStatusIndex updates the runtime status index.
func (c *ControlPlane) SetRuntimeStatusIndex(idx *aggregator.RuntimeStatusIndex) {
	c.observability.SetRuntimeStatusIndex(idx)
}

// SetServerInitIndex updates the server init status index.
func (c *ControlPlane) SetServerInitIndex(idx *aggregator.ServerInitIndex) {
	c.observability.SetServerInitIndex(idx)
}

// SetSubAgent sets the active SubAgent implementation.
func (c *ControlPlane) SetSubAgent(agent domain.SubAgent) {
	c.automation.SetSubAgent(agent)
}

// IsSubAgentEnabledForCaller reports whether SubAgent is enabled for a caller.
func (c *ControlPlane) IsSubAgentEnabledForCaller(caller string) bool {
	return c.automation.IsSubAgentEnabledForCaller(caller)
}

// IsSubAgentEnabled reports whether SubAgent is enabled.
func (c *ControlPlane) IsSubAgentEnabled() bool {
	return c.automation.IsSubAgentEnabled()
}

// GetToolSnapshotForCaller returns the tool snapshot for a caller.
func (c *ControlPlane) GetToolSnapshotForCaller(caller string) (domain.ToolSnapshot, error) {
	return c.discovery.GetToolSnapshotForCaller(caller)
}

// AutomaticMCP filters tools using the automatic MCP flow.
func (c *ControlPlane) AutomaticMCP(ctx context.Context, caller string, params domain.AutomaticMCPParams) (domain.AutomaticMCPResult, error) {
	return c.automation.AutomaticMCP(ctx, caller, params)
}

// AutomaticEval evaluates a tool call using the automatic MCP flow.
func (c *ControlPlane) AutomaticEval(ctx context.Context, caller string, params domain.AutomaticEvalParams) (json.RawMessage, error) {
	return c.automation.AutomaticEval(ctx, caller, params)
}

// GetBootstrapProgress returns bootstrap progress.
func (c *ControlPlane) GetBootstrapProgress(ctx context.Context) (domain.BootstrapProgress, error) {
	if c.state.bootstrapManager == nil {
		return domain.BootstrapProgress{State: domain.BootstrapCompleted}, nil
	}
	return c.state.bootstrapManager.GetProgress(), nil
}

// WatchBootstrapProgress streams bootstrap progress updates.
func (c *ControlPlane) WatchBootstrapProgress(ctx context.Context) (<-chan domain.BootstrapProgress, error) {
	ch := make(chan domain.BootstrapProgress, 1)

	if c.state.bootstrapManager == nil {
		close(ch)
		return ch, nil
	}

	go func() {
		defer close(ch)

		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		// Send initial progress
		select {
		case ch <- c.state.bootstrapManager.GetProgress():
		case <-ctx.Done():
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				progress := c.state.bootstrapManager.GetProgress()

				select {
				case ch <- progress:
				default:
				}

				// Stop watching if bootstrap is done
				if progress.State == domain.BootstrapCompleted || progress.State == domain.BootstrapFailed {
					return
				}
			}
		}
	}()

	return ch, nil
}
