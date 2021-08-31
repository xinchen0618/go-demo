// Package gox golang增强方法
package gox

import (
	"fmt"
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
