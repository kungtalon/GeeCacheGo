package cache

import "sync"

type syncCache struct {
	mu         sync.Mutex
	cache      Cache
	cacheBytes int64
}

func (c *syncCache) add(key string, value ByteView) (ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		c.cache = New("lru", c.cacheBytes, nil)
	}
	c.cache.Add(key, value)
	return
}

func (c *syncCache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		return
	}

	if v, ok := c.cache.Get(key); ok {
		return v.(ByteView), ok
	}

	return
}
