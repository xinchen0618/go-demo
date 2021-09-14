package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"go-demo/config"
	"go-demo/config/di"
	"go-demo/internal/service"
	"go-demo/pkg/ginx"
	"go-demo/pkg/gox"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gohouse/gorose/v2"
	"go.uber.org/zap"
)

// 这里定义一个空结构体用于为大量的controller方法做分类
type accountController struct {
}

// AccountController 这里仅需结构体零值, router通过controller.XxxController.Xxx的形式引用旗下定义的方法
var AccountController accountController

func (accountController) PostUserLogin(c *gin.Context) { // 先生成JWT, 再记录redis白名单
	jsonBody, err := ginx.GetJsonBody(c, []string{"user_name:用户名:string:+", "password:密码:string:+"})
	if err != nil {
		return
	}

	user, err := di.Db().Table("t_users").Fields("user_id").Where(gorose.Data{"user_name": jsonBody["user_name"], "password": jsonBody["password"]}).First()
	if err != nil {
		ginx.InternalError(c, err)
		return
	}
	if 0 == len(user) {
		c.JSON(400, gin.H{"code": "UserInvalid", "message": "用户名或密码不正确"})
		return
	}

	// JWT
	loginTtl := 86400 * 30 * time.Second // 登录有效时长
	claims := &jwt.StandardClaims{
		Audience:  jsonBody["user_name"].(string),
		ExpiresAt: time.Now().Add(loginTtl).Unix(),
		Id:        strconv.FormatInt(user["user_id"].(int64), 10),
		IssuedAt:  time.Now().Unix(),
		Issuer:    "go-demo:UserLogin",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.GetString("jwt_secret")))
	if err != nil {
		ginx.InternalError(c, err)
		return
	}
	// redis登录白名单
	tokenAtoms := strings.Split(tokenString, ".")
	payload, err := json.Marshal(claims)
	if err != nil {
		ginx.InternalError(c, err)
		return
	}
	key := "jwt:" + claims.Id + ":" + tokenAtoms[2]
	if err = di.JwtRedis().Set(context.Background(), key, payload, loginTtl).Err(); err != nil {
		ginx.InternalError(c, err)
		return
	}

	c.JSON(200, gin.H{"user_id": user["user_id"], "token": tokenString})
}

func (accountController) DeleteUserLogout(c *gin.Context) {
	userId := c.GetInt64("userId")

	// 删除对应redis白名单记录
	tokenString := c.Request.Header.Get("Authorization")[7:]
	tokenAtoms := strings.Split(tokenString, ".")
	key := "jwt:" + strconv.FormatInt(userId, 10) + ":" + tokenAtoms[2]
	if err := di.JwtRedis().Del(context.Background(), key).Err(); err != nil {
		ginx.InternalError(c, err)
		return
	}

	c.JSON(204, gin.H{})
}

func (accountController) GetUsers(c *gin.Context) {
	pageItems, err := ginx.GetPageItems(ginx.PageQuery{
		GinCtx:     c,
		Db:         di.Db(),
		Select:     "user_id,user_name,money,created_at,updated_at",
		From:       "t_users",
		Where:      "user_id > ?",
		BindParams: []interface{}{5},
		OrderBy:    "user_id DESC",
	})
	if err != nil {
		return
	}

	// 多线程读
	wpg := di.WorkerPool().Group()
	for _, item := range pageItems.Items {
		item := item
		wpg.Submit(func() {
			userCounts := service.CacheService.Get(di.Db(), "t_user_counts", "user_id", item["user_id"])
			item["counts"] = 0
			if counts, ok := userCounts["counts"]; ok {
				item["counts"] = counts
			}
		})
	}
	wpg.Wait()

	c.JSON(200, pageItems)
}

func (accountController) GetUsersById(c *gin.Context) {
	userId, err := ginx.FilterParam(c, "用户id", c.Param("user_id"), "+int", false)
	if err != nil {
		return
	}

	user := service.CacheService.Get(di.Db(), "t_users", "user_id", userId.(int64))
	if 0 == len(user) {
		c.JSON(404, gin.H{"code": "UserNotFound", "message": "用户不存在"})
		return
	}

	c.JSON(200, user)
}

func (accountController) PostUsers(c *gin.Context) {
	jsonBody, err := ginx.GetJsonBody(c, []string{"counts:数量:+int:*"})
	if err != nil {
		return
	}

	var counts int64 = 100
	if _, ok := jsonBody["counts"]; ok {
		counts = jsonBody["counts"].(int64)
	}

	for i := int64(0); i < counts; i++ {
		// 多线程写
		di.WorkerPool().Submit(func() {
			db := di.Db()
			if err := db.Begin(); err != nil {
				zap.L().Error(err.Error())
				return
			}

			userName := fmt.Sprintf("U%d", gox.RandInt64(111111, 999999))
			user, err := db.Table("t_users").Fields("user_id").Where(gorose.Data{"user_name": userName}).First()
			if err != nil {
				zap.L().Error(err.Error())
				_ = db.Rollback()
				return
			}
			var userId int64
			if len(user) > 0 { // 记录存在
				userId = user["user_id"].(int64)
			} else { // 记录不存在
				userId, err = db.Table("t_users").Data(gorose.Data{"user_name": userName}).InsertGetId()
				if err != nil {
					zap.L().Error(err.Error())
					_ = db.Rollback()
					return
				}
			}
			sql := "INSERT INTO t_user_counts(user_id,counts) VALUES(?,?) ON DUPLICATE KEY UPDATE counts = counts + 1"
			if _, err = db.Execute(sql, userId, gox.RandInt64(1, 9)); err != nil {
				zap.L().Error(err.Error())
				_ = db.Rollback()
				return
			}

			if err := db.Commit(); err != nil {
				zap.L().Error(err.Error())
				_ = db.Rollback()
				return
			}
		})
	}

	c.JSON(201, gin.H{"counts": counts})
}
