package gox

import (
	"math/rand"
	"time"
)

// RandInt64 生成区间随机数
//
//	双闭区间[min, max]
//	@param min int64
//	@param max int64
//	@return int64
func RandInt64(min, max int64) int64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return r.Int63n(max-min+1) + min
}
