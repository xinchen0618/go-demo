// Package gox golang增强方法
package gox

import (
	"encoding/json"

	"go.uber.org/zap"
)

// Map2struct map转struct
//  @param m map[string]any
//  @param sp any 接收结果结构体的指针
//  @return error
func Map2struct(m map[string]any, sp any) error {
	b, err := json.Marshal(m)
	if err != nil {
		zap.L().Error(err.Error())
		return err
	}
	if err := json.Unmarshal(b, sp); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}
