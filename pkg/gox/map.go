// Package gox golang增强方法
package gox

import (
	"encoding/json"

	"go.uber.org/zap"
)

// TypeCast 类型转换
//  @param o any 原数据
//  @param p any 目标结果的指针
//  @return error
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
