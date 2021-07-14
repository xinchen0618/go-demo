package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-test/config"
	"go-test/di"
	"go-test/service"
	"go-test/util"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gohouse/gorose/v2"
	"github.com/spf13/viper"
)

type AccountController struct {
}

func (accountController *AccountController) PostUserLogin(c *gin.Context) { // 先生成JWT, 再记录redis白名单
	jsonBody, err := util.GetJsonBody(c, []string{"user_name:用户名:string:+", "password:密码:string:+"})
	if err != nil {
		return
	}

	user, err := di.Db().Table("t_users").Fields("user_id").Where(gorose.Data{"user_name": jsonBody["user_name"], "password": jsonBody["password"]}).First()
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
		Id:        strconv.FormatInt(user["user_id"].(int64), 10),
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
	if err = di.JwtRedis().Set(context.Background(), "jwt:"+claims.Id+":"+tokenAtoms[2], payload, loginTtl).Err(); err != nil {
		panic(err)
	}

	c.JSON(200, gin.H{"user_id": user["user_id"], "token": tokenString})
}

func (accountController *AccountController) DeleteUserLogout(c *gin.Context) {
	// 登录校验
	userId, err := service.CheckUserLogin(c)
	if err != nil {
		return
	}

	// 删除对应redis白名单记录
	tokenString := c.Request.Header.Get("X-Token")
	tokenAtoms := strings.Split(tokenString, ".")
	if err := di.JwtRedis().Del(context.Background(), "jwt:"+strconv.FormatInt(userId, 10)+":"+tokenAtoms[2]).Err(); err != nil {
		panic(err)
	}

	c.JSON(204, gin.H{})
}

func (accountController *AccountController) GetUsers(c *gin.Context) {
	// 登录校验
	//if _, err := service.CheckUserLogin(c); err != nil {
	//	return
	//}

	result, err := util.GetPageItems(map[string]interface{}{
		"ginContext": c,
		"db":         di.Db(),
		"select":     "user_id,user_name,money,created_at,updated_at",
		"from":       "t_users",
		"where":      "user_id > ?",
		"bindParams": []interface{}{5},
		"orderBy":    "user_id DESC",
	})
	if err != nil {
		return
	}

	// 多线程读
	var wg sync.WaitGroup
	for _, item := range result["items"].([]gorose.Data) {
		wg.Add(1)
		go func(item gorose.Data) {
			defer wg.Done()

			userCounts, err := di.Db().Table("t_user_counts").Fields("counts").Where(gorose.Data{"user_id": item["user_id"]}).First()
			if err != nil {
				log.Println(err)
				return
			}
			item["counts"] = userCounts["counts"]
		}(item)
	}
	wg.Wait()

	c.JSON(200, result)
}

func (accountController *AccountController) GetUsersById(c *gin.Context) {
	userId, err := util.FilterParam(c, "用户id", c.Param("user_id"), "+int", false)
	if err != nil {
		return
	}

	// cache
	key := fmt.Sprintf(config.RedisUser, userId)
	userStr, err := di.CacheRedis().Get(context.Background(), key).Result()
	if err != nil && "redis: nil" != err.Error() {
		panic(err)
	}
	if userStr != "" {
		var user gorose.Data
		if err = json.Unmarshal([]byte(userStr), &user); err != nil {
			panic(err)
		}

		c.JSON(200, user)
		return
	}

	user, err := di.Db().Table("t_users").Where(gorose.Data{"user_id": userId}).First()
	if err != nil {
		panic(err)
	}
	if 0 == len(user) {
		c.JSON(404, gin.H{"status": "UserNotFound", "message": "用户不存在"})
		return
	}
	userBytes, err := json.Marshal(user)
	if err != nil {
		panic(err)
	}
	if err = di.CacheRedis().Set(context.Background(), key, userBytes, time.Second*30).Err(); err != nil {
		panic(err)
	}

	c.JSON(200, user)
}

func (accountController *AccountController) PostUsers(c *gin.Context) {
	jsonBody, err := util.GetJsonBody(c, []string{"counts:数量:+int:*"})
	if err != nil {
		return
	}

	counts, ok := jsonBody["counts"]
	if !ok {
		counts = 100
	}

	startTime := time.Now().UnixNano()

	for i := int64(0); i < counts.(int64); i++ {
		// 多线程写
		go func() {
			db := di.Db()
			err := db.Begin()
			if err != nil {
				log.Println(err)
				return
			}

			rand.Seed(time.Now().UnixNano())
			userName := strconv.Itoa(rand.Int())
			user, err := db.Table("t_users").Fields("user_id").Where(gorose.Data{"user_name": userName}).First()
			if err != nil {
				log.Println(err)
				_ = db.Rollback()
				return
			}
			userId := int64(0)
			if len(user) > 0 { // 记录存在
				userId = user["user_id"].(int64)
			} else { // 记录不存在
				userId, err = db.Table("t_users").Data(gorose.Data{"user_name": userName}).InsertGetId()
				if err != nil {
					log.Println(err)
					_ = db.Rollback()
					return
				}
			}
			sql := "INSERT INTO t_user_counts(user_id,counts) VALUES(?,1) ON DUPLICATE KEY UPDATE counts = counts + 1"
			if _, err = db.Execute(sql, userId); err != nil {
				log.Println(err)
				_ = db.Rollback()
				return
			}

			err = db.Commit()
			if err != nil {
				log.Println(err)
				_ = db.Rollback()
				return
			}
		}()
	}

	timeCost := time.Now().UnixNano() - startTime

	c.JSON(201, gin.H{"time_cost": fmt.Sprintf("%dns", timeCost)})
}
