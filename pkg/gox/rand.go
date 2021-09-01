package gox

import (
	"math/rand"
	"time"
)

// RandInt64 生成区间随机数
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
