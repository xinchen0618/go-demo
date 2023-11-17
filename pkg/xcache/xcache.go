// Package xcache 自定义缓存操作函数
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
//	缓存使用 JSON 编码.
//
//	p 为接收缓存数据的指针.
//	f() 返回的 any 为需要缓存的数据, 返回 error 时数据不缓存.
//
//	函数返回 error 表示取数据失败.
func GetOrSet(p any, cache *redis.Client, key string, ttl time.Duration, f func() (any, error)) error {
	if _, err, _ := sg.Do(key, func() (any, error) {
		var resultBytes []byte
		// 取数据
		resultCache, err := cache.Get(context.Background(), key).Result()
		switch err {
		case nil: // 正常拿到缓存
			resultBytes = []byte(resultCache)
		case redis.Nil: // 缓存不存在
			result, err := f()
			if err != nil {
				return nil, err
			}
			resultBytes, err = json.Marshal(result)
			if err != nil {
				zap.L().Error(err.Error())
				return nil, err
			}
			if err := cache.Set(context.Background(), key, resultBytes, ttl).Err(); err != nil {
				zap.L().Error(err.Error())
				return nil, err
			}
		default: // redis 异常
			zap.L().Error(err.Error())
			return nil, err
		}

		// 返回数据
		if err := json.Unmarshal(resultBytes, &p); err != nil {
			zap.L().Error(err.Error())
			return nil, err
		}

		return nil, nil
	}); err != nil {
		return err
	}
	return nil
}

// GinCache 获取或者设置 Gin 缓存
//
//	缓存使用 JSON 编码.
//	函数中出现 error 会向客户端输出错误. f() 中可调用 c.
//
//	p 为接收缓存数据的指针.
//	f() 返回的 any 为需要缓存的数据, 返回 error 时数据不缓存.
//
//	函数返回 error 表示取数据失败.
func GinCache(p any, c *gin.Context, cache *redis.Client, key string, ttl time.Duration, f func() (any, error)) error {
	if _, err, _ := sg.Do(key, func() (any, error) {
		var resultBytes []byte
		// 取数据
		resultCache, err := cache.Get(context.Background(), key).Result()
		switch err {
		case nil: // 正常拿到缓存
			resultBytes = []byte(resultCache)
		case redis.Nil: // 缓存不存在
			result, err := f()
			if err != nil {
				return nil, err
			}
			resultBytes, err = json.Marshal(result)
			if err != nil {
				ginx.InternalError(c, err)
				return nil, err
			}
			if err := cache.Set(context.Background(), key, resultBytes, ttl).Err(); err != nil {
				ginx.InternalError(c, err)
				return nil, err
			}
		default: // redis异常
			ginx.InternalError(c, err)
			return nil, err
		}

		if err := json.Unmarshal(resultBytes, &p); err != nil {
			ginx.InternalError(c, err)
			return nil, err
		}
		return nil, nil
	}); err != nil {
		return err
	}
	return nil
}
