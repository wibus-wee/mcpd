package core

import (
	"sort"
	"time"

	"mcpv/internal/domain"
)

func SortedServerTypes[T any](specs map[string]T) []string {
	if len(specs) == 0 {
		return nil
	}
	serverTypes := make([]string, 0, len(specs))
	for serverType := range specs {
		serverTypes = append(serverTypes, serverType)
	}
	sort.Strings(serverTypes)
	return serverTypes
}

func RefreshTimeout(cfg domain.RuntimeConfig) time.Duration {
	return cfg.RouteTimeout()
}
