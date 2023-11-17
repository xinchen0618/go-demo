// Package di 服务注入
package di

import (
	"sync"

	"github.com/go-redis/cache/v9"
	"github.com/goccy/go-json"
)

var (
	goCache     *cache.Cache
	goCacheOnce sync.Once
)

func Cache() *cache.Cache {
	goCacheOnce.Do(func() {
		goCache = cache.New(&cache.Options{
			Redis: CacheRedis(),
			Marshal: func(i any) ([]byte, error) {
				result, err := json.Marshal(i)
				if err != nil {
					Logger().Error(err.Error())
					return nil, err
				}
				return result, nil
			},
			Unmarshal: func(bytes []byte, i any) error {
				if err := json.Unmarshal(bytes, i); err != nil {
					Logger().Error(err.Error())
					return err
				}
				return nil
			},
		})
	})

	return goCache
}
