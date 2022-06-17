package gincache

import (
	"context"
	"time"

	"go-demo/pkg/ginx"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"
	"golang.org/x/sync/singleflight"
)

var sg singleflight.Group

// GetOrSet 获取或者设置业务缓存
//  @param c *gin.Context
//  @param cache *redis.Client
//  @param key string
//  @param ttl time.Duration
//  @param f func() (any, error)
//  @return any 返回的是json.Unmarshal的数据
//  @return error
func GetOrSet(c *gin.Context, cache *redis.Client, key string, ttl time.Duration, f func() (any, error)) (any, error) {
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
			resultCache, err = jsoniter.MarshalToString(result)
			if err != nil {
				ginx.InternalError(c, err)
				return nil, err
			}
			if err := cache.Set(context.Background(), key, resultCache, ttl).Err(); err != nil {
				ginx.InternalError(c, err)
				return nil, err
			}
		}

		var resultAny any
		if err := jsoniter.UnmarshalFromString(resultCache, &resultAny); err != nil {
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
