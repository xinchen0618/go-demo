// Package di 服务注入
package di

import (
	"fmt"
	"sync"

	"go-demo/config"

	"github.com/alitto/pond"
	"go.uber.org/zap"
)

var (
	workerPool *pond.WorkerPool
	wpOnce     sync.Once
)

// Pool 公共 Goroutine 池
func Pool() *pond.WorkerPool {
	wpOnce.Do(func() {
		workerPool = pond.New(config.GetInt("worker_pool"), 0, pond.PanicHandler(func(a any) {
			zap.L().Error(fmt.Sprint(a))
		}))
	})

	return workerPool
}

// PoolSeparate 独享 Goroutine 池
//
//	一次请求提交大量数据, 使用独享 Goroutine 池起限流作用.
func PoolSeparate(maxWorkers int) *pond.WorkerPool {
	return pond.New(maxWorkers, 0, pond.PanicHandler(func(a any) {
		zap.L().Error(fmt.Sprint(a))
	}))
}
