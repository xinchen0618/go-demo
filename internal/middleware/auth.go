// Package middleware Gin 中间件
package middleware

import (
	"context"
	"fmt"
	"math"
	"strings"

	"go-demo/config"
	"go-demo/config/di"
	"go-demo/internal/consts"
	"go-demo/pkg/ginx"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

// JWTParse JWT 解析
//
//	解析成功会将 userID 或者 adminID 存入 Gin 上下文.
func JWTParse(userType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := lo.Substring(c.Request.Header.Get("Authorization"), 7, math.MaxUint) // Authorization: Bearer <token>
		// JWT校验
		jwtToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			return []byte(config.GetString("jwt_secret")), nil
		})
		if err != nil { // token 无效
			c.Next()
			return
		}
		claims, ok := jwtToken.Claims.(jwt.MapClaims)
		if !ok || !jwtToken.Valid { // token 秘钥/时间等校验未通过
			c.Next()
			return
		}
		// 白名单校验
		tokenAtoms := strings.Split(tokenString, ".")
		key := fmt.Sprintf(consts.JWTLogin, userType, claims["jti"], tokenAtoms[2])
		if n, err := di.JWTRedis().Exists(context.Background(), key).Result(); err != nil {
			di.Logger().Error(err.Error())
			c.Next()
			return
		} else if n == 0 { // 不在白名单内
			c.Next()
			return
		}
		// id 存入 Gin 上下文
		id := cast.ToInt64(claims["jti"])
		if userType == consts.UserJWT {
			c.Set("userID", id) // 后续的处理函数可以用过 c.GetInt64("userID") 来获取当前请求的用户 id
		} else if userType == consts.AdminJWT {
			c.Set("adminID", id) // 后续的处理函数可以用过 c.GetInt64("adminID") 来获取当前请求的用户 id
		}
		c.Next()
	}
}

// UserAuth 用户鉴权
//
//	登录即可.
func UserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetInt64("userID") == 0 {
			ginx.Error(c, 401, "UserUnauthorized", "您未登录或登录已过期, 请重新登录")
			return
		}
		c.Next()
	}
}
