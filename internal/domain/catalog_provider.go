package domain

import "context"

// CatalogUpdateSource describes why a catalog update occurred.
type CatalogUpdateSource string

const (
	// CatalogUpdateSourceBootstrap indicates a bootstrap-triggered update.
	CatalogUpdateSourceBootstrap CatalogUpdateSource = "bootstrap"
	// CatalogUpdateSourceWatch indicates a filesystem watcher update.
	CatalogUpdateSourceWatch CatalogUpdateSource = "watch"
	// CatalogUpdateSourceManual indicates a manual reload update.
	CatalogUpdateSourceManual CatalogUpdateSource = "manual"
)

// CatalogUpdate carries a snapshot and diff for catalog changes.
type CatalogUpdate struct {
	Snapshot CatalogState
	Diff     CatalogDiff
	Source   CatalogUpdateSource
}

// CatalogProvider provides catalog snapshots and change notifications.
type CatalogProvider interface {
	Snapshot(ctx context.Context) (CatalogState, error)
	Watch(ctx context.Context) (<-chan CatalogUpdate, error)
	Reload(ctx context.Context) error
}
