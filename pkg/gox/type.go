// Package gox Golang 增强函数
package gox

import (
	"fmt"

	"github.com/spf13/cast"
)

// Decimal 浮点转 decimal
//
//	超过精度会四舍五入, 不足精度会补0. 比如, Decimal(1.2, 2), 返回 "1.20"
func Decimal(f float64, precision int) string {
	return fmt.Sprintf("%."+cast.ToString(precision)+"f", Round(f, precision))
}
