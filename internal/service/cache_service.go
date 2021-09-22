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
//	@receiver cacheService
//	@param db gorose.IOrm
//	@param table string
//	@param primaryKey string
//	@param id interface{} 整数
//	@return bool
//	@return error
func (cacheService) Set(db gorose.IOrm, table string, primaryKey string, id interface{}) (bool, error) {
	sql := fmt.Sprintf("/*FORCE_MASTER*/ SELECT * FROM %s WHERE %s = %d LIMIT 1", table, primaryKey, id) // 查主库, 避免主从同步延迟的问题
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
	if err = di.CacheRedis().Set(context.Background(), key, dataBytes, time.Hour*24*30).Err(); err != nil {
		zap.L().Error(err.Error())
		return false, err
	}

	return true, nil
}

// Get 获取资源缓存
//	@receiver cacheService
//	@param db gorose.IOrm
//	@param table string
//	@param primaryKey string
//	@param id interface{} 整数
//	@return gorose.Data
//	@return error
func (cacheService) Get(db gorose.IOrm, table string, primaryKey string, id interface{}) (gorose.Data, error) {
	key := fmt.Sprintf(consts.CacheResource, table, id)
	v, err, _ := sg.Do(key, func() (interface{}, error) {
		dataCache, err := di.CacheRedis().Get(context.Background(), key).Result()
		if err != nil {
			if redis.Nil == err { // 缓存不存在
				ok, err := CacheService.Set(db, table, primaryKey, id)
				if err != nil {
					return gorose.Data{}, err
				}
				if ok { // 设置缓存成功
					dataCache, err = di.CacheRedis().Get(context.Background(), key).Result()
					if err != nil {
						zap.L().Error(err.Error())
						return gorose.Data{}, err
					}
				} else { // 记录不存在
					return gorose.Data{}, nil
				}

			} else { // redis异常
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
	if err != nil {
		return gorose.Data{}, err
	}

	return v.(gorose.Data), nil
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
