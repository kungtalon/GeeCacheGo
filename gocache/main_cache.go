package gocache

import (
	"errors"
	"gocache/cache"
	"gocache/singleflight"
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
	peers     PeerPicker
	// use singleflight.Group to make sure that
	// each key is only fetched once
	loader *singleflight.Group
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
		loader:    &singleflight.Group{},
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
func (g *Group) Get(key string) (value cache.ByteView, err error) {
	if key == "" {
		return cache.ByteView{}, errors.New("Key is required")
	}

	v, ok := g.mainCache.get(key)
	if !ok {
		log.Printf("[GeeCache] Cache miss with key: %s", key)
		vbytes, err := g.getter.Get(key)
		if err != nil {
			return cache.ByteView{}, err
		}
		value = cache.ByteView{vbytes}
		g.mainCache.add(key, value)
		return value, err
	}

	log.Printf("[GeeCache] Cache hit with key: %s, value: %s", key, v.String())
	return v, nil
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value cache.ByteView, err error) {
	// each key is only fetched once (either locally or remotely)
	// regardless of the number of concurrent callers.
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}

		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(cache.ByteView), nil
	}
	return
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (cache.ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return cache.ByteView{}, err
	}
	return cache.ByteView{bytes}, nil
}

func (g *Group) getLocally(key string) (cache.ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return cache.ByteView{}, err
	}
	value := cache.ByteView{cache.CloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value cache.ByteView) {
	g.mainCache.add(key, value)
}
