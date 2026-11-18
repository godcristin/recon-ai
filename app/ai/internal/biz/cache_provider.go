package biz

import (
	"time"
)

const defaultCacheTTL = 7 * 24 * time.Hour

func NewCache() *Cache {
	return &Cache{m: make(map[string]entry), ttl: defaultCacheTTL}
}
