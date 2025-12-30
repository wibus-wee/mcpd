package domain

import "context"

type CatalogLoader interface {
	Load(ctx context.Context, path string) (Catalog, error)
}
