package middleware

import (
	"context"
	"fmt"
	"time"

	"go-demo/config/consts"
	"go-demo/config/di"
	"go-demo/pkg/ginx"
	"go-demo/pkg/gox"

	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
	"github.com/spf13/cast"
)

// QpsLimit QPS限流
func QpsLimit(qps int) gin.HandlerFunc {
	quantum := cast.ToInt64(qps)
	bucket := ratelimit.NewBucketWithQuantum(time.Second, quantum, quantum)
	return func(c *gin.Context) {
		if bucket.TakeAvailable(1) < 1 {
			ginx.Error(c, 429, "TooManyRequests", "服务繁忙, 请稍后重试")
			return
		}
		c.Next()
	}
}

// SubmitLimit 提交频率限制
//
//	主要用于防重.
func SubmitLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		var uid string // md5(id+method+path)
		// 优先取用户id作为唯一标识, 如果没有则取ip+agent作为唯一标识
		if c.GetInt64("userID") > 0 {
			uid = cast.ToString(c.GetInt64("userID"))
		} else if c.GetInt64("adminID") > 0 {
			uid = cast.ToString(c.GetInt64("adminID"))
		} else {
			uid = c.ClientIP() + ":" + c.Request.UserAgent()
		}
		uid = gox.Md5(uid + ":" + c.Request.Method + ":" + c.Request.URL.Path)
		key := fmt.Sprintf(consts.SubmitLimit, uid)
		ok, err := di.CacheRedis().SetNX(context.Background(), key, 1, 2*time.Second).Result() // 2秒/次
		if err != nil {
			ginx.InternalError(c, err)
			return
		}
		if !ok {
			ginx.Error(c, 429, "SubmitLimit", "手快了, 请稍后~~")
			return
		}
		c.Next()
	}
}
