package service

import (
	"context"
	"fmt"
	"go-demo/config/consts"
	"go-demo/config/di"
	"go-demo/pkg/gox"
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
	dataBytes, err := gox.GobEncode(data[0])
	if err != nil {
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
		if err := gox.GobDecode([]byte(dataCache), &dataMap); err != nil {
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
