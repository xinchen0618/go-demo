package xcache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"
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
			resultCache, err = jsoniter.MarshalToString(result)
			if err != nil {
				zap.L().Error(err.Error())
				return nil, err
			}
			if err := cache.Set(context.Background(), key, resultCache, ttl).Err(); err != nil {
				zap.L().Error(err.Error())
				return nil, err
			}
		}

		var resultAny any
		if err := jsoniter.UnmarshalFromString(resultCache, &resultAny); err != nil {
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
