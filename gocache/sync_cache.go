package gocache

import (
	cache2 "gocache/cache"
	"sync"
)

type syncCache struct {
	mu         sync.Mutex
	cache      cache2.Cache
	cacheBytes int64
}

func (c *syncCache) add(key string, value cache2.ByteView) (ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		c.cache = cache2.New("lru", c.cacheBytes, nil)
	}
	c.cache.Add(key, value)
	return
}

func (c *syncCache) get(key string) (value cache2.ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		return
	}

	if v, ok := c.cache.Get(key); ok {
		return v.(cache2.ByteView), ok
	}

	return
}
