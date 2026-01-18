package domain

import "context"

// CatalogLoader loads catalog data from a path.
type CatalogLoader interface {
	Load(ctx context.Context, path string) (Catalog, error)
}
