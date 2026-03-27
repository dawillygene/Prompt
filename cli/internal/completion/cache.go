package completion

import (
	"sync"
	"time"
)

// CacheEntry represents a cached completion result
type CacheEntry struct {
	Data      []string
	Timestamp time.Time
}

// Cache provides a time-based cache for completion results
type Cache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
}

// NewCache creates a new completion cache with the specified TTL
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves a value from the cache if it exists and hasn't expired
func (c *Cache) Get(key string) ([]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.Timestamp) > c.ttl {
		return nil, false
	}

	return entry.Data, true
}

// Set stores a value in the cache
func (c *Cache) Set(key string, data []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
	}
}

// Invalidate removes a specific key from the cache
func (c *Cache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
}

// CleanExpired removes all expired entries from the cache
func (c *Cache) CleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.Sub(entry.Timestamp) > c.ttl {
			delete(c.entries, key)
		}
	}
}
