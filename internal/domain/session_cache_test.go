package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionCache_Deduplication(t *testing.T) {
	cache := NewSessionCache(time.Hour, 100)

	// First request should need full schema
	assert.True(t, cache.NeedsFull("caller1", "tool1", "hash123"))

	// Update cache
	cache.Update("caller1", map[string]string{"tool1": "hash123"})

	// Same hash should not need full
	assert.False(t, cache.NeedsFull("caller1", "tool1", "hash123"))

	// Different hash should need full
	assert.True(t, cache.NeedsFull("caller1", "tool1", "hash456"))

	// Different caller should need full
	assert.True(t, cache.NeedsFull("caller2", "tool1", "hash123"))
}

func TestSessionCache_Invalidation(t *testing.T) {
	cache := NewSessionCache(time.Hour, 100)

	// Setup
	cache.Update("caller1", map[string]string{"tool1": "hash123"})
	assert.False(t, cache.NeedsFull("caller1", "tool1", "hash123"))

	// Invalidate
	cache.Invalidate("caller1")

	// Should need full again
	assert.True(t, cache.NeedsFull("caller1", "tool1", "hash123"))
}

func TestSessionCache_TTLExpiration(t *testing.T) {
	// Use very short TTL
	cache := NewSessionCache(10*time.Millisecond, 100)

	cache.Update("caller1", map[string]string{"tool1": "hash123"})
	assert.False(t, cache.NeedsFull("caller1", "tool1", "hash123"))

	// Wait for TTL to expire
	time.Sleep(15 * time.Millisecond)

	// Should need full after TTL expires
	assert.True(t, cache.NeedsFull("caller1", "tool1", "hash123"))
}

func TestSessionCache_MaxSize(t *testing.T) {
	cache := NewSessionCache(time.Hour, 3)

	cache.Update("caller1", map[string]string{"tool1": "hash1"})
	cache.Update("caller2", map[string]string{"tool2": "hash2"})
	cache.Update("caller3", map[string]string{"tool3": "hash3"})

	require.Equal(t, 3, cache.Size())

	// Adding a 4th should evict the oldest
	time.Sleep(1 * time.Millisecond) // Ensure different timestamp
	cache.Update("caller4", map[string]string{"tool4": "hash4"})

	require.Equal(t, 3, cache.Size())

	// caller4 should exist
	assert.False(t, cache.NeedsFull("caller4", "tool4", "hash4"))
}

func TestSessionCache_Cleanup(t *testing.T) {
	cache := NewSessionCache(10*time.Millisecond, 100)

	cache.Update("caller1", map[string]string{"tool1": "hash1"})
	cache.Update("caller2", map[string]string{"tool2": "hash2"})

	require.Equal(t, 2, cache.Size())

	// Wait for TTL
	time.Sleep(15 * time.Millisecond)

	// Cleanup should remove expired entries
	cache.Cleanup()

	require.Equal(t, 0, cache.Size())
}

func TestSessionCache_UpdateMergesSchemas(t *testing.T) {
	cache := NewSessionCache(time.Hour, 100)

	// First update
	cache.Update("caller1", map[string]string{"tool1": "hash1"})
	assert.False(t, cache.NeedsFull("caller1", "tool1", "hash1"))
	assert.True(t, cache.NeedsFull("caller1", "tool2", "hash2"))

	// Second update adds more tools
	cache.Update("caller1", map[string]string{"tool2": "hash2"})

	// Both should now be cached
	assert.False(t, cache.NeedsFull("caller1", "tool1", "hash1"))
	assert.False(t, cache.NeedsFull("caller1", "tool2", "hash2"))
}
