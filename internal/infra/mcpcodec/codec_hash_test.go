package mcpcodec

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mcpv/internal/domain"
)

// TestHashToolDefinition_Deterministic verifies that hashing is deterministic.
func TestHashToolDefinition_Deterministic(t *testing.T) {
	tests := []struct {
		name     string
		tool1    domain.ToolDefinition
		tool2    domain.ToolDefinition
		sameHash bool
	}{
		{
			name: "identical tools produce same hash",
			tool1: domain.ToolDefinition{
				Name:        "test_tool",
				Description: "A test tool",
				InputSchema: map[string]any{"type": "object"},
			},
			tool2: domain.ToolDefinition{
				Name:        "test_tool",
				Description: "A test tool",
				InputSchema: map[string]any{"type": "object"},
			},
			sameHash: true,
		},
		{
			name: "different names produce different hashes",
			tool1: domain.ToolDefinition{
				Name:        "tool_a",
				Description: "A test tool",
			},
			tool2: domain.ToolDefinition{
				Name:        "tool_b",
				Description: "A test tool",
			},
			sameHash: false,
		},
		{
			name: "different descriptions produce different hashes",
			tool1: domain.ToolDefinition{
				Name:        "test_tool",
				Description: "Description A",
			},
			tool2: domain.ToolDefinition{
				Name:        "test_tool",
				Description: "Description B",
			},
			sameHash: false,
		},
		{
			name: "different schemas produce different hashes",
			tool1: domain.ToolDefinition{
				Name:        "test_tool",
				InputSchema: map[string]any{"type": "object"},
			},
			tool2: domain.ToolDefinition{
				Name:        "test_tool",
				InputSchema: map[string]any{"type": "array"},
			},
			sameHash: false,
		},
		{
			name:     "empty tools produce same hash",
			tool1:    domain.ToolDefinition{},
			tool2:    domain.ToolDefinition{},
			sameHash: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1, err1 := HashToolDefinition(tt.tool1)
			hash2, err2 := HashToolDefinition(tt.tool2)

			require.NoError(t, err1)
			require.NoError(t, err2)

			if tt.sameHash {
				assert.Equal(t, hash1, hash2, "Expected identical hashes")
			} else {
				assert.NotEqual(t, hash1, hash2, "Expected different hashes")
			}
		})
	}
}

// TestHashToolDefinition_Concurrent verifies thread-safe hashing.
func TestHashToolDefinition_Concurrent(t *testing.T) {
	tool := domain.ToolDefinition{
		Name:        "concurrent_tool",
		Description: "Test concurrent hashing",
		InputSchema: map[string]any{"type": "object"},
	}

	const goroutines = 100
	hashes := make([]string, goroutines)
	errs := make([]error, goroutines)
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			hash, err := HashToolDefinition(tool)
			hashes[idx] = hash
			errs[idx] = err
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		require.NoError(t, err, "hash error at index %d", i)
	}

	expectedHash := hashes[0]
	for i, hash := range hashes {
		assert.Equal(t, expectedHash, hash, "Hash mismatch at index %d", i)
	}
}

