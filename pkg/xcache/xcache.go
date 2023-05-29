package xcache

import (
	"context"
	"time"

	"go-demo/pkg/ginx"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

var sg singleflight.Group

// GetOrSet 获取或设置自定义缓存
//
//	缓存数据使用json编解码
//	@param cache *redis.Client
//	@param key string
//	@param ttl time.Duration
//	@param f func() (any, error)
//	@return any 返回的是json.Unmarshal的数据
//	@return error
func GetOrSet(cache *redis.Client, key string, ttl time.Duration, f func() (any, error)) (any, error) {
	result, err, _ := sg.Do(key, func() (any, error) {
		var resultCache string
		resultCache, err := cache.Get(context.Background(), key).Result()
		switch err {
		case nil:
		case redis.Nil: // 缓存不存在
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
		default: // redis异常
			zap.L().Error(err.Error())
			return nil, err
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
//
//	发生错误会向客户端输出500错误
//	@param c *gin.Context
//	@param cache *redis.Client
//	@param key string
//	@param ttl time.Duration
//	@param f func() (any, error)
//	@return any 返回的是json.Unmarshal的数据
//	@return error
func GinCache(c *gin.Context, cache *redis.Client, key string, ttl time.Duration, f func() (any, error)) (any, error) {
	result, err, _ := sg.Do(key, func() (any, error) {
		var resultCache string
		resultCache, err := cache.Get(context.Background(), key).Result()
		switch err {
		case nil:
		case redis.Nil: // 缓存不存在
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
		default: // redis异常
			ginx.InternalError(c, err)
			return nil, err
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
