package discovery

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"go.uber.org/zap"

	"mcpv/internal/app/controlplane/registry"
	"mcpv/internal/app/runtime"
	"mcpv/internal/domain"
	"mcpv/internal/infra/hashutil"
)

type ToolDiscoveryService struct {
	*Service[domain.ToolSnapshot]
}

func NewToolDiscoveryService(state State, registry *registry.ClientRegistry) *ToolDiscoveryService {
	service := &ToolDiscoveryService{}
	base := NewDiscoveryService(state, registry, Options[domain.ToolSnapshot]{
		GetIndex: func(rt *runtime.State) snapshotIndex[domain.ToolSnapshot] { return rt.Tools() },
	})
	service.Service = base
	base.filterSnapshot = service.filterToolSnapshot
	base.serverSnapshotForList = service.serverSnapshotForList
	return service
}

// ListTools lists tools visible to a client.
func (d *ToolDiscoveryService) ListTools(ctx context.Context, client string) (domain.ToolSnapshot, error) {
	return d.ListSnapshot(ctx, client)
}

// ListToolCatalog returns the full tool catalog snapshot.
func (d *ToolDiscoveryService) ListToolCatalog(_ context.Context) (domain.ToolCatalogSnapshot, error) {
	runtime := d.state.RuntimeState()
	if runtime == nil || runtime.Tools() == nil {
		return domain.ToolCatalogSnapshot{}, nil
	}
	live := runtime.Tools().Snapshot().Tools
	cached := runtime.Tools().CachedSnapshot().Tools

	cachedAt := make(map[string]time.Time)
	if len(cached) > 0 {
		cache := d.metadataCache()
		if cache != nil {
			for _, tool := range cached {
				specKey := tool.SpecKey
				if specKey == "" {
					continue
				}
				if _, ok := cachedAt[specKey]; ok {
					continue
				}
				if ts, ok := cache.GetCachedAt(specKey); ok {
					cachedAt[specKey] = ts
				}
			}
		}
	}

	return buildToolCatalogSnapshot(d.state.Logger(), live, cached, cachedAt), nil
}

// WatchTools streams tool snapshots for a client.
func (d *ToolDiscoveryService) WatchTools(ctx context.Context, client string) (<-chan domain.ToolSnapshot, error) {
	return d.WatchSnapshots(ctx, client)
}

// CallTool executes a tool on behalf of a client.
func (d *ToolDiscoveryService) CallTool(ctx context.Context, client, name string, args json.RawMessage, routingKey string) (json.RawMessage, error) {
	serverName, err := d.resolveClientServer(client)
	if err != nil {
		return nil, err
	}
	runtime := d.state.RuntimeState()
	if runtime == nil || runtime.Tools() == nil {
		return nil, domain.ErrToolNotFound
	}
	if serverName != "" {
		if _, ok := runtime.Tools().ResolveForServer(serverName, name); !ok {
			return nil, domain.ErrToolNotFound
		}
		ctx = domain.WithRouteContext(ctx, domain.RouteContext{Client: client})
		ctx = domain.WithStartCause(ctx, domain.StartCause{
			Reason:   domain.StartCauseToolCall,
			Client:   client,
			ToolName: name,
		})
		return runtime.Tools().CallToolForServer(ctx, serverName, name, args, routingKey)
	}
	visibleSpecKeys, err := d.resolveVisibleSpecKeys(client)
	if err != nil {
		return nil, err
	}
	target, ok := runtime.Tools().Resolve(name)
	if !ok {
		return nil, domain.ErrToolNotFound
	}
	visibleSpecSet := toSpecKeySet(visibleSpecKeys)
	if target.SpecKey != "" {
		if _, ok := visibleSpecSet[target.SpecKey]; !ok {
			return nil, domain.ErrToolNotFound
		}
	} else if !d.isServerVisible(visibleSpecSet, target.ServerType) {
		return nil, domain.ErrToolNotFound
	}
	ctx = domain.WithRouteContext(ctx, domain.RouteContext{Client: client})
	ctx = domain.WithStartCause(ctx, domain.StartCause{
		Reason:   domain.StartCauseToolCall,
		Client:   client,
		ToolName: name,
	})
	return runtime.Tools().CallTool(ctx, name, args, routingKey)
}

// CallToolAll executes a tool without client visibility checks.
func (d *ToolDiscoveryService) CallToolAll(ctx context.Context, name string, args json.RawMessage, routingKey string) (json.RawMessage, error) {
	runtime := d.state.RuntimeState()
	if runtime == nil || runtime.Tools() == nil {
		return nil, domain.ErrToolNotFound
	}
	if _, ok := runtime.Tools().Resolve(name); !ok {
		return nil, domain.ErrToolNotFound
	}
	ctx = domain.WithRouteContext(ctx, domain.RouteContext{Client: domain.InternalUIClientName})
	ctx = domain.WithStartCause(ctx, domain.StartCause{
		Reason:   domain.StartCauseToolCall,
		Client:   domain.InternalUIClientName,
		ToolName: name,
	})
	return runtime.Tools().CallTool(ctx, name, args, routingKey)
}

