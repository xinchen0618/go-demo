package services

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go-test/di"
	"strings"
	"time"
)

// CheckUserLogin 登录权限
// 先校验JWT, 再校验redis白名单
// 校验不通过方法会向客户端返回4xx错误, 调用方法时捕获到error直接结束业务逻辑即可
func CheckUserLogin(c *gin.Context) (userId int64, resErr error) {
	tokenString := c.Request.Header.Get("X-Token")
	if "" == tokenString { // 没有携带token
		c.JSON(401, gin.H{"status": "UserUnauthorized", "message": "用户未登录, 请登录"})
		resErr = errors.New("UserUnauthorized")
		return
	}

	// JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(viper.GetString("jwtSecret")), nil
	})
	if err != nil { // 非法token
		c.JSON(401, gin.H{"status": "UserUnauthorized", "message": "用户未登录, 请登录"})
		resErr = errors.New("UserUnauthorized")
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if int64(claims["expired_at"].(float64)) <= time.Now().Unix() {
			c.JSON(401, gin.H{"status": "UserUnauthorized", "message": "登录已过期, 请重新登录"})
			resErr = errors.New("UserUnauthorized")
			return
		}
		// redis校验
		tokenAtoms := strings.Split(tokenString, ".")
		_, err := di.JwtRedis.Get(di.Ctx, tokenAtoms[2]).Result()
		if err != nil {
			if "redis: nil" == err.Error() {
				c.JSON(401, gin.H{"status": "UserUnauthorized", "message": "登录已过期, 请重新登录"})
				resErr = errors.New("UserUnauthorized")
				return
			}
			panic(err) // redis服务异常
		}
		userId = int64(claims["user_id"].(float64))
		return

	} else {
		c.JSON(401, gin.H{"status": "UserUnauthorized", "message": "用户未登录, 请登录"})
		resErr = errors.New("UserUnauthorized")
		return
	}
}
