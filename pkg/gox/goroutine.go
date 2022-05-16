package gox

import (
	"fmt"

	"go.uber.org/zap"
)

// Go 开启一个goroutine
//  这里会对goroutine进行recover包装, 避免因为野生goroutine报panic导致主协程崩溃退出
//  goroutine应优先使用WorkerPool
//  @param f func()
func Go(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				zap.L().Error(fmt.Sprint(r))
			}
		}()
		f()
	}()
}
