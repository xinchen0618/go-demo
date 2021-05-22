package controllers

import (
	"github.com/gin-gonic/gin"
	"go-test/di"
	"go-test/services"
	"time"
)

// UserLogin 绑定为json
type UserLogin struct {
	UserName string `json:"user_name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func PostUserLogin(c *gin.Context) {
	var json UserLogin
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(500, gin.H{"status": "EmptyParam", "message": "用户名和密码不得为空"})
		return
	}

	sql := "SELECT user_id FROM t_users WHERE user_name = ? AND password = ? LIMIT 1"
	res, err := di.Db.Query(sql, json.UserName, json.Password)
	if err != nil {
		panic(err)
	}
	if 0 == len(res) {
		c.JSON(500, gin.H{"status": "InvalidUser", "message": "用户名或密码不正确"})
		return
	}

	token := services.GenToken()
	err = di.Sess.HSet(di.Ctx, token, "user_id", res[0]["user_id"]).Err()
	if err != nil {
		panic(err)
	}
	di.Sess.Expire(di.Ctx, token, 30*3600*time.Second)

	c.JSON(201, gin.H{
		"token":   token,
		"user_id": res[0]["user_id"],
	})
}
