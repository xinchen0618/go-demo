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

type cacheService struct {
}

var CacheService cacheService

var sg singleflight.Group

// Set 为资源设置缓存
//	@receiver *cacheService
//	@param db gorose.IOrm
//	@param table string
//	@param primaryKey string
//	@param id interface{} 整数
//	@return bool
func (cacheService) Set(db gorose.IOrm, table string, primaryKey string, id interface{}) bool {
	sql := fmt.Sprintf("/*FORCE_MASTER*/ SELECT * FROM %s WHERE %s = %d LIMIT 1", table, primaryKey, id) // 查主库, 避免主从同步延迟的问题
	data, err := db.Query(sql)
	if err != nil {
		zap.L().Error(err.Error())
		return false
	}
	if 0 == len(data) {
		return false
	}
	dataBytes, err := json.Marshal(data[0])
	if err != nil {
		zap.L().Error(err.Error())
		return false
	}
	key := fmt.Sprintf(consts.CacheResource, table, id)
	if err = di.CacheRedis().Set(context.Background(), key, dataBytes, time.Hour*24*30).Err(); err != nil {
		zap.L().Error(err.Error())
		return false
	}

	return true
}

// Get 获取资源缓存
//	缓存不存在时会设置缓存
//	@receiver *cacheService
//	@param db gorose.IOrm
//	@param table string
//	@param primaryKey string
//	@param id interface{} 整数
//	@return gorose.Data
func (cacheService) Get(db gorose.IOrm, table string, primaryKey string, id interface{}) gorose.Data {
	key := fmt.Sprintf(consts.CacheResource, table, id)
	v, _, _ := sg.Do(key, func() (interface{}, error) {
		dataCache, err := di.CacheRedis().Get(context.Background(), key).Result()
		if err != nil {
			if redis.Nil == err { // 缓存不存在
				if CacheService.Set(db, table, primaryKey, id) {
					dataCache, err = di.CacheRedis().Get(context.Background(), key).Result()
					if err != nil {
						zap.L().Error(err.Error())
						return gorose.Data{}, err
					}
				} else {
					return gorose.Data{}, nil
				}
			} else {
				zap.L().Error(err.Error())
				return gorose.Data{}, err
			}
		}

		var dataMap gorose.Data
		if err := json.Unmarshal([]byte(dataCache), &dataMap); err != nil {
			zap.L().Error(err.Error())
			return gorose.Data{}, err
		}

		return dataMap, nil
	})

	return v.(gorose.Data)
}

// Delete 删除资源缓存
//	@receiver *cacheService
//	@param table string
//	@param id interface{} 整数
//	@return bool
func (cacheService) Delete(table string, id interface{}) bool {
	key := fmt.Sprintf(consts.CacheResource, table, id)
	if err := di.CacheRedis().Del(context.Background(), key).Err(); err != nil {
		zap.L().Error(err.Error())
		return false
	}

	return true
}
