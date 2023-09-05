// Package gox Golang 增强函数
package gox

import "math"

// Round 四舍五入
//
//	ROUND_HALF_UP 模式实现. 返回将 val 根据指定精度 precision（十进制小数点后数字的数目）进行四舍五入的结果. precision 也可以是负数或零.
func Round(val float64, precision int) float64 {
	p := math.Pow10(precision)
	return math.Floor(val*p+0.5) / p
}

// Fib 斐波那契数列
func Fib(n int) int {
	x, y := 0, 1
	for i := 0; i < n; i++ {
		x, y = y, x+y
	}
	return x
}
