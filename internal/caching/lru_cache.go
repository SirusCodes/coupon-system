package caching

import (
	"time"

	e "github.com/hashicorp/golang-lru/v2/expirable"
)

// LRUCache is an LRU cache implementation with expirability.
type LRUCache[K comparable, V any] struct {
	cache      *e.LRU[K, V]
	defaultTTL time.Duration
}

// NewLRUCache creates a new LRUCache.
func NewLRUCache[K comparable, V any](maxEntries int, defaultTTL time.Duration) *LRUCache[K, V] {
	c := e.NewLRU[K, V](maxEntries, nil, defaultTTL)
	return &LRUCache[K, V]{
		cache:      c,
		defaultTTL: defaultTTL,
	}
}

// Get retrieves a value from the cache.
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	val, ok := c.cache.Get(key)
	return val, ok
}

// Set adds or updates a value in the cache with a TTL.
func (c *LRUCache[K, V]) Set(key K, value V) {
	c.cache.Add(key, value)
}

// Delete removes a value from the cache.
func (c *LRUCache[K, V]) Delete(key K) {
	c.cache.Remove(key)
}
