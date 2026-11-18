package biz

import (
	"sync"
	"time"
)

// Cache 归因结果缓存：相同差异上下文在 TTL 内直接复用，避免重复调用模型。
// 进程内实现，够用且无额外依赖；后续可替换为 Redis。
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
