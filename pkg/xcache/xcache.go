package xcache

import (
	"context"
	"time"

	"go-demo/pkg/ginx"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

var sg singleflight.Group

// GetOrSet 获取或设置自定义缓存
//  @param cache *redis.Client
//  @param key string
//  @param ttl time.Duration
//  @param f func() (any, error)
//  @return any 返回的是json.Unmarshal的数据
//  @return error
func GetOrSet(cache *redis.Client, key string, ttl time.Duration, f func() (any, error)) (any, error) {
	result, err, _ := sg.Do(key, func() (any, error) {
		var resultCache string
		resultCache, err := cache.Get(context.Background(), key).Result()
		if err != nil {
			if err != redis.Nil {
				zap.L().Error(err.Error())
				return nil, err
			}

			// 缓存不存在
			result, err := f()
			if err != nil {
				return nil, err
			}
			resultBytes, err := json.Marshal(result)
			if err != nil {
				zap.L().Error(err.Error())
				return nil, err
			}
			if err := cache.Set(context.Background(), key, resultBytes, ttl).Err(); err != nil {
				zap.L().Error(err.Error())
				return nil, err
			}
			resultCache = string(resultBytes)
		}

		var resultAny any
		if err := json.Unmarshal([]byte(resultCache), &resultAny); err != nil {
			zap.L().Error(err.Error())
			return nil, err
		}
		return resultAny, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GinCache 获取或者设置业务缓存
//  @param c *gin.Context
//  @param cache *redis.Client
//  @param key string
//  @param ttl time.Duration
//  @param f func() (any, error)
//  @return any 返回的是json.Unmarshal的数据
//  @return error
func GinCache(c *gin.Context, cache *redis.Client, key string, ttl time.Duration, f func() (any, error)) (any, error) {
	result, err, _ := sg.Do(key, func() (any, error) {
		var resultCache string
		resultCache, err := cache.Get(context.Background(), key).Result()
		if err != nil {
			if err != redis.Nil {
				ginx.InternalError(c, err)
				return nil, err
			}

			// 缓存不存在
			result, err := f()
			if err != nil {
				return nil, err
			}
			resultBytes, err := json.Marshal(result)
			if err != nil {
				ginx.InternalError(c, err)
				return nil, err
			}
			if err := cache.Set(context.Background(), key, resultBytes, ttl).Err(); err != nil {
				ginx.InternalError(c, err)
				return nil, err
			}
			resultCache = string(resultBytes)
		}

		var resultAny any
		if err := json.Unmarshal([]byte(resultCache), &resultAny); err != nil {
			ginx.InternalError(c, err)
			return nil, err
		}
		return resultAny, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
