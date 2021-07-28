package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go-demo/config"
	"go-demo/di"
	"time"

	"github.com/gohouse/gorose/v2"
)

type cacheService struct {
}

var CacheService *cacheService

// Set 为资源设置缓存
//	@receiver *cacheService
//	@param db gorose.IOrm
//	@param table string
//	@param primaryKey string
//	@param id int64
//	@return bool
func (*cacheService) Set(db gorose.IOrm, table string, primaryKey string, id int64) bool {
	sql := fmt.Sprintf("/*FORCE_MASTER*/ SELECT * FROM %s WHERE %s = %d LIMIT 1", table, primaryKey, id) // 查主库, 解决主从同步延迟的问题
	data, err := db.Query(sql)
	if err != nil {
		di.Logger().Error(err.Error())
		return false
	}
	if 0 == len(data) {
		return false
	}
	dataBytes, err := json.Marshal(data[0])
	if err != nil {
		di.Logger().Error(err.Error())
		return false
	}
	key := fmt.Sprintf(config.RedisResourceInfo, table, id)
	err = di.CacheRedis().Set(context.Background(), key, dataBytes, time.Hour*24*30).Err()
	if err != nil {
		di.Logger().Error(err.Error())
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
//	@param id int64
//	@return gorose.Data
func (*cacheService) Get(db gorose.IOrm, table string, primaryKey string, id int64) gorose.Data {
	key := fmt.Sprintf(config.RedisResourceInfo, table, id)
	dataCache, err := di.CacheRedis().Get(context.Background(), key).Result()
	if err != nil {
		if "redis: nil" == err.Error() { // 缓存不存在
			if CacheService.Set(db, table, primaryKey, id) {
				return CacheService.Get(db, table, primaryKey, id)
			} else {
				return gorose.Data{}
			}
		}
	}
	var dataMap gorose.Data // 缓存存在
	err = json.Unmarshal([]byte(dataCache), &dataMap)
	if err != nil {
		di.Logger().Error(err.Error())
		return gorose.Data{}
	}

	return dataMap
}

// Delete 删除资源缓存
//	@receiver *cacheService
//	@param table string
//	@param id int64
//	@return bool
func (*cacheService) Delete(table string, id int64) bool {
	key := fmt.Sprintf(config.RedisResourceInfo, table, id)
	err := di.CacheRedis().Del(context.Background(), key).Err()
	if err != nil {
		di.Logger().Error(err.Error())
		return false
	}

	return true
}
