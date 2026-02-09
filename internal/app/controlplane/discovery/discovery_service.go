package discovery

import (
	"context"

	"mcpv/internal/app/controlplane/registry"
	"mcpv/internal/app/runtime"
)

type snapshotIndex[Snapshot any] interface {
	Snapshot() Snapshot
	Subscribe(context.Context) <-chan Snapshot
	SnapshotForServer(serverName string) (Snapshot, bool)
}

type Options[Snapshot any] struct {
	GetIndex               func(*runtime.State) snapshotIndex[Snapshot]
	FilterSnapshot         func(Snapshot, []string) Snapshot
	ServerSnapshotForList  func(serverName string, index snapshotIndex[Snapshot]) (Snapshot, bool)
	ServerSnapshotForWatch func(serverName string, index snapshotIndex[Snapshot]) (Snapshot, bool)
}

type Service[Snapshot any] struct {
	discoverySupport
	getIndex               func(*runtime.State) snapshotIndex[Snapshot]
	filterSnapshot         func(Snapshot, []string) Snapshot
	serverSnapshotForList  func(serverName string, index snapshotIndex[Snapshot]) (Snapshot, bool)
	serverSnapshotForWatch func(serverName string, index snapshotIndex[Snapshot]) (Snapshot, bool)
}

func NewDiscoveryService[Snapshot any](state State, registry *registry.ClientRegistry, opts Options[Snapshot]) *Service[Snapshot] {
	service := &Service[Snapshot]{
		discoverySupport:       newDiscoverySupport(state, registry),
		getIndex:               opts.GetIndex,
		filterSnapshot:         opts.FilterSnapshot,
		serverSnapshotForList:  opts.ServerSnapshotForList,
		serverSnapshotForWatch: opts.ServerSnapshotForWatch,
	}
	if service.filterSnapshot == nil {
		service.filterSnapshot = func(snapshot Snapshot, _ []string) Snapshot {
			return snapshot
		}
	}
	if service.serverSnapshotForList == nil {
		service.serverSnapshotForList = func(serverName string, index snapshotIndex[Snapshot]) (Snapshot, bool) {
			return index.SnapshotForServer(serverName)
		}
	}
	if service.serverSnapshotForWatch == nil {
		service.serverSnapshotForWatch = func(serverName string, index snapshotIndex[Snapshot]) (Snapshot, bool) {
			return index.SnapshotForServer(serverName)
		}
	}
	return service
}

func (d *Service[Snapshot]) ListSnapshot(_ context.Context, client string) (Snapshot, error) {
	serverName, err := d.resolveClientServer(client)
	if err != nil {
		return zeroSnapshot[Snapshot](), err
	}
	runtimeState := d.state.RuntimeState()
	if runtimeState == nil {
		return zeroSnapshot[Snapshot](), nil
	}
	index := d.getIndex(runtimeState)
	if index == nil {
		return zeroSnapshot[Snapshot](), nil
	}
	if serverName != "" {
		snapshot, ok := d.serverSnapshotForList(serverName, index)
		if !ok {
			return zeroSnapshot[Snapshot](), nil
		}
		return snapshot, nil
	}
	visibleSpecKeys, err := d.resolveVisibleSpecKeys(client)
	if err != nil {
		return zeroSnapshot[Snapshot](), err
	}
	snapshot := index.Snapshot()
	return d.filterSnapshot(snapshot, visibleSpecKeys), nil
}

func (d *Service[Snapshot]) ListSnapshotAll(_ context.Context) (Snapshot, error) {
	runtimeState := d.state.RuntimeState()
	if runtimeState == nil {
		return zeroSnapshot[Snapshot](), nil
	}
	index := d.getIndex(runtimeState)
	if index == nil {
		return zeroSnapshot[Snapshot](), nil
	}
	return index.Snapshot(), nil
}

func (d *Service[Snapshot]) WatchSnapshots(ctx context.Context, client string) (<-chan Snapshot, error) {
	if _, err := d.resolveClientServer(client); err != nil {
		return closedSnapshotChannel[Snapshot](), err
	}
	runtimeState := d.state.RuntimeState()
	if runtimeState == nil {
		return closedSnapshotChannel[Snapshot](), nil
	}
	index := d.getIndex(runtimeState)
	if index == nil {
		return closedSnapshotChannel[Snapshot](), nil
	}

	output := make(chan Snapshot, 1)
	indexCh := index.Subscribe(ctx)
	changes := d.registry.WatchClientChanges(ctx)

	go func() {
		defer close(output)
		last := index.Snapshot()
		d.sendFilteredSnapshot(output, client, last)
		for {
			select {
			case <-ctx.Done():
				return
			case snapshot, ok := <-indexCh:
				if !ok {
					return
				}
				last = snapshot
				d.sendFilteredSnapshot(output, client, snapshot)
			case event, ok := <-changes:
				if !ok {
					return
				}
				if event.Client == client {
					d.sendFilteredSnapshot(output, client, last)
				}
			}
		}
	}()

	return output, nil
}

func (d *Service[Snapshot]) sendFilteredSnapshot(ch chan<- Snapshot, client string, snapshot Snapshot) {
	serverName, err := d.resolveClientServer(client)
	if err != nil {
		return
	}
	runtimeState := d.state.RuntimeState()
	if runtimeState == nil {
		return
	}
	index := d.getIndex(runtimeState)
	if index == nil {
		return
	}
	if serverName != "" {
		serverSnapshot, ok := d.serverSnapshotForWatch(serverName, index)
		if !ok {
			return
		}
		select {
		case ch <- serverSnapshot:
		default:
		}
		return
	}
	visibleSpecKeys, err := d.resolveVisibleSpecKeys(client)
	if err != nil {
		return
	}
	filtered := d.filterSnapshot(snapshot, visibleSpecKeys)
	select {
	case ch <- filtered:
	default:
	}
}

func closedSnapshotChannel[Snapshot any]() chan Snapshot {
	ch := make(chan Snapshot)
	close(ch)
	return ch
}

func zeroSnapshot[Snapshot any]() Snapshot {
	var zero Snapshot
	return zero
}
