// Package di 服务注入
package di

import (
	"fmt"
	"sync"

	"go-demo/config"

	"github.com/redis/go-redis/v9"
)

/******************** 缓存 redis ********************/
var (
	cacheRedis     *redis.Client
	cacheRedisOnce sync.Once
)

// CacheRedis 缓存 redis 实例
//
//	删除缓存数据不会引发业务错误
func CacheRedis() *redis.Client {
	cacheRedisOnce.Do(func() {
		cacheRedis = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", config.GetString("redis_host"), config.GetInt("redis_port")),
			Password: config.GetString("redis_auth"),
			DB:       config.GetInt("redis_index_cache"),
		})
	})

	return cacheRedis
}

/******************** 存储 redis ********************/
var (
	storageRedis     *redis.Client
	storageRedisOnce sync.Once
)

// StorageRedis 存储 redis 实例
//
//	删除存储数据会引发业务错误
func StorageRedis() *redis.Client {
	storageRedisOnce.Do(func() {
		storageRedis = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", config.GetString("redis_host"), config.GetInt("redis_port")),
			Password: config.GetString("redis_auth"),
			DB:       config.GetInt("redis_index_storage"),
		})
	})

	return storageRedis
}

/******************** jwt redis ********************/
var (
	jwtRedis     *redis.Client
	jwtRedisOnce sync.Once
)

// JWTRedis JWT redis 实例
func JWTRedis() *redis.Client {
	jwtRedisOnce.Do(func() {
		jwtRedis = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", config.GetString("redis_host"), config.GetInt("redis_port")),
			Password: config.GetString("redis_auth"),
			DB:       config.GetInt("redis_index_jwt"),
		})
	})

	return jwtRedis
}
