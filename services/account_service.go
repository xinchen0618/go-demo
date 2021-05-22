package services

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go-test/di"
)

// CheckAuth 校验权限
func CheckAuth(c *gin.Context) error {
	token := c.Request.Header.Get("X-Token")
	if "" == token {
		c.JSON(401, gin.H{"status": "UserUnauthorized", "message": "用户未登录或登录已过期, 请重新登录"})

		return errors.New("UserUnauthorized")
	}

	_, err := di.Sess.HGet(di.Ctx, token, "user_id").Result()
	if err != nil {
		c.JSON(401, gin.H{"status": "UserUnauthorized", "message": "用户未登录或登录已过期, 请重新登录"})

		return errors.New("UserUnauthorized")
	}

	return nil
}
