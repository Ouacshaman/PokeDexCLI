package pokecache

import (
	"sync"
	"time"
)

type Cache struct{
	cache map[string]cacheEntry
	mu sync.Mutex
	interval time.Duration
}

type cacheEntry struct{
	createdAt time.Time
	val []byte
}

func NewCache(t time.Duration) *Cache{
	res := Cache{
		cache: make(map[string]cacheEntry),
		interval: t,
	}
	go res.reapLoop(res.interval)
	return &res
}

func (c Cache) Add(key string, value []byte){
	c.mu.Lock()
	defer c.mu.Unlock()
	entry := cacheEntry{
		createdAt: time.Now(),
		val: value,
	}
	c.cache[key] = entry
}

func (c Cache) Get(key string) ([]byte, bool){
	c.mu.Lock()
	defer c.mu.Unlock()
	elem, ok := c.cache[key]
	if !ok{
		return nil,false
	}
	return elem.val, true
}

func (c *Cache) reapLoop(dur time.Duration) {
    ticker := time.NewTicker(dur)
    defer ticker.Stop()
    for range ticker.C {
		c.mu.Lock()
		for key, val := range c.cache{
			creationTime := val.createdAt
			interval := time.Now().Sub(creationTime)
			if interval > dur{
				delete(c.cache, key)
			}
		}
		c.mu.Unlock()
    }
}










