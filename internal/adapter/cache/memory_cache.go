package cache

import (
	"context"
	"sync"
	"time"

	"github.com/yoadey/shiftmanager/internal/port"
)

// entry is a single cached value with an optional expiry time.
type entry struct {
	value     string
	expiresAt time.Time // zero means "never expires"
}

func (e entry) expired(now time.Time) bool {
	return !e.expiresAt.IsZero() && now.After(e.expiresAt)
}

// MemoryCache is an in-process implementation of port.CacheService backed by a
// sync.Map with per-key TTL support. It is safe for concurrent use.
type MemoryCache struct {
	data sync.Map // map[string]entry

	stopOnce sync.Once
	stop     chan struct{}
}

// compile-time assertion that MemoryCache satisfies the port interface.
var _ port.CacheService = (*MemoryCache)(nil)

// NewMemoryCache creates a MemoryCache and starts a background janitor that
// periodically evicts expired entries.
func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{stop: make(chan struct{})}
	go c.janitor(time.Minute)
	return c
}

// Get returns the value stored for key or port.ErrCacheMiss when absent/expired.
func (c *MemoryCache) Get(_ context.Context, key string) (string, error) {
	v, ok := c.data.Load(key)
	if !ok {
		return "", port.ErrCacheMiss
	}
	e := v.(entry)
	if e.expired(time.Now()) {
		c.data.Delete(key)
		return "", port.ErrCacheMiss
	}
	return e.value, nil
}

// Set stores value under key. A non-positive ttl stores the value without expiry.
func (c *MemoryCache) Set(_ context.Context, key, value string, ttl time.Duration) error {
	e := entry{value: value}
	if ttl > 0 {
		e.expiresAt = time.Now().Add(ttl)
	}
	c.data.Store(key, e)
	return nil
}

// Delete removes key from the cache.
func (c *MemoryCache) Delete(_ context.Context, key string) error {
	c.data.Delete(key)
	return nil
}

// Flush removes all entries from the cache.
func (c *MemoryCache) Flush(_ context.Context) error {
	c.data.Range(func(k, _ any) bool {
		c.data.Delete(k)
		return true
	})
	return nil
}

// Close stops the background janitor. Safe to call multiple times.
func (c *MemoryCache) Close() {
	c.stopOnce.Do(func() { close(c.stop) })
}

func (c *MemoryCache) janitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-c.stop:
			return
		case now := <-ticker.C:
			c.data.Range(func(k, v any) bool {
				if v.(entry).expired(now) {
					c.data.Delete(k)
				}
				return true
			})
		}
	}
}
