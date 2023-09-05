// Package gox Golang 增强函数
package gox

import (
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

// TypeCast 类型转换
//
//	o 为原数据, p 为接收目标结果的指针, 目标结果为结构体时 json tag 是必须的.
func TypeCast(o any, p any) error {
	b, err := json.Marshal(o)
	if err != nil {
		zap.L().Error(err.Error())
		return err
	}
	if err := json.Unmarshal(b, p); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}
