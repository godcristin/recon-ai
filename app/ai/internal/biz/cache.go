package biz

import (
	"sync"
	"time"
)

type Cache struct {
	mu  sync.RWMutex
	m   map[string]entry
	ttl time.Duration
}

type entry struct {
	out    Output
	expire time.Time
}

func (c *Cache) Get(key string) (Output, bool) {
	c.mu.RLock()
	e, ok := c.m[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expire) {
		return Output{}, false
	}
	return e.out, true
}

func (c *Cache) Set(key string, out Output) {
	c.mu.Lock()
	c.m[key] = entry{out: out, expire: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.m)
}
