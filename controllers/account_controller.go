package controllers

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go-test/di"
	"go-test/services"
	"go-test/utils"
	"strings"
	"time"
)

func PostUserLogin(c *gin.Context) { // 先生成JWT, 再记录redis白名单
	jsonBody, err := utils.GetJsonBody(c, []string{"user_name:用户名:string:+", "password:密码:string:+"})
	if err != nil {
		return
	}

	user, err := di.Db.Query("SELECT user_id FROM t_users WHERE user_name = ? AND password = ? LIMIT 1",
		jsonBody["user_name"], jsonBody["password"])
	if err != nil {
		panic(err)
	}
	if 0 == len(user) {
		c.JSON(400, gin.H{"status": "InvalidUser", "message": "用户名或密码不正确"})
		return
	}

	// JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user[0]["user_id"],
		"expired_at": time.Now().Unix() + 86400*30,
	})
	tokenString, err := token.SignedString([]byte(viper.GetString("jwtSecret")))
	if err != nil {
		panic(err)
	}
	// redis存储
	tokenAtoms := strings.Split(tokenString, ".")
	if err = di.JwtRedis.Set(di.Ctx, tokenAtoms[2], user[0]["user_id"], 30*86400*time.Second).Err(); err != nil {
		panic(err)
	}
	c.JSON(200, gin.H{"user_id": user[0]["user_id"], "token": tokenString})
}

func DeleteUserLogout(c *gin.Context) {
	// 登录校验
	if _, err := services.CheckUserLogin(c); err != nil {
		return
	}

	// 删除对应redis白名单记录
	tokenString := c.Request.Header.Get("X-Token")
	tokenAtoms := strings.Split(tokenString, ".")
	if err := di.JwtRedis.Del(di.Ctx, tokenAtoms[2]).Err(); err != nil {
		panic(err)
	}

	c.JSON(204, gin.H{})
}
