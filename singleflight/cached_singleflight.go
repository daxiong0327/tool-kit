// Package singleflight provides a duplicate function call suppression
// mechanism with caching support.
package singleflight

import (
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"
)

// Result represents the result of a Do call
type Result struct {
	Value     interface{}
	Err       error
	Shared    bool
	FromCache bool
}

// CachedGroup represents a singleflight instance with caching capability
type CachedGroup struct {
	sf    singleflight.Group
	cache *cache.Cache
}

// NewCachedGroup creates a new CachedGroup instance
func NewCachedGroup(defaultExpiration, cleanupInterval time.Duration) *CachedGroup {
	return &CachedGroup{
		sf:    singleflight.Group{},
		cache: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Get retrieves a value from cache directly
func (g *CachedGroup) Get(key string) (interface{}, bool) {
	return g.cache.Get(key)
}

// Set sets a value in cache directly
func (g *CachedGroup) Set(key string, value interface{}, d time.Duration) {
	g.cache.Set(key, value, d)
}

// Delete removes a value from cache
func (g *CachedGroup) Delete(key string) {
	g.cache.Delete(key)
}

// Do executes the function and caches its result
func (g *CachedGroup) Do(key string, ttl time.Duration, fn func() (interface{}, error)) Result {
	// Check cache first
	if v, found := g.cache.Get(key); found {
		return Result{Value: v, FromCache: true}
	}

	// Use singleflight to handle concurrent calls
	v, err, shared := g.sf.Do(key, func() (interface{}, error) {
		// Double check cache
		if v, found := g.cache.Get(key); found {
			return v, nil
		}

		return fn()
	})

	result := Result{
		Value:  v,
		Err:    err,
		Shared: shared,
	}

	// Cache successful results
	if err == nil {
		g.cache.Set(key, v, ttl)
	}

	return result
}

// DoWithFallback executes the function with a fallback value
func (g *CachedGroup) DoWithFallback(key string, ttl time.Duration, fn func() (interface{}, error), fallback interface{}) interface{} {
	result := g.Do(key, ttl, fn)
	if result.Err != nil {
		return fallback
	}
	return result.Value
}

// Forget removes both the in-flight operation and cached value
func (g *CachedGroup) Forget(key string) {
	g.sf.Forget(key)
	g.cache.Delete(key)
}

// FlushCache removes all items from the cache
func (g *CachedGroup) FlushCache() {
	g.cache.Flush()
}

// GetCacheStats returns basic stats about the cache
func (g *CachedGroup) GetCacheStats() (items int, hits int64, misses int64) {
	return g.cache.ItemCount(), 0, 0
}
