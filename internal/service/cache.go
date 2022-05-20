package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-demo/config/consts"
	"go-demo/config/di"
	"go-demo/pkg/dbx"
	"go-demo/pkg/ginx"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gohouse/gorose/v2"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

var cacheSg singleflight.Group

type dbCache struct{}

var DbCache dbCache

// set 设置资源缓存
//	@receiver dbCache
//	@param db gorose.IOrm
//	@param table string
//	@param primaryKey string
//	@param id any 整数
//	@return bool
//	@return error
func (dbCache) set(db gorose.IOrm, table string, primaryKey string, id any) (bool, error) {
	sql := fmt.Sprintf("SELECT * FROM %s WHERE %s = %d LIMIT 1", table, primaryKey, id)
	data, err := dbx.FetchOne(db, sql)
	if err != nil {
		return false, err
	}
	if 0 == len(data) {
		return false, nil
	}
	dataBytes, err := msgpack.Marshal(data)
	if err != nil {
		return false, err
	}
	key := fmt.Sprintf(consts.CacheDb, table, id)
	if err := di.CacheRedis().Set(context.Background(), key, dataBytes, 24*time.Hour).Err(); err != nil {
		zap.L().Error(err.Error())
		return false, err
	}

	return true, nil
}

// Get 获取资源缓存
//  缓存不存在时会建立
//	@receiver dbCache
//	@param db gorose.IOrm
//	@param table string
//	@param primaryKey string
//	@param id any 整数
//	@return map[string]any
//	@return error
func (dbCache) Get(db gorose.IOrm, table string, primaryKey string, id any) (map[string]any, error) {
	key := fmt.Sprintf(consts.CacheDb, table, id)
	v, err, _ := cacheSg.Do(key, func() (any, error) {
		dataCache, err := di.CacheRedis().Get(context.Background(), key).Result()
		if err != nil {
			if err != redis.Nil { // redis异常
				zap.L().Error(err.Error())
				return map[string]any{}, err
			}

			// 缓存不存在
			ok, err := DbCache.set(db, table, primaryKey, id)
			if err != nil {
				return map[string]any{}, err
			}
			if !ok { // 记录不存在
				return map[string]any{}, nil
			}
			// 设置缓存成功
			dataCache, err = di.CacheRedis().Get(context.Background(), key).Result()
			if err != nil {
				zap.L().Error(err.Error())
				return map[string]any{}, err
			}
		}
		var dataMap map[string]any
		if err := msgpack.Unmarshal([]byte(dataCache), &dataMap); err != nil {
			return map[string]any{}, err
		}
		return dataMap, nil
	})
	if err != nil {
		return map[string]any{}, err
	}

	return v.(map[string]any), nil
}

// Delete 删除资源缓存
//  @receiver dbCache
//  @param table string
//  @param ids ...any 整数
//  @return error
func (dbCache) Delete(table string, ids ...any) error {
	if 0 == len(ids) {
		return nil
	}

	for _, id := range ids {
		key := fmt.Sprintf(consts.CacheDb, table, id)
		if err := di.CacheRedis().Del(context.Background(), key).Err(); err != nil {
			zap.L().Error(err.Error())
			return err
		}
	}

	return nil
}

type ginCache struct{}

var GinCache ginCache

// GetOrSet 获取或者设置业务缓存
//	@receiver ginCache
//	@param key string
//	@param ttl time.Duration 缓存时长
//	@param f func() (any, error)
//	@return any 返回的是json.Unmarshal的数据
//	@return error
func (ginCache) GetOrSet(c *gin.Context, key string, ttl time.Duration, f func() (any, error)) (any, error) {
	result, err, _ := cacheSg.Do(key, func() (any, error) {
		var resultCache string
		resultCache, err := di.CacheRedis().Get(context.Background(), key).Result()
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
			if err := di.CacheRedis().Set(context.Background(), key, resultBytes, ttl).Err(); err != nil {
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

type cache struct{}

var Cache cache

// GetOrSet 获取或设置自定义缓存
//  @receiver cache
//  @param key string
//  @param ttl time.Duration
//  @param f func() (any, error)
//  @return any 返回的是json.Unmarshal的数据
//  @return error
func (cache) GetOrSet(key string, ttl time.Duration, f func() (any, error)) (any, error) {
	result, err, _ := cacheSg.Do(key, func() (any, error) {
		var resultCache string
		resultCache, err := di.CacheRedis().Get(context.Background(), key).Result()
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
			if err := di.CacheRedis().Set(context.Background(), key, resultBytes, ttl).Err(); err != nil {
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
