package domain

import "time"

// CatalogState captures the current catalog snapshot and metadata.
type CatalogState struct {
	Store    ProfileStore
	Summary  CatalogSummary
	Revision uint64
	LoadedAt time.Time
}

// NewCatalogState builds a catalog state from a profile store.
func NewCatalogState(store ProfileStore, revision uint64, loadedAt time.Time) (CatalogState, error) {
	if loadedAt.IsZero() {
		loadedAt = time.Now()
	}
	summary, err := BuildCatalogSummary(store)
	if err != nil {
		return CatalogState{}, err
	}
	return CatalogState{
		Store:    store,
		Summary:  summary,
		Revision: revision,
		LoadedAt: loadedAt,
	}, nil
}
