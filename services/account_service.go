package services

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go-test/di"
)

// CheckAuth 校验权限
// 校验不通过方法会向客户端返回4xx错误, 调用方法时捕获到error直接结束业务逻辑即可
func CheckAuth(c *gin.Context) error {
	token := c.Request.Header.Get("X-Token")
	if "" == token {
		c.JSON(401, gin.H{"status": "UserUnauthorized", "message": "用户未登录或登录已过期, 请重新登录"})

		return errors.New("UserUnauthorized")
	}

	_, err := di.Sess.HGet(di.Ctx, token, "user_id").Result()
	if err != nil {
		if "redis: nil" == err.Error() {
			c.JSON(401, gin.H{"status": "UserUnauthorized", "message": "用户未登录或登录已过期, 请重新登录"})
			return errors.New("UserUnauthorized")
		}
		panic(err) // redis服务异常
	}

	return nil
}
