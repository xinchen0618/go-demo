// Package gox Golang 增强函数
package gox

import (
	"fmt"

	"go.uber.org/zap"
)

// Go 开启一个 Goroutine
//
//	这里会对 Goroutine 进行 recover 包装, 避免因为野生 Goroutine 报 panic 导致主线程崩溃退出.
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
