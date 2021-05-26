package controllers

import (
	"github.com/gin-gonic/gin"
	"go-test/di"
	"go-test/utils"
	"time"
)

func PostUserLogin(c *gin.Context) {
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

	token := utils.GenToken()
	if err = di.Sess.HSet(di.Ctx, token, "user_id", user[0]["user_id"]).Err(); err != nil {
		panic(err)
	}
	if err = di.Sess.Expire(di.Ctx, token, 30*3600*time.Second).Err(); err != nil {
		panic(err)
	}
	c.JSON(200, gin.H{"token": token})
}
