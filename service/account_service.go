package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"go-test/di"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// CheckUserLogin 登录校验
// 	先校验JWT, 再校验redis白名单
// 	校验不通过方法会向客户端返回4xx错误, 调用方法时捕获到error直接结束业务逻辑即可
//	@param c *gin.Context
//	@return int64
//	@return error
func CheckUserLogin(c *gin.Context) (int64, error) {
	tokenString := c.Request.Header.Get("X-Token")
	if "" == tokenString { // 没有携带token
		c.JSON(401, gin.H{"status": "UserUnauthorized", "message": "用户未登录或登录已过期, 请重新登录"})
		return 0, errors.New("UserUnauthorized")
	}

	// JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(viper.GetString("jwtSecret")), nil
	})
	if err != nil {
		c.JSON(401, gin.H{"status": "UserUnauthorized", "message": "用户未登录或登录已过期, 请重新登录"})
		return 0, errors.New("UserUnauthorized")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// redis白名单校验
		tokenAtoms := strings.Split(tokenString, ".")
		_, err := di.JwtRedis().Get(context.Background(), "jwt:"+claims["jti"].(string)+":"+tokenAtoms[2]).Result()
		if err != nil {
			if "redis: nil" == err.Error() {
				c.JSON(401, gin.H{"status": "UserUnauthorized", "message": "用户未登录或登录已过期, 请重新登录"})
				return 0, errors.New("UserUnauthorized")
			}
			panic(err) // redis服务异常
		}
		userId, err := strconv.ParseInt(claims["jti"].(string), 10, 64)
		if err != nil {
			panic(err)
		}
		return userId, nil

	} else {
		c.JSON(401, gin.H{"status": "UserUnauthorized", "message": "用户未登录或登录已过期, 请重新登录"})
		return 0, errors.New("UserUnauthorized")
	}
}
