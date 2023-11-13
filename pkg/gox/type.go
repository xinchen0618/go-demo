// Package gox Golang 增强函数
package gox

import (
	"errors"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

// Copy 将数据从一个类型复制到另一个类型
//
//	支持多级数据结构, 但字段值不可跨类型, 比如, float 复制到 string 是不可以的.
//
//	o 为原数据, p 为接收目标结果的指针. 涉及到结构体, 并且字段名不一致时, 通过 json tag 指定.
func Copy(o any, p any) error {
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

// WideCopy 可跨值类型的将数据从一个类型复制到另一个类型
//
//	仅支持一级数据结构, 字段值可跨类型, 比如, float 复制到 string 是可以的.
//
// o 为原数据, p 为接收目标结果的指针. 涉及到结构体, 并且字段名不一致时, 通过 json tag 指定.
func WideCopy(o any, p any) error {
	om := map[string]any{}
	if err := Copy(o, &om); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	pm := map[string]any{}
	if err := Copy(p, &pm); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	for k, v := range pm {
		var err error
		switch v.(type) {
		case float64:
			pm[k], err = cast.ToFloat64E(om[k])
			if err != nil {
				zap.L().Error(err.Error())
				return fmt.Errorf("value cast error: %s, %w"+k, err)
			}
		case string:
			pm[k], err = cast.ToStringE(om[k])
			if err != nil {
				zap.L().Error(err.Error())
				return fmt.Errorf("value cast error: %s, %w"+k, err)
			}
		case bool:
			pm[k], err = cast.ToBoolE(om[k])
			if err != nil {
				zap.L().Error(err.Error())
				return fmt.Errorf("value cast error: %s, %w"+k, err)
			}
		default:
			return errors.New("unsupported value type: " + k)
		}
	}

	if err := Copy(pm, p); err != nil {
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
