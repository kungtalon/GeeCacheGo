package cache

import (
	"errors"
	"log"
	"sync"
)

// Getter loads data for a key if cache misses
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc is a functional interface which will implement Getter
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group is a cache namespace
type Group struct {
	name      string
	getter    Getter
	mainCache syncCache
}

var (
	groups     = make(map[string]*Group)
	groupsLock sync.RWMutex
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter!")
	}

	groupsLock.Lock()
	defer groupsLock.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: syncCache{cacheBytes: cacheBytes},
	}

	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	groupsLock.RLock()
	defer groupsLock.RUnlock()
	g := groups[name]
	return g
}

// Get returns the value from the cache, call getter if missed
func (g *Group) Get(key string) (value ByteView, err error) {
	if key == "" {
		return ByteView{}, errors.New("Key is required")
	}

	v, ok := g.mainCache.get(key)
	if !ok {
		log.Printf("[GeeCache] Cache miss with key: %s", key)
		vbytes, err := g.getter.Get(key)
		if err != nil {
			return ByteView{}, err
		}
		value = ByteView{vbytes}
		g.mainCache.add(key, value)
		return value, err
	}

	log.Printf("[GeeCache] Cache hit with key: %s, value: %s", key, v.String())
	return v, nil
}
