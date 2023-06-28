package middleware

import (
	"strings"

	"go-demo/config/consts"
	"go-demo/internal/service"
	"go-demo/pkg/ginx"

	"github.com/gin-gonic/gin"
)

// UserJWTParse 用户JWT解析
//
//	解析成功会将userId存入gin上下文.
func UserJWTParse() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("Authorization") // Authorization: Bearer <token>
		if !strings.HasPrefix(tokenString, "Bearer ") {
			c.Next()
			return
		}
		tokenString = tokenString[7:]

		// JWT校验
		userID, err := service.Auth.JWTCheck(consts.UserJWT, tokenString)
		if err != nil {
			ginx.InternalError(c, nil)
			return
		}
		if userID == 0 {
			c.Next()
			return
		}

		c.Set("userID", userID) // 后续的处理函数可以用过c.GetInt64("userID")来获取当前请求的用户信息
		c.Next()
	}
}

// UserAuth 用户鉴权
func UserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetInt64("userID") == 0 {
			ginx.Error(c, 401, "UserUnauthorized", "您未登录或登录已过期, 请重新登录")
			return
		}
		c.Next()
	}
}
