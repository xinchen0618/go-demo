// Package gox golang增强方法
package gox

import (
	"fmt"
	"go-demo/config/di"
)

// InSlice 元素是否在切片中
//	支持int64/string/float64类型
//	@param needle interface{}
//	@param haystack interface{}
//	@return bool
func InSlice(needle interface{}, haystack interface{}) bool {
	switch key := needle.(type) {
	case int64:
		for _, item := range haystack.([]int64) {
			if key == item {
				return true
			}
		}
	case string:
		for _, item := range haystack.([]string) {
			if key == item {
				return true
			}
		}
	case float64:
		for _, item := range haystack.([]float64) {
			if key == item {
				return true
			}
		}
	default:
		di.Logger().Error(fmt.Sprintf("func InSlice no handled type: %T", needle))
		return false
	}

	return false
}
