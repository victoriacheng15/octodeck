package api

import (
	"sync"
	"time"
)

type cacheEntry struct {
	value     any
	expiresAt time.Time
}

// Cache provides a thread-safe in-memory store with per-item TTL expiration.
type Cache struct {
	mu         sync.RWMutex
	items      map[string]cacheEntry
	defaultTTL time.Duration
}

// NewCache constructs a cache instance with a baseline TTL.
func NewCache(defaultTTL time.Duration) *Cache {
	return &Cache{
		items:      make(map[string]cacheEntry),
		defaultTTL: defaultTTL,
	}
}

// Get retrieves an unexpired item by key.
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.expiresAt) {
		return nil, false
	}

	return item.value, true
}

// Set stores an item with an explicit TTL, falling back to defaultTTL if non-positive.
func (c *Cache) Set(key string, value any, ttl time.Duration) {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

// Delete removes an item from cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear flushes all cached entries immediately.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]cacheEntry)
}
