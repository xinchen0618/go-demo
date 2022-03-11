// Package gox golang增强方法
package gox

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// PrintMap 向console打印map
//	@param m map[string]interface{}
func PrintMap(m map[string]interface{}) {
	fmt.Println("{")
	for k, v := range m {
		fmt.Printf("\t%s: %T %#v\n", k, v, v)
	}
	fmt.Println("}")
}

// Map2struct map转struct
//  @param m map[string]interface{}
//  @param sp interface{} 接收结果结构体的指针
//  @return error
func Map2struct(m map[string]interface{}, sp interface{}) error {
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
