package biz

import (
	"time"
)

// defaultCacheTTL 归因结果缓存时间
const defaultCacheTTL = 7 * 24 * time.Hour

// NewCache 进程内缓存，默认 TTL 7 天
func NewCache() *Cache {
	return &Cache{m: make(map[string]entry), ttl: defaultCacheTTL}
}