// TestHashToolDefinitions_ListHashing verifies list hashing behavior.
func TestHashToolDefinitions_ListHashing(t *testing.T) {
	tool1 := domain.ToolDefinition{Name: "tool1", Description: "First tool"}
	tool2 := domain.ToolDefinition{Name: "tool2", Description: "Second tool"}
	tool3 := domain.ToolDefinition{Name: "tool3", Description: "Third tool"}

	tests := []struct {
		name     string
		list1    []domain.ToolDefinition
		list2    []domain.ToolDefinition
		sameHash bool
	}{
		{
			name:     "identical lists produce same hash",
			list1:    []domain.ToolDefinition{tool1, tool2},
			list2:    []domain.ToolDefinition{tool1, tool2},
			sameHash: true,
		},
		{
			name:     "different order produces different hash",
			list1:    []domain.ToolDefinition{tool1, tool2},
			list2:    []domain.ToolDefinition{tool2, tool1},
			sameHash: false,
		},
		{
			name:     "different length produces different hash",
			list1:    []domain.ToolDefinition{tool1, tool2},
			list2:    []domain.ToolDefinition{tool1, tool2, tool3},
			sameHash: false,
		},
		{
			name:     "empty lists produce same hash",
			list1:    []domain.ToolDefinition{},
			list2:    []domain.ToolDefinition{},
			sameHash: true,
		},
		{
			name:     "nil lists produce same hash",
			list1:    nil,
			list2:    nil,
			sameHash: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1, err1 := HashToolDefinitions(tt.list1)
			hash2, err2 := HashToolDefinitions(tt.list2)

			require.NoError(t, err1)
			require.NoError(t, err2)

			if tt.sameHash {
				assert.Equal(t, hash1, hash2, "Expected identical hashes")
			} else {
				assert.NotEqual(t, hash1, hash2, "Expected different hashes")
			}
		})
	}
}

// TestHashResourceDefinitions_ListHashing verifies resource list hashing.
func TestHashResourceDefinitions_ListHashing(t *testing.T) {
	resource1 := domain.ResourceDefinition{URI: "file:///a", Name: "resource1"}
	resource2 := domain.ResourceDefinition{URI: "file:///b", Name: "resource2"}

	tests := []struct {
		name     string
		list1    []domain.ResourceDefinition
		list2    []domain.ResourceDefinition
		sameHash bool
	}{
		{
			name:     "identical lists produce same hash",
			list1:    []domain.ResourceDefinition{resource1, resource2},
			list2:    []domain.ResourceDefinition{resource1, resource2},
			sameHash: true,
		},
		{
			name:     "different order produces different hash",
			list1:    []domain.ResourceDefinition{resource1, resource2},
			list2:    []domain.ResourceDefinition{resource2, resource1},
			sameHash: false,
		},
		{
			name:     "empty lists produce same hash",
			list1:    []domain.ResourceDefinition{},
			list2:    []domain.ResourceDefinition{},
			sameHash: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1, err1 := HashResourceDefinitions(tt.list1)
			hash2, err2 := HashResourceDefinitions(tt.list2)

			require.NoError(t, err1)
			require.NoError(t, err2)

			if tt.sameHash {
				assert.Equal(t, hash1, hash2)
			} else {
				assert.NotEqual(t, hash1, hash2)
			}
		})
	}
}

// TestHashPromptDefinitions_ListHashing verifies prompt list hashing.
func TestHashPromptDefinitions_ListHashing(t *testing.T) {
	prompt1 := domain.PromptDefinition{Name: "prompt1", Description: "First"}
	prompt2 := domain.PromptDefinition{Name: "prompt2", Description: "Second"}

	tests := []struct {
		name     string
		list1    []domain.PromptDefinition
		list2    []domain.PromptDefinition
		sameHash bool
	}{
		{
			name:     "identical lists produce same hash",
			list1:    []domain.PromptDefinition{prompt1, prompt2},
			list2:    []domain.PromptDefinition{prompt1, prompt2},
			sameHash: true,
		},
		{
			name:     "different order produces different hash",
			list1:    []domain.PromptDefinition{prompt1, prompt2},
			list2:    []domain.PromptDefinition{prompt2, prompt1},
			sameHash: false,
		},
		{
			name:     "empty lists produce same hash",
			list1:    []domain.PromptDefinition{},
			list2:    []domain.PromptDefinition{},
			sameHash: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1, err1 := HashPromptDefinitions(tt.list1)
			hash2, err2 := HashPromptDefinitions(tt.list2)

			require.NoError(t, err1)
			require.NoError(t, err2)

			if tt.sameHash {
				assert.Equal(t, hash1, hash2)
			} else {
				assert.NotEqual(t, hash1, hash2)
			}
		})
	}
}
