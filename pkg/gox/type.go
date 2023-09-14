// Package gox Golang 增强函数
package gox

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

// Cast 类型转换
//
//	o 为原数据, p 为接收目标结果的指针, 目标结果为结构体时 json tag 是必须的.
func Cast(o any, p any) error {
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

// Decimal 浮点转 decimal
//
//	超过精度会四舍五入, 不足精度会补0. 比如, Decimal(1.2, 2), 返回 "1.20"
func Decimal(f float64, precision int) string {
	return fmt.Sprintf("%."+cast.ToString(precision)+"f", Round(f, precision))
}
