// Package gox Golang 增强函数
package gox

import (
	"sync"
	"sync/atomic"
)

// Once 一个功能更加强大的 Once
type Once struct {
	m    sync.Mutex
	done uint32
}

// Do 传入的函数 f 有返回值 error，如果初始化失败，需要返回失败的 error, Do 方法会把这个 error 返回给调用者
func (o *Once) Do(f func() error) error {
	if atomic.LoadUint32(&o.done) == 1 { // fast path
		return nil
	}
	return o.slowDo(f)
}

// 如果还没有初始化
func (o *Once) slowDo(f func() error) error {
	o.m.Lock()
	defer o.m.Unlock()
	var err error
	if o.done == 0 { // 双检查，还没有初始化
		err = f()
		if err == nil { // 初始化成功才将标记置为已初始化
			atomic.StoreUint32(&o.done, 1)
		}
	}
	return err
}
