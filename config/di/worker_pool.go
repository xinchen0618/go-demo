package di

import (
	"fmt"
	"go-demo/config"
	"sync"

	"github.com/alitto/pond"
	"go.uber.org/zap"
)

var (
	workerPool *pond.WorkerPool
	wpOnce     sync.Once
)

func WorkerPool() *pond.WorkerPool {
	wpOnce.Do(func() {
		workerPool = pond.New(config.GetInt("worker_pool"), config.GetInt("worker_pool"), pond.PanicHandler(func(i interface{}) {
			zap.L().Error(fmt.Sprint(i))
		}))
	})

	return workerPool
}
