package middleware

import (
	"context"
	"fmt"
	"go-demo/config"
	"go-demo/config/consts"
	"go-demo/config/di"
	"go-demo/pkg/ginx"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/cast"
)

// UserAuth 用户登录
//	@return gin.HandlerFunc
func UserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("Authorization") // Authorization: Bearer <token>
		if !strings.HasPrefix(tokenString, "Bearer ") {
			ginx.Error(c, 401, "UserUnauthorized", "用户未登录或登录已过期, 请重新登录")
			return
		}
		tokenString = tokenString[7:]

		// JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.GetString("jwt_secret")), nil
		})
		if err != nil {
			ginx.Error(c, 401, "UserUnauthorized", "用户未登录或登录已过期, 请重新登录")
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid { // redis白名单校验
			tokenAtoms := strings.Split(tokenString, ".")
			key := fmt.Sprintf(consts.JwtUserLogin, claims["jti"], tokenAtoms[2])
			if _, err := di.JwtRedis().Get(context.Background(), key).Result(); err != nil {
				if err != redis.Nil {
					ginx.InternalError(c, err) // redis服务异常
					return
				}
				ginx.Error(c, 401, "UserUnauthorized", "用户未登录或登录已过期, 请重新登录")
				return
			}
			c.Set("userId", cast.ToInt64(claims["jti"])) // 后续的处理函数可以用过c.GetInt64("userId")来获取当前请求的用户信息
			c.Next()

		} else {
			ginx.Error(c, 401, "UserUnauthorized", "用户未登录或登录已过期, 请重新登录")
			return
		}
	}
}
