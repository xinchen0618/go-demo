package middleware

import (
	"context"
	"fmt"
	"go-demo/config"
	"go-demo/config/consts"
	"go-demo/config/di"
	"go-demo/pkg/ginx"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// UserAuth 用户登录
//	@return gin.HandlerFunc
func UserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("Authorization") // Authorization: Bearer <token>
		if !strings.HasPrefix(tokenString, "Bearer ") {
			c.JSON(401, gin.H{"code": "UserUnauthorized", "message": "用户未登录或登录已过期, 请重新登录"})
			c.Abort()
			return
		}
		tokenString = tokenString[7:]

		// JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.GetString("jwt_secret")), nil
		})
		if err != nil {
			c.JSON(401, gin.H{"code": "UserUnauthorized", "message": "用户未登录或登录已过期, 请重新登录"})
			c.Abort()
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid { // redis白名单校验
			tokenAtoms := strings.Split(tokenString, ".")
			key := fmt.Sprintf(consts.JwtUserLogin, claims["jti"], tokenAtoms[2])
			if _, err := di.JwtRedis().Get(context.Background(), key).Result(); err != nil {
				if redis.Nil == err {
					c.JSON(401, gin.H{"code": "UserUnauthorized", "message": "用户未登录或登录已过期, 请重新登录"})
					c.Abort()
					return
				}
				ginx.InternalError(c, err) // redis服务异常
				c.Abort()
				return
			}
			userId, err := strconv.ParseInt(claims["jti"].(string), 10, 64)
			if err != nil {
				ginx.InternalError(c, err)
				c.Abort()
				return
			}
			c.Set("userId", userId) // 后续的处理函数可以用过c.GetInt64("userId")来获取当前请求的用户信息
			c.Next()

		} else {
			c.JSON(401, gin.H{"code": "UserUnauthorized", "message": "用户未登录或登录已过期, 请重新登录"})
			c.Abort()
			return
		}
	}
}
