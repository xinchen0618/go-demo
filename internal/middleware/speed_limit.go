package middleware

import (
	"context"
	"fmt"
	"go-demo/config/consts"
	"go-demo/config/di"
	"go-demo/pkg/ginx"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
	"github.com/spf13/cast"
)

// QpsLimit 限流QPS
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
		var id int64
		if c.GetInt64("userId") > 0 {
			id = c.GetInt64("userId")
		} else if c.GetInt64("adminId") > 0 {
			id = c.GetInt64("adminId")
		}

		key := fmt.Sprintf(consts.SubmitLimit, id, c.Request.Method, c.Request.URL.Path)
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
