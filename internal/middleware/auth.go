package middleware

import (
	"strings"

	"go-demo/config/consts"
	"go-demo/internal/service"
	"go-demo/pkg/ginx"

	"github.com/gin-gonic/gin"
)

// UserJwtParse 用户JWT解析
//
//	解析成功会将userId存入gin上下文
//	@return gin.HandlerFunc
func UserJwtParse() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("Authorization") // Authorization: Bearer <token>
		if !strings.HasPrefix(tokenString, "Bearer ") {
			c.Next()
			return
		}
		tokenString = tokenString[7:]

		// JWT校验
		userId, err := service.Auth.JwtCheck(consts.UserJwt, tokenString)
		if err != nil {
			ginx.InternalError(c, nil)
			return
		}
		if 0 == userId {
			c.Next()
			return
		}

		c.Set("userId", userId) // 后续的处理函数可以用过c.GetInt64("userId")来获取当前请求的用户信息
		c.Next()
	}
}

// UserAuth 用户鉴权
//
//	@return gin.HandlerFunc
func UserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if 0 == c.GetInt64("userId") {
			ginx.Error(c, 401, "UserUnauthorized", "您未登录或登录已过期, 请重新登录")
			return
		}
		c.Next()
	}
}
