package gox

import (
	"math/rand"
	"time"
)

// RandInt64 生成区间随机数
//	左闭右开区间[min, max)
//	@param min int64
//	@param max int64
//	@return int64
func RandInt64(min, max int64) int64 {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(max-min) + min
}

// RandString 生成随机字符串
//  @param n int
//  @return string
func RandString(n int) string {
	letterBytes := "23456789abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
