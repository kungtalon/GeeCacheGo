package cache

import (
	"container/list"
	"errors"
	"strings"
)

type Value interface {
	Len() int
}

type Cache interface {
	Get(string) (Value, bool)
	Add(string, Value)
	Pop()
	Len() int
}

type LRUCache struct {
	maxBytes   int64
	nbytes     int64
	ll         *list.List
	dictionary map[string]*list.Element
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

const (
	LRU = "lru"
	LFU = "lfu"
)

func New(cacheType string, maxBytes int64, onEvicted func(string, Value)) Cache {
	if strings.EqualFold(cacheType, LRU) {
		return &LRUCache{
			maxBytes:   maxBytes,
			ll:         list.New(),
			dictionary: make(map[string]*list.Element),
			OnEvicted:  onEvicted,
		}
	} else if strings.EqualFold(cacheType, LFU) {
		panic(errors.New("Unimplemented Cache Type!"))
	}
	return nil
}

func (c *LRUCache) Get(key string) (value Value, isOk bool) {
	if ele, ok := c.dictionary[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

func (c *LRUCache) Pop() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.dictionary, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *LRUCache) Add(key string, value Value) {
	if ele, ok := c.dictionary[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.dictionary[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.Pop()
	}
}

func (c *LRUCache) Len() int {
	return c.ll.Len()
}
