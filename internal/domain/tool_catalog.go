package domain

import "time"

// ToolSource indicates where tool metadata was obtained.
type ToolSource string

const (
	ToolSourceLive  ToolSource = "live"
	ToolSourceCache ToolSource = "cache"
)

// ToolCatalogEntry represents a tool with origin metadata.
type ToolCatalogEntry struct {
	Definition ToolDefinition
	Source     ToolSource
	CachedAt   time.Time
}

// ToolCatalogSnapshot is a merged snapshot of tool metadata.
type ToolCatalogSnapshot struct {
	ETag  string
	Tools []ToolCatalogEntry
}