// GetToolSnapshotForClient returns the tool snapshot for a client.
func (d *ToolDiscoveryService) GetToolSnapshotForClient(client string) (domain.ToolSnapshot, error) {
	return d.ListTools(context.Background(), client)
}

func (d *ToolDiscoveryService) serverSnapshotForList(serverName string, index snapshotIndex[domain.ToolSnapshot]) (domain.ToolSnapshot, bool) {
	snapshot, ok := index.SnapshotForServer(serverName)
	if ok && len(snapshot.Tools) > 0 {
		return snapshot, true
	}
	cached := d.cachedToolSnapshotForServer(serverName)
	if len(cached.Tools) > 0 {
		return cached, true
	}
	if !ok {
		return domain.ToolSnapshot{}, false
	}
	return snapshot, true
}

func (d *ToolDiscoveryService) filterToolSnapshot(snapshot domain.ToolSnapshot, visibleSpecKeys []string) domain.ToolSnapshot {
	if len(snapshot.Tools) == 0 {
		return domain.ToolSnapshot{}
	}
	visibleServers, visibleSpecSet := d.visibleServers(visibleSpecKeys)
	filtered := make([]domain.ToolDefinition, 0, len(snapshot.Tools))
	for _, tool := range snapshot.Tools {
		if tool.ServerName != "" {
			if _, ok := visibleServers[tool.ServerName]; !ok {
				continue
			}
		} else if tool.SpecKey != "" {
			if _, ok := visibleSpecSet[tool.SpecKey]; !ok {
				continue
			}
		}
		filtered = append(filtered, tool)
	}
	if len(filtered) == 0 {
		return domain.ToolSnapshot{}
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].SpecKey != filtered[j].SpecKey {
			return filtered[i].SpecKey < filtered[j].SpecKey
		}
		if filtered[i].Name != filtered[j].Name {
			return filtered[i].Name < filtered[j].Name
		}
		return filtered[i].ServerName < filtered[j].ServerName
	})
	return domain.ToolSnapshot{
		ETag:  hashutil.ToolETag(d.state.Logger(), filtered),
		Tools: filtered,
	}
}

func (d *ToolDiscoveryService) cachedToolSnapshotForServer(serverName string) domain.ToolSnapshot {
	if serverName == "" {
		return domain.ToolSnapshot{}
	}
	cache := d.metadataCache()
	if cache == nil {
		return domain.ToolSnapshot{}
	}
	serverSpecKeys := d.state.ServerSpecKeys()
	specKey := serverSpecKeys[serverName]
	if specKey == "" {
		return domain.ToolSnapshot{}
	}
	tools, ok := cache.GetTools(specKey)
	if !ok || len(tools) == 0 {
		return domain.ToolSnapshot{}
	}

	filtered := make([]domain.ToolDefinition, 0, len(tools))
	for _, tool := range tools {
		if tool.Name == "" {
			continue
		}
		toolDef := tool
		if toolDef.SpecKey == "" {
			toolDef.SpecKey = specKey
		}
		if toolDef.ServerName == "" {
			toolDef.ServerName = serverName
		}
		filtered = append(filtered, toolDef)
	}
	if len(filtered) == 0 {
		return domain.ToolSnapshot{}
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Name < filtered[j].Name
	})
	return domain.ToolSnapshot{
		ETag:  hashutil.ToolETag(d.state.Logger(), filtered),
		Tools: filtered,
	}
}

func buildToolCatalogSnapshot(logger *zap.Logger, liveTools []domain.ToolDefinition, cachedTools []domain.ToolDefinition, cachedAt map[string]time.Time) domain.ToolCatalogSnapshot {
	entries := make(map[string]domain.ToolCatalogEntry)
	for _, tool := range cachedTools {
		entries[toolCatalogKey(tool)] = buildToolCatalogEntry(tool, domain.ToolSourceCache, cachedAt)
	}
	for _, tool := range liveTools {
		entries[toolCatalogKey(tool)] = buildToolCatalogEntry(tool, domain.ToolSourceLive, cachedAt)
	}

	list := make([]domain.ToolCatalogEntry, 0, len(entries))
	for _, entry := range entries {
		list = append(list, entry)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Definition.SpecKey != list[j].Definition.SpecKey {
			return list[i].Definition.SpecKey < list[j].Definition.SpecKey
		}
		if list[i].Definition.Name != list[j].Definition.Name {
			return list[i].Definition.Name < list[j].Definition.Name
		}
		return list[i].Definition.ServerName < list[j].Definition.ServerName
	})

	return domain.ToolCatalogSnapshot{
		Tools: list,
		ETag:  hashutil.ToolCatalogETag(logger, list),
	}
}

func buildToolCatalogEntry(tool domain.ToolDefinition, source domain.ToolSource, cachedAt map[string]time.Time) domain.ToolCatalogEntry {
	entry := domain.ToolCatalogEntry{
		Definition: tool,
		Source:     source,
	}
	if ts, ok := cachedAt[tool.SpecKey]; ok {
		entry.CachedAt = ts
	}
	return entry
}

func toolCatalogKey(tool domain.ToolDefinition) string {
	name := tool.Name
	specKey := tool.SpecKey
	if specKey == "" {
		specKey = tool.ServerName
	}
	return specKey + "\x00" + name
}
