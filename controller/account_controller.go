package controller

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"go-test/di"
	"go-test/service"
	"go-test/util"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func PostUserLogin(c *gin.Context) { // 先生成JWT, 再记录redis白名单
	jsonBody, err := util.GetJsonBody(c, []string{"user_name:用户名:string:+", "password:密码:string:+"})
	if err != nil {
		return
	}

	user, err := di.Db().Query("SELECT user_id FROM t_users WHERE user_name = ? AND password = ? LIMIT 1",
		jsonBody["user_name"], jsonBody["password"])
	if err != nil {
		panic(err)
	}
	if 0 == len(user) {
		c.JSON(400, gin.H{"status": "InvalidUser", "message": "用户名或密码不正确"})
		return
	}

	// JWT
	loginTtl := 86400 * 30 * time.Second // 登录有效时长
	claims := &jwt.StandardClaims{
		Audience:  jsonBody["user_name"].(string),
		ExpiresAt: time.Now().Add(loginTtl).Unix(),
		Id:        strconv.FormatInt(user[0]["user_id"].(int64), 10),
		IssuedAt:  time.Now().Unix(),
		Issuer:    "go-test:UserLogin",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(viper.GetString("jwtSecret")))
	if err != nil {
		panic(err)
	}
	// redis登录白名单
	tokenAtoms := strings.Split(tokenString, ".")
	payload, err := json.Marshal(claims)
	if err != nil {
		panic(err)
	}
	if err = di.JwtRedis().Set(di.Ctx(), "jwt:"+claims.Id+":"+tokenAtoms[2], payload, loginTtl).Err(); err != nil {
		panic(err)
	}
	c.JSON(200, gin.H{"user_id": user[0]["user_id"], "token": tokenString})
}

func DeleteUserLogout(c *gin.Context) {
	// 登录校验
	userId, err := service.CheckUserLogin(c)
	if err != nil {
		return
	}

	// 删除对应redis白名单记录
	tokenString := c.Request.Header.Get("X-Token")
	tokenAtoms := strings.Split(tokenString, ".")
	if err := di.JwtRedis().Del(di.Ctx(), "jwt:"+strconv.FormatInt(userId, 10)+":"+tokenAtoms[2]).Err(); err != nil {
		panic(err)
	}

	c.JSON(204, gin.H{})
}

func GetUsers(c *gin.Context) {
	// 登录校验
	if _, err := service.CheckUserLogin(c); err != nil {
		return
	}

	res, err := util.GetPageItems(map[string]interface{}{
		"ginContext": c,
		"db":         di.Db(),
		"select":     "u.user_id,u.user_name,u.money,u.created_at,u.updated_at,uc.counts",
		"from":       "t_users AS u JOIN t_user_counts AS uc ON u.user_id = uc.user_id",
		"where":      "u.user_id > ?",
		"bindParams": []interface{}{5},
		"orderBy":    "user_id DESC",
	})
	if err != nil {
		return
	}
	c.JSON(200, res)
}
