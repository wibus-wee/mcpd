package discovery

import (
	"context"
	"encoding/json"
	"sort"

	"mcpv/internal/app/controlplane/registry"
	"mcpv/internal/app/runtime"
	"mcpv/internal/domain"
	"mcpv/internal/infra/hashutil"
)

type PromptDiscoveryService struct {
	*Service[domain.PromptSnapshot]
}

func NewPromptDiscoveryService(state State, registry *registry.ClientRegistry) *PromptDiscoveryService {
	service := &PromptDiscoveryService{}
	base := NewDiscoveryService(state, registry, Options[domain.PromptSnapshot]{
		GetIndex: func(rt *runtime.State) snapshotIndex[domain.PromptSnapshot] { return rt.Prompts() },
	})
	service.Service = base
	base.filterSnapshot = service.filterPromptSnapshot
	return service
}

// ListPrompts lists prompts visible to a client.
func (d *PromptDiscoveryService) ListPrompts(ctx context.Context, client string, cursor string) (domain.PromptPage, error) {
	snapshot, err := d.ListSnapshot(ctx, client)
	if err != nil {
		return domain.PromptPage{}, err
	}
	return paginatePrompts(snapshot, cursor)
}

// ListPromptsAll lists prompts across all servers.
func (d *PromptDiscoveryService) ListPromptsAll(ctx context.Context, cursor string) (domain.PromptPage, error) {
	snapshot, err := d.ListSnapshotAll(ctx)
	if err != nil {
		return domain.PromptPage{}, err
	}
	return paginatePrompts(snapshot, cursor)
}

// WatchPrompts streams prompt snapshots for a client.
func (d *PromptDiscoveryService) WatchPrompts(ctx context.Context, client string) (<-chan domain.PromptSnapshot, error) {
	return d.WatchSnapshots(ctx, client)
}

// GetPrompt resolves a prompt for a client.
func (d *PromptDiscoveryService) GetPrompt(ctx context.Context, client, name string, args json.RawMessage) (json.RawMessage, error) {
	serverName, err := d.resolveClientServer(client)
	if err != nil {
		return nil, err
	}
	runtime := d.state.RuntimeState()
	if runtime == nil || runtime.Prompts() == nil {
		return nil, domain.ErrPromptNotFound
	}
	if serverName != "" {
		if _, ok := runtime.Prompts().ResolveForServer(serverName, name); !ok {
			return nil, domain.ErrPromptNotFound
		}
		ctx = domain.WithRouteContext(ctx, domain.RouteContext{Client: client})
		return runtime.Prompts().GetPromptForServer(ctx, serverName, name, args)
	}
	visibleSpecKeys, err := d.resolveVisibleSpecKeys(client)
	if err != nil {
		return nil, err
	}
	target, ok := runtime.Prompts().Resolve(name)
	if !ok {
		return nil, domain.ErrPromptNotFound
	}
	visibleSpecSet := toSpecKeySet(visibleSpecKeys)
	if target.SpecKey != "" {
		if _, ok := visibleSpecSet[target.SpecKey]; !ok {
			return nil, domain.ErrPromptNotFound
		}
	} else if !d.isServerVisible(visibleSpecSet, target.ServerType) {
		return nil, domain.ErrPromptNotFound
	}
	ctx = domain.WithRouteContext(ctx, domain.RouteContext{Client: client})
	return runtime.Prompts().GetPrompt(ctx, name, args)
}

// GetPromptAll resolves a prompt without client visibility checks.
func (d *PromptDiscoveryService) GetPromptAll(ctx context.Context, name string, args json.RawMessage) (json.RawMessage, error) {
	runtime := d.state.RuntimeState()
	if runtime == nil || runtime.Prompts() == nil {
		return nil, domain.ErrPromptNotFound
	}
	if _, ok := runtime.Prompts().Resolve(name); !ok {
		return nil, domain.ErrPromptNotFound
	}
	ctx = domain.WithRouteContext(ctx, domain.RouteContext{Client: domain.InternalUIClientName})
	return runtime.Prompts().GetPrompt(ctx, name, args)
}

func (d *PromptDiscoveryService) filterPromptSnapshot(snapshot domain.PromptSnapshot, visibleSpecKeys []string) domain.PromptSnapshot {
	if len(snapshot.Prompts) == 0 {
		return domain.PromptSnapshot{}
	}
	visibleServers, visibleSpecSet := d.visibleServers(visibleSpecKeys)
	filtered := make([]domain.PromptDefinition, 0, len(snapshot.Prompts))
	for _, prompt := range snapshot.Prompts {
		if prompt.ServerName != "" {
			if _, ok := visibleServers[prompt.ServerName]; !ok {
				continue
			}
		} else if prompt.SpecKey != "" {
			if _, ok := visibleSpecSet[prompt.SpecKey]; !ok {
				continue
			}
		}
		filtered = append(filtered, prompt)
	}
	if len(filtered) == 0 {
		return domain.PromptSnapshot{}
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Name < filtered[j].Name
	})
	return domain.PromptSnapshot{
		ETag:    hashutil.PromptETag(d.state.Logger(), filtered),
		Prompts: filtered,
	}
}
