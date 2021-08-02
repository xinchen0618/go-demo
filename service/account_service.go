package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"go-demo/di"
	"go-demo/util"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

// 这里定义一个空结构体用于为大量的service方法做分类
type accountService struct {
}

// AccountService 这里不需要实例化, 外部通过service.XxxService.Xxx()的形式调用旗下定义的方法
var AccountService *accountService

// CheckUserLogin 登录校验
// 	先校验JWT, 再校验redis白名单
// 	校验不通过方法会向客户端返回4xx错误, 调用方法时捕获到error直接结束业务逻辑即可
//	@receiver *accountService
//	@param c *gin.Context
//	@return int64
//	@return error
func (*accountService) CheckUserLogin(c *gin.Context) (int64, error) {
	tokenString := c.Request.Header.Get("Authorization") // Authorization: Bearer <token>
	if !strings.HasPrefix(tokenString, "Bearer ") {
		c.JSON(401, gin.H{"status": "UserUnauthorized", "message": "用户未登录或登录已过期, 请重新登录"})
		return 0, errors.New("UserUnauthorized")
	}
	tokenString = tokenString[7:]

	// JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(viper.GetString("jwtSecret")), nil
	})
	if err != nil {
		c.JSON(401, gin.H{"status": "UserUnauthorized", "message": "用户未登录或登录已过期, 请重新登录"})
		return 0, errors.New("UserUnauthorized")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid { // redis白名单校验
		tokenAtoms := strings.Split(tokenString, ".")
		key := "jwt:" + claims["jti"].(string) + ":" + tokenAtoms[2]
		_, err := di.JwtRedis().Get(context.Background(), key).Result()
		if err != nil {
			if redis.Nil == err {
				c.JSON(401, gin.H{"status": "UserUnauthorized", "message": "用户未登录或登录已过期, 请重新登录"})
				return 0, errors.New("UserUnauthorized")
			}
			util.InternalError(c, err) // redis服务异常
			return 0, errors.New("InternalError")
		}
		userId, err := strconv.ParseInt(claims["jti"].(string), 10, 64)
		if err != nil {
			util.InternalError(c, err)
			return 0, errors.New("InternalError")
		}
		return userId, nil

	} else {
		c.JSON(401, gin.H{"status": "UserUnauthorized", "message": "用户未登录或登录已过期, 请重新登录"})
		return 0, errors.New("UserUnauthorized")
	}
}
