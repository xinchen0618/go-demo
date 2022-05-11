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
//	@param qps int
//	@return gin.HandlerFunc
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
//  主要用于防重
//  @return gin.HandlerFunc
func SubmitLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先取用户id作为唯一标识, 如果没有则取ip+agent作为唯一标识
		var key string
		if c.GetInt64("userId") > 0 {
			key = cast.ToString(c.GetInt64("userId"))
		} else if c.GetInt64("adminId") > 0 {
			key = cast.ToString(c.GetInt64("adminId"))
		} else {
			key = c.ClientIP() + ":" + c.Request.UserAgent()
		}
		key += ":" + c.Request.Method + ":" + c.Request.URL.Path
		key, err := gox.Md5x(key)
		if err != nil {
			ginx.InternalError(c, nil)
			return
		}
		key = fmt.Sprintf(consts.SubmitLimit, key)
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
