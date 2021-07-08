package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-test/di"
	"go-test/service"
	"go-test/util"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gohouse/gorose/v2"
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
	for key, _ := range result["items"].([]gorose.Data) {
		wg.Add(1)

		go func(key int) {
			defer wg.Done()

			sql := "SELECT counts FROM t_user_counts WHERE user_id = ? LIMIT 1"
			userCount, err := di.Db().Query(sql, result["items"].([]gorose.Data)[key]["user_id"])
			if err != nil {
				log.Printf("%v\n", err)
				return
			}
			result["items"].([]gorose.Data)[key]["counts"] = userCount[0]["counts"]
		}(key)
	}
	wg.Wait()

	c.JSON(200, result)
}

func PostUsers(c *gin.Context) {
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
				log.Printf("%v\n", err)
				return
			}

			rand.Seed(time.Now().UnixNano())
			userName := strconv.Itoa(rand.Int())
			sql := "SELECT user_id FROM t_users WHERE user_name = ? LIMIT 1"
			user, err := db.Query(sql, userName)
			if err != nil {
				log.Printf("%v\n", err)
				return
			}
			if 0 == len(user) { // 记录不存在
				userId, err := db.Table("t_users").Data(gorose.Data{"user_name": userName}).InsertGetId()
				if err != nil {
					log.Printf("%v\n", err)
					return
				}
				if _, err = db.Table("t_user_counts").Data(gorose.Data{"user_id": userId, "counts": 1}).Insert(); err != nil {
					log.Printf("%v\n", err)
					return
				}
			} else { // 记录存在
				userId := user[0]["user_id"].(int64)
				sql = "UPDATE t_user_counts SET counts = counts + 1 WHERE user_id = ?"
				if _, err = db.Execute(sql, userId); err != nil {
					log.Printf("%v\n", err)
					return
				}
			}

			err = db.Commit()
			if err != nil {
				log.Printf("%v\n", err)
				return
			}
		}()
	}

	timeCost := time.Now().UnixNano() - startTime

	c.JSON(201, gin.H{"time_cost": fmt.Sprintf("%d ns", timeCost)})
}
