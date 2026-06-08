package compression

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

type ContextCache struct {
	mu       sync.RWMutex
	entries  map[string]*CacheEntry
	maxSize  int
	hitCount int
	missCount int
}

func NewContextCache(maxSize int) *ContextCache {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &ContextCache{
		entries: make(map[string]*CacheEntry),
		maxSize: maxSize,
	}
}

func (c *ContextCache) Get(key string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		c.missCount++
		return nil, false
	}
	entry.HitCount++
	c.hitCount++
	return entry, true
}

func (c *ContextCache) Set(key string, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= c.maxSize {
		c.evictLRU()
	}

	entry.CreatedAt = time.Now()
	c.entries[key] = entry
}

func (c *ContextCache) KeyForContext(ctx Context) string {
	h := sha256.New()
	h.Write([]byte(ctx.SystemPrompt))
	h.Write([]byte(ctx.AgentMemory))
	h.Write([]byte(ctx.Config))
	h.Write([]byte(ctx.SecurityRules))

	for _, msg := range ctx.Messages {
		h.Write([]byte(msg.Role))
		h.Write([]byte(msg.Content))
	}

	return fmt.Sprintf("%x", h.Sum(nil)[:16])
}

func (c *ContextCache) Stats() (hits, misses, size int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hitCount, c.missCount, len(c.entries)
}

func (c *ContextCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, entry := range c.entries {
		if first || entry.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.CreatedAt
			first = false
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}
