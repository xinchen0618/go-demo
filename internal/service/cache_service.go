package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go-demo/config/consts"
	"go-demo/config/di"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gohouse/gorose/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

type cacheService struct{}

var (
	CacheService cacheService
	cacheSg      singleflight.Group
)

// Set 设置资源缓存
//	@receiver cacheService
//	@param db gorose.IOrm
//	@param table string
//	@param primaryKey string
//	@param id interface{} 整数
//	@return bool
//	@return error
func (cacheService) Set(db gorose.IOrm, table string, primaryKey string, id interface{}) (bool, error) {
	sql := fmt.Sprintf("SELECT * FROM %s WHERE %s = %d LIMIT 1", table, primaryKey, id)
	data, err := db.Query(sql)
	if err != nil {
		zap.L().Error(err.Error())
		return false, err
	}
	if 0 == len(data) {
		return false, nil
	}
	dataBytes, err := json.Marshal(data[0])
	if err != nil {
		zap.L().Error(err.Error())
		return false, err
	}
	key := fmt.Sprintf(consts.CacheResource, table, id)
	if err := di.CacheRedis().Set(context.Background(), key, dataBytes, time.Hour*24).Err(); err != nil {
		zap.L().Error(err.Error())
		return false, err
	}

	return true, nil
}

// Get 获取资源缓存
//	方法返回的是json.Unmarshal的数据
//	@receiver cacheService
//	@param db gorose.IOrm
//	@param table string
//	@param primaryKey string
//	@param id interface{} 整数
//	@return map[string]interface{}
//	@return error
func (cacheService) Get(db gorose.IOrm, table string, primaryKey string, id interface{}) (map[string]interface{}, error) {
	key := fmt.Sprintf(consts.CacheResource, table, id)
	v, err, _ := cacheSg.Do(key, func() (interface{}, error) {
		dataCache, err := di.CacheRedis().Get(context.Background(), key).Result()
		if err != nil {
			if err != redis.Nil { // redis异常
				zap.L().Error(err.Error())
				return map[string]interface{}{}, err
			}

			// 缓存不存在
			ok, err := CacheService.Set(db, table, primaryKey, id)
			if err != nil {
				return map[string]interface{}{}, err
			}
			if !ok { // 记录不存在
				return map[string]interface{}{}, nil
			}
			// 设置缓存成功
			dataCache, err = di.CacheRedis().Get(context.Background(), key).Result()
			if err != nil {
				zap.L().Error(err.Error())
				return map[string]interface{}{}, err
			}
		}
		var dataMap map[string]interface{}
		if err := json.Unmarshal([]byte(dataCache), &dataMap); err != nil {
			zap.L().Error(err.Error())
			return map[string]interface{}{}, err
		}
		return dataMap, nil
	})
	if err != nil {
		return map[string]interface{}{}, err
	}

	return v.(map[string]interface{}), nil
}

// Delete 删除资源缓存
//	@receiver *cacheService
//	@param table string
//	@param id interface{} 整数
//	@return error
func (cacheService) Delete(table string, id interface{}) error {
	key := fmt.Sprintf(consts.CacheResource, table, id)
	if err := di.CacheRedis().Del(context.Background(), key).Err(); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}

// GetOrSet 获取或者设置业务缓存
//	方法返回的是json.Unmarshal的数据
//	@receiver cacheService
//	@param key string
//	@param ttl int64 缓存时长(秒)
//	@param f func() (interface{}, error)
//	@return interface{}
//	@return error
func (cacheService) GetOrSet(key string, ttl int64, f func() (interface{}, error)) (interface{}, error) {
	result, err, _ := cacheSg.Do(key, func() (interface{}, error) {
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
			if err := di.CacheRedis().Set(context.Background(), key, resultBytes, time.Second*time.Duration(ttl)).Err(); err != nil {
				zap.L().Error(err.Error())
				return nil, err
			}
			resultCache = string(resultBytes)
		}

		var resultMap interface{}
		if err := json.Unmarshal([]byte(resultCache), &resultMap); err != nil {
			zap.L().Error(err.Error())
			return nil, err
		}
		return resultMap, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
