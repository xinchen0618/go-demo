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

// WorkerPool 公共Goroutine池
//  @return *pond.WorkerPool
func WorkerPool() *pond.WorkerPool {
	wpOnce.Do(func() {
		workerPool = pond.New(config.GetInt("worker_pool"), 1, pond.PanicHandler(func(a any) {
			zap.L().Error(fmt.Sprint(a))
		}))
	})

	return workerPool
}

// WorkerPoolSeparate 独享Goroutine池
//  一次请求提交大量数据, 使用独享Goroutine池起限流作用
//  @param maxWorkers int
//  @return *pond.WorkerPool
func WorkerPoolSeparate(maxWorkers int) *pond.WorkerPool {
	return pond.New(maxWorkers, 1, pond.PanicHandler(func(a any) {
		zap.L().Error(fmt.Sprint(a))
	}))
}
