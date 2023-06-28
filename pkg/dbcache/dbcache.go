package dbcache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-demo/pkg/dbx"
	"go-demo/pkg/gox"
	"go-demo/pkg/xcache"

	"github.com/gohouse/gorose/v2"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

const (
	_dbcacheTableRecord     = "dbcache:table:%s:version:%d:key:%d"      // 表记录缓存 dbcache:table:<table_name>:version:<version>:key:<primary_id>
	_dbcacheTablePrimaryKey = "dbcache:table:%s:version:%d:primary_key" // 表主键缓存 dbcache:table:<table_name>:version:<version>:primary_key
	_dbcacheTableVersion    = "dbcache:table:%s:version"                // 表版本 dbcache:table:<table_name>:version
)

var sg singleflight.Group

// set 设置DB缓存
func set(cache *redis.Client, db gorose.IOrm, table string, id any) (bool, error) {
	primaryKey, err := tablePrimaryKey(cache, db, table)
	if err != nil {
		return false, err
	}
	version, err := tableVersion(cache, table)
	if err != nil {
		return false, err
	}

	sql := fmt.Sprintf("SELECT * FROM %s WHERE %s = %d LIMIT 1", table, primaryKey, id)
	data, err := dbx.FetchOne(db, sql)
	if err != nil {
		return false, err
	}
	if len(data) == 0 {
		return false, nil
	}
	dataBytes, err := msgpack.Marshal(data)
	if err != nil {
		return false, err
	}
	key := fmt.Sprintf(_dbcacheTableRecord, table, version, id)
	if err := cache.Set(context.Background(), key, dataBytes, 24*time.Hour).Err(); err != nil {
		zap.L().Error(err.Error())
		return false, err
	}

	return true, nil
}

// Get 获取DB记录返回map并维护缓存
//
//	这里使用msgpack编码缓存数据目的在于解码缓存保持数据类型不变.
//	使用dbcache.Get()或dbcache.Take()方法获取DB记录, 在更新和删除DB记录时, 必须使用dbcache.Update()和dbcache.Delete()方法自动维护缓存, 或dbcache.Expired()手动清除缓存.
func Get(cache *redis.Client, db gorose.IOrm, table string, id any) (map[string]any, error) {
	version, err := tableVersion(cache, table)
	if err != nil {
		return nil, err
	}
	id = cast.ToInt64(id)
	key := fmt.Sprintf(_dbcacheTableRecord, table, version, id)
	v, err, _ := sg.Do(key, func() (any, error) {
		dataCache, err := cache.Get(context.Background(), key).Result()
		switch err {
		case nil:
		case redis.Nil: // 缓存不存在
			ok, err := set(cache, db, table, id)
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
		default: // redis异常
			zap.L().Error(err.Error())
			return nil, err
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
//
//	使用dbcache.Get()或dbcache.Take()方法获取DB记录, 在更新和删除DB记录时, 必须使用dbcache.Update()和dbcache.Delete()方法自动维护缓存, 或dbcache.Expired()手动清除缓存.
//	p 为接收结果的指针.
func Take(p any, cache *redis.Client, db gorose.IOrm, table string, id any) error {
	data, err := Get(cache, db, table, id)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	if err := gox.TypeCast(data, p); err != nil {
		return err
	}

	return nil
}

// Update 更新DB记录并维护缓存
func Update(cache *redis.Client, db gorose.IOrm, table string, data map[string]any, where string, params ...any) (affectedRows int64, err error) {
	// 清除缓存
	primaryKey, err := tablePrimaryKey(cache, db, table)
	if err != nil {
		return 0, err
	}
	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", primaryKey, table, where)
	ids, err := dbx.FetchColumn(db, sql, params...)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
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
func Delete(cache *redis.Client, db gorose.IOrm, table string, where string, params ...any) (affectedRows int64, err error) {
	// 清除缓存
	primaryKey, err := tablePrimaryKey(cache, db, table)
	if err != nil {
		return 0, err
	}
	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", primaryKey, table, where)
	ids, err := dbx.FetchColumn(db, sql, params...)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
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
func Expired(cache *redis.Client, table string, ids ...any) error {
	if len(ids) == 0 {
		return nil
	}

	version, err := tableVersion(cache, table)
	if err != nil {
		return err
	}

	for _, id := range ids {
		id = cast.ToInt64(id)
		key := fmt.Sprintf(_dbcacheTableRecord, table, version, id)
		if err := cache.Del(context.Background(), key).Err(); err != nil {
			zap.L().Error(err.Error())
			return err
		}
	}

	return nil
}

// tablePrimaryKey 获取表主键
func tablePrimaryKey(cache *redis.Client, db gorose.IOrm, table string) (string, error) {
	version, err := tableVersion(cache, table)
	if err != nil {
		return "", err
	}
	key := fmt.Sprintf(_dbcacheTablePrimaryKey, table, version)
	primaryKey, err := xcache.GetOrSet(cache, key, 90*24*time.Hour, func() (any, error) {
		sql := "SHOW COLUMNS FROM " + table
		cols, err := dbx.FetchAll(db, sql)
		if err != nil {
			return "", err
		}
		for _, col := range cols {
			if cast.ToString(col["Key"]) == "PRI" {
				return cast.ToString(col["Field"]), nil
			}
		}

		zap.L().Error("fail to get " + table + " primary key")
		return "", errors.New("fail to get " + table + " primary key")
	})

	return cast.ToString(primaryKey), err
}

// tableVersion 表版本
//
//	更新表版本用于过期与之相关的所有缓存数据.
func tableVersion(cache *redis.Client, table string) (int64, error) {
	key := fmt.Sprintf(_dbcacheTableVersion, table)
	version, err := xcache.GetOrSet(cache, key, 90*24*time.Hour, func() (any, error) {
		return time.Now().Unix(), nil
	})

	return cast.ToInt64(version), err
}
