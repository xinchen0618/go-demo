// Package xcache 自定义缓存操作函数
// 缓存使用 JSON 编码.
package xcache

import (
	"context"
	"errors"
	"time"

	"go-demo/pkg/ginx"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

var sg singleflight.Group

type OnceReq struct {
	Value  any          // 接收缓存数据的指针
	GinCtx *gin.Context // 选填，用于 Gin 向客户端输出 4xx/500 错误, 调用时捕获到 error 直接结束业务逻辑即可
	Cache  *redis.Client
	Key    string
	TTL    time.Duration       // 默认 1 小时
	Do     func() (any, error) // 返回的 any 为需要缓存的数据, 返回 error 时数据不缓存.
}

// Once 获取或设置缓存
//
//	函数返回 error 表示取数据失败.
func Once(req OnceReq) error {
	value, err, _ := sg.Do(req.Key, func() (any, error) {
		// 取数据
		valueCache, err := req.Cache.Get(context.Background(), req.Key).Result()
		switch {
		case err == nil: // 正常拿到缓存
			return []byte(valueCache), nil
		case errors.Is(err, redis.Nil): // 缓存不存在
			value, err := req.Do()
			if err != nil {
				return nil, err
			}
			valueBytes, err := json.Marshal(value)
			if err != nil {
				zap.L().Error(err.Error())
				if req.GinCtx != nil {
					ginx.InternalError(req.GinCtx, nil)
				}
				return nil, err
			}
			ttl := req.TTL
			if ttl == 0 {
				ttl = time.Hour
			}
			if err := req.Cache.Set(context.Background(), req.Key, valueBytes, ttl).Err(); err != nil {
				zap.L().Error(err.Error())
				if req.GinCtx != nil {
					ginx.InternalError(req.GinCtx, nil)
				}
				return nil, err
			}
			return valueBytes, nil
		default: // redis 异常
			zap.L().Error(err.Error())
			if req.GinCtx != nil {
				ginx.InternalError(req.GinCtx, nil)
			}
			return nil, err
		}
	})
	if err != nil {
		return err
	}

	// 返回数据
	if err := json.Unmarshal(value.([]byte), req.Value); err != nil {
		zap.L().Error(err.Error())
		if req.GinCtx != nil {
			ginx.InternalError(req.GinCtx, nil)
		}
		return err
	}

	return nil
}

type GetReq struct {
	GinCtx *gin.Context // 选填，用于 Gin 向客户端输出 4xx/500 错误, 调用时捕获到 error 直接结束业务逻辑即可
	Cache  *redis.Client
	Key    string
	Value  any // 接收结构的指针
}

// Get 从缓存中获取数据
func Get(req GetReq) error {
	valueCache, err := req.Cache.Get(context.Background(), req.Key).Result()
	switch {
	case err == nil: // 正常拿到缓存
		if err := json.Unmarshal([]byte(valueCache), req.Value); err != nil {
			zap.L().Error(err.Error())
			if req.GinCtx != nil {
				ginx.InternalError(req.GinCtx, nil)
			}
			return err
		}
		return nil
	case errors.Is(err, redis.Nil): // 缓存不存在
		return nil
	default:
		zap.L().Error(err.Error())
		if req.GinCtx != nil {
			ginx.InternalError(req.GinCtx, nil)
		}
		return err
	}
}

type SetReq struct {
	GinCtx *gin.Context // 选填，用于 Gin 向客户端输出 4xx/500 错误, 调用时捕获到 error 直接结束业务逻辑即可
	Cache  *redis.Client
	Key    string
	Value  any // 需要缓存的数据
	TTL    time.Duration
}

// Set 设置缓存
func Set(req SetReq) error {
	value, err := json.Marshal(req.Value)
	if err != nil {
		zap.L().Error(err.Error())
		if req.GinCtx != nil {
			ginx.InternalError(req.GinCtx, nil)
		}
		return err
	}
	if err := req.Cache.Set(context.Background(), req.Key, value, req.TTL).Err(); err != nil {
		zap.L().Error(err.Error())
		if req.GinCtx != nil {
			ginx.InternalError(req.GinCtx, nil)
		}
		return err
	}
	return nil
}

type DelReq struct {
	GinCtx *gin.Context // 选填， 用于 Gin 向客户端输出 4xx/500 错误, 调用时捕获到 error 直接结束业务逻辑即可
	Cache  *redis.Client
	Key    string
}

// Del 删除缓存
func Del(req DelReq) error {
	if err := req.Cache.Del(context.Background(), req.Key).Err(); err != nil {
		zap.L().Error(err.Error())
		if req.GinCtx != nil {
			ginx.InternalError(req.GinCtx, nil)
		}
		return err
	}
	return nil
}
