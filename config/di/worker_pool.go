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
		workerPool = pond.New(config.GetInt("worker_pool"), 0, pond.PanicHandler(func(i interface{}) {
			zap.L().Error(fmt.Sprint(i))
		}))
	})

	return workerPool
}

// WorkerPoolSeparate 独享Goroutine池
//  @param maxWorkers int
//  @return *pond.WorkerPool
func WorkerPoolSeparate(maxWorkers int) *pond.WorkerPool {
	return pond.New(maxWorkers, 0, pond.PanicHandler(func(i interface{}) {
		zap.L().Error(fmt.Sprint(i))
	}))
}
