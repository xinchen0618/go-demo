package dbcache

import (
	"context"
	"fmt"
	"time"

	"go-demo/pkg/dbx"
	"go-demo/pkg/gox"

	"github.com/go-redis/redis/v8"
	"github.com/gohouse/gorose/v2"
	"github.com/spf13/cast"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

const dbcacheKey = "dbcache:%s:%d" // DB缓存 dbcache:<table_name>:<primary_id>

var sg singleflight.Group

// set 设置DB缓存
//  @param cache *redis.Client
//  @param db gorose.IOrm
//  @param table string
//  @param primaryKey string
//  @param id any
//  @return bool
//  @return error
func set(cache *redis.Client, db gorose.IOrm, table string, primaryKey string, id any) (bool, error) {
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
	key := fmt.Sprintf(dbcacheKey, table, id)
	if err := cache.Set(context.Background(), key, dataBytes, 24*time.Hour).Err(); err != nil {
		zap.L().Error(err.Error())
		return false, err
	}

	return true, nil
}

// Get 获取DB记录返回map并维护缓存
//  使用dbcache.Get()或dbcache.Take()方法获取DB记录, 在更新和删除DB记录时, 必须使用dbcache.Update()和dbcache.Delete()方法自动维护缓存, 或dbcache.Expired()手动清除缓存
//  @param cache *redis.Client
//  @param db gorose.IOrm
//  @param table string
//  @param primaryKey string
//  @param id any
//  @return map[string]any
//  @return error
func Get(cache *redis.Client, db gorose.IOrm, table string, primaryKey string, id any) (map[string]any, error) {
	id = cast.ToInt64(id)
	key := fmt.Sprintf(dbcacheKey, table, id)
	v, err, _ := sg.Do(key, func() (any, error) {
		dataCache, err := cache.Get(context.Background(), key).Result()
		if err != nil {
			if err != redis.Nil { // redis异常
				zap.L().Error(err.Error())
				return nil, err
			}

			// 缓存不存在
			ok, err := set(cache, db, table, primaryKey, id)
			if err != nil {
				return nil, err
			}
			if !ok { // 记录不存在
				return map[string]any{}, nil
			}
			// 设置缓存成功
			dataCache, err = cache.Get(context.Background(), key).Result()
			if err != nil {
				zap.L().Error(err.Error())
				return nil, err
			}
		}
		var dataMap map[string]any
		if err := msgpack.Unmarshal([]byte(dataCache), &dataMap); err != nil {
			return nil, err
		}
		return dataMap, nil
	})
	if err != nil {
		return nil, err
	}

	return v.(map[string]any), nil
}

// Take 获取DB记录至struct并维护缓存
//  使用dbcache.Get()或dbcache.Take()方法获取DB记录, 在更新和删除DB记录时, 必须使用dbcache.Update()和dbcache.Delete()方法自动维护缓存, 或dbcache.Expired()手动清除缓存
//  @param p any 接收结果的指针
//  @param cache *redis.Client
//  @param db gorose.IOrm
//  @param table string
//  @param primaryKey string
//  @param id any
//  @return error
func Take(p any, cache *redis.Client, db gorose.IOrm, table string, primaryKey string, id any) error {
	data, err := Get(cache, db, table, primaryKey, id)
	if err != nil {
		return err
	}
	if 0 == len(data) {
		return nil
	}
	if err := gox.TypeCast(data, p); err != nil {
		return err
	}

	return nil
}

// Update 更新DB记录并维护缓存
//  @param cache *redis.Client
//  @param db gorose.IOrm
//  @param table string
//  @param primaryKey string
//  @param data map[string]any
//  @param where string
//  @param params ...any
//  @return affectedRows int64
//  @return err error
func Update(cache *redis.Client, db gorose.IOrm, table string, primaryKey string, data map[string]any, where string, params ...any) (affectedRows int64, err error) {
	// 清除缓存
	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", primaryKey, table, where)
	ids, err := dbx.FetchColumn(db, sql, params...)
	if err != nil {
		return 0, err
	}
	if 0 == len(ids) {
		return 0, nil
	}
	if err := Expired(cache, table, ids...); err != nil {
		return 0, err
	}

	// 更新数据
	affectedRows, err = dbx.Update(db, table, data, where, params...)
	if err != nil {
		return 0, err
	}

	return affectedRows, nil
}

// Delete 删除DB记录并维护缓存
//  @param cache *redis.Client
//  @param db gorose.IOrm
//  @param table string
//  @param primaryKey string
//  @param where string
//  @param params ...any
//  @return affectedRows int64
//  @return err error
func Delete(cache *redis.Client, db gorose.IOrm, table string, primaryKey string, where string, params ...any) (affectedRows int64, err error) {
	// 清除缓存
	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", primaryKey, table, where)
	ids, err := dbx.FetchColumn(db, sql, params...)
	if err != nil {
		return 0, err
	}
	if 0 == len(ids) {
		return 0, nil
	}
	if err := Expired(cache, table, ids...); err != nil {
		return 0, err
	}

	// 删除数据
	affectedRows, err = dbx.Delete(db, table, where, params...)
	if err != nil {
		return 0, err
	}

	return affectedRows, nil
}

// Expired 过期缓存
//  @param cache *redis.Client
//  @param table string
//  @param ids ...any
//  @return error
func Expired(cache *redis.Client, table string, ids ...any) error {
	if 0 == len(ids) {
		return nil
	}

	for _, id := range ids {
		id = cast.ToInt64(id)
		key := fmt.Sprintf(dbcacheKey, table, id)
		if err := cache.Del(context.Background(), key).Err(); err != nil {
			zap.L().Error(err.Error())
			return err
		}
	}

	return nil
}
