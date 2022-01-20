package middleware

import (
	"context"
	"fmt"
	"go-demo/config/consts"
	"go-demo/config/di"
	"go-demo/pkg/ginx"
	"time"

	"github.com/gin-gonic/gin"
)

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
		ok, err := di.CacheRedis().SetNX(context.Background(), key, 1, 2*time.Second).Result()
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
