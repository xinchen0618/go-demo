package di

import (
	"sync"

	"github.com/go-redis/cache/v9"
)

var (
	goRedisCache     *cache.Cache
	goRedisCacheOnce sync.Once
)

// Cache go-redis cache
func Cache() *cache.Cache {
	goRedisCacheOnce.Do(func() {
		goRedisCache = cache.New(&cache.Options{
			Redis: CacheRedis(),
		})
	})

	return goRedisCache
}
