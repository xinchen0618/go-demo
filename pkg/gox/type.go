// Package gox Golang 增强函数
package gox

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

// CopyViaJSON 将数据从一个类型复制到另一个类型
//
//	支持多级数据结构, 字段值是不可跨类型的, 比如, float 复制到 string 是不可以的.
//
//	source 为原数据, destination 为接收目标结果的指针. 涉及到结构体, 并且字段名不一致时, 通过 json tag 指定.
func CopyViaJSON(source any, destination any) error {
	interim, err := json.Marshal(source)
	if err != nil {
		zap.L().Error(err.Error())
		return err
	}
	if err := json.Unmarshal(interim, destination); err != nil {
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
