package di

import (
	"fmt"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

// 缓存redis
var (
	cacheRedis     *redis.Client
	cacheRedisOnce sync.Once
)

func CacheRedis() *redis.Client {
	cacheRedisOnce.Do(func() {
		cacheRedis = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", viper.Get("redis.host"), viper.Get("redis.port")),
			Password: viper.GetString("redis.auth"),
			DB:       viper.GetInt("redis.index.cache"),
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
			Addr:     fmt.Sprintf("%s:%d", viper.Get("redis.host"), viper.Get("redis.port")),
			Password: viper.GetString("redis.auth"),
			DB:       viper.GetInt("redis.index.jwt"),
		})
	})

	return jwtRedis
}
