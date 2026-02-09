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

type ResourceDiscoveryService struct {
	*Service[domain.ResourceSnapshot]
}

func NewResourceDiscoveryService(state State, registry *registry.ClientRegistry) *ResourceDiscoveryService {
	service := &ResourceDiscoveryService{}
	base := NewDiscoveryService(state, registry, Options[domain.ResourceSnapshot]{
		GetIndex: func(rt *runtime.State) snapshotIndex[domain.ResourceSnapshot] { return rt.Resources() },
	})
	service.Service = base
	base.filterSnapshot = service.filterResourceSnapshot
	return service
}

// ListResources lists resources visible to a client.
func (d *ResourceDiscoveryService) ListResources(ctx context.Context, client string, cursor string) (domain.ResourcePage, error) {
	snapshot, err := d.ListSnapshot(ctx, client)
	if err != nil {
		return domain.ResourcePage{}, err
	}
	return paginateResources(snapshot, cursor)
}

// ListResourcesAll lists resources across all servers.
func (d *ResourceDiscoveryService) ListResourcesAll(ctx context.Context, cursor string) (domain.ResourcePage, error) {
	snapshot, err := d.ListSnapshotAll(ctx)
	if err != nil {
		return domain.ResourcePage{}, err
	}
	return paginateResources(snapshot, cursor)
}

// WatchResources streams resource snapshots for a client.
func (d *ResourceDiscoveryService) WatchResources(ctx context.Context, client string) (<-chan domain.ResourceSnapshot, error) {
	return d.WatchSnapshots(ctx, client)
}

// ReadResource reads a resource on behalf of a client.
func (d *ResourceDiscoveryService) ReadResource(ctx context.Context, client, uri string) (json.RawMessage, error) {
	serverName, err := d.resolveClientServer(client)
	if err != nil {
		return nil, err
	}
	runtime := d.state.RuntimeState()
	if runtime == nil || runtime.Resources() == nil {
		return nil, domain.ErrResourceNotFound
	}
	if serverName != "" {
		if _, ok := runtime.Resources().ResolveForServer(serverName, uri); !ok {
			return nil, domain.ErrResourceNotFound
		}
		ctx = domain.WithRouteContext(ctx, domain.RouteContext{Client: client})
		return runtime.Resources().ReadResourceForServer(ctx, serverName, uri)
	}
	visibleSpecKeys, err := d.resolveVisibleSpecKeys(client)
	if err != nil {
		return nil, err
	}
	target, ok := runtime.Resources().Resolve(uri)
	if !ok {
		return nil, domain.ErrResourceNotFound
	}
	visibleSpecSet := toSpecKeySet(visibleSpecKeys)
	if target.SpecKey != "" {
		if _, ok := visibleSpecSet[target.SpecKey]; !ok {
			return nil, domain.ErrResourceNotFound
		}
	} else if !d.isServerVisible(visibleSpecSet, target.ServerType) {
		return nil, domain.ErrResourceNotFound
	}
	ctx = domain.WithRouteContext(ctx, domain.RouteContext{Client: client})
	return runtime.Resources().ReadResource(ctx, uri)
}

// ReadResourceAll reads a resource without client visibility checks.
func (d *ResourceDiscoveryService) ReadResourceAll(ctx context.Context, uri string) (json.RawMessage, error) {
	runtime := d.state.RuntimeState()
	if runtime == nil || runtime.Resources() == nil {
		return nil, domain.ErrResourceNotFound
	}
	if _, ok := runtime.Resources().Resolve(uri); !ok {
		return nil, domain.ErrResourceNotFound
	}
	ctx = domain.WithRouteContext(ctx, domain.RouteContext{Client: domain.InternalUIClientName})
	return runtime.Resources().ReadResource(ctx, uri)
}

func (d *ResourceDiscoveryService) filterResourceSnapshot(snapshot domain.ResourceSnapshot, visibleSpecKeys []string) domain.ResourceSnapshot {
	if len(snapshot.Resources) == 0 {
		return domain.ResourceSnapshot{}
	}
	visibleServers, visibleSpecSet := d.visibleServers(visibleSpecKeys)
	filtered := make([]domain.ResourceDefinition, 0, len(snapshot.Resources))
	for _, resource := range snapshot.Resources {
		if resource.ServerName != "" {
			if _, ok := visibleServers[resource.ServerName]; !ok {
				continue
			}
		} else if resource.SpecKey != "" {
			if _, ok := visibleSpecSet[resource.SpecKey]; !ok {
				continue
			}
		}
		filtered = append(filtered, resource)
	}
	if len(filtered) == 0 {
		return domain.ResourceSnapshot{}
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].URI < filtered[j].URI
	})
	return domain.ResourceSnapshot{
		ETag:      hashutil.ResourceETag(d.state.Logger(), filtered),
		Resources: filtered,
	}
}
