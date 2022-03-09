package di

import (
	"fmt"
	"sync"

	"go-demo/config"

	"github.com/go-redis/redis/v8"
)

// 缓存redis
var (
	cacheRedis     *redis.Client
	cacheRedisOnce sync.Once
)

func CacheRedis() *redis.Client {
	cacheRedisOnce.Do(func() {
		cacheRedis = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", config.Get("redis_host"), config.Get("redis_port")),
			Password: config.GetString("redis_auth"),
			DB:       config.GetInt("redis_index_cache"),
		})
	})

	return cacheRedis
}

// jwt redis
var (
	jwtRedis     *redis.Client
	jwtRedisOnce sync.Once
)

func JwtRedis() *redis.Client {
	jwtRedisOnce.Do(func() {
		jwtRedis = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", config.Get("redis_host"), config.Get("redis_port")),
			Password: config.GetString("redis_auth"),
			DB:       config.GetInt("redis_index_jwt"),
		})
	})

	return jwtRedis
}
