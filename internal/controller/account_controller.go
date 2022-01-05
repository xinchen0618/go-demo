package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"go-demo/config"
	"go-demo/config/consts"
	"go-demo/config/di"
	"go-demo/internal/service"
	"go-demo/pkg/dbx"
	"go-demo/pkg/ginx"
	"go-demo/pkg/gox"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 这里定义一个空结构体用于为大量的controller方法做分类
type accountController struct{}

// AccountController 这里仅需结构体零值, router通过controller.XxxController.Xxx的形式引用旗下定义的方法
var AccountController accountController

func (accountController) PostUserLogin(c *gin.Context) { // 先生成JWT, 再记录redis白名单
	jsonBody, err := ginx.GetJsonBody(c, []string{"user_name:用户名:string:+", "password:密码:string:+"})
	if err != nil {
		return
	}

	sql := "SELECT user_id FROM t_users WHERE user_name=? AND password=? LIMIT 1"
	user, err := dbx.FetchOne(di.Db(), sql, jsonBody["user_name"], jsonBody["password"])
	if err != nil {
		ginx.InternalError(c)
		return
	}
	if 0 == len(user) {
		ginx.Error(c, 400, "UserInvalid", "用户名或密码不正确")
		return
	}

	// JWT
	loginTtl := 30 * 24 * time.Hour // 登录有效时长
	claims := &jwt.StandardClaims{
		Audience:  jsonBody["user_name"].(string),
		ExpiresAt: time.Now().Add(loginTtl).Unix(),
		Id:        strconv.FormatInt(user["user_id"].(int64), 10),
		IssuedAt:  time.Now().Unix(),
		Issuer:    "account:PostUserLogin",
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
	key := fmt.Sprintf(consts.JwtUserLogin, claims.Id, tokenAtoms[2])
	if err = di.JwtRedis().Set(context.Background(), key, payload, loginTtl).Err(); err != nil {
		ginx.InternalError(c, err)
		return
	}

	ginx.Success(c, 200, gin.H{"user_id": user["user_id"], "token": tokenString})
}

func (accountController) DeleteUserLogout(c *gin.Context) {
	userId := c.GetInt64("userId")

	// 删除对应redis白名单记录
	tokenString := c.Request.Header.Get("Authorization")[7:]
	tokenAtoms := strings.Split(tokenString, ".")
	key := fmt.Sprintf(consts.JwtUserLogin, userId, tokenAtoms[2])
	if err := di.JwtRedis().Del(context.Background(), key).Err(); err != nil {
		ginx.InternalError(c, err)
		return
	}

	ginx.Success(c, 204)
}

func (accountController) GetUsers(c *gin.Context) {
	queries, err := ginx.GetQueries(c, []string{"page:页码:+int:1", "per_page:页大小:+int:12"})
	if err != nil {
		return
	}
	key, err := gox.Md5x(queries)
	if err != nil {
		ginx.InternalError(c)
		return
	}
	key = fmt.Sprintf(consts.CacheUsers, key)
	pageItems, err := ginx.GetOrSetCache(c, key, 3*time.Second, func() (interface{}, error) {
		pageItems, err := ginx.GetPageItems(ginx.PageQuery{
			GinCtx: c,
			Db:     di.Db(),
			Select: "user_id,user_name,money,created_at,updated_at",
			From:   "t_users",
			//Where:      "user_id > ?",
			//BindParams: []interface{}{5},
			OrderBy: "user_id DESC",
		})
		if err != nil {
			return ginx.PageItems{}, err
		}

		// 多线程读
		wpg := di.WorkerPool().Group()
		for _, item := range pageItems.Items {
			item := item
			wpg.Submit(func() {
				userCounts, _ := service.CacheService.Get(di.Db(), "t_user_counts", "user_id", item["user_id"])
				item["counts"] = 0
				if counts, ok := userCounts["counts"]; ok {
					item["counts"] = counts
				}
			})
		}
		wpg.Wait()

		return pageItems, nil
	})
	if err != nil {
		return
	}

	ginx.Success(c, 200, pageItems)
}

func (accountController) GetUsersById(c *gin.Context) {
	userId, err := ginx.FilterParam(c, "用户id", c.Param("user_id"), "+int", false)
	if err != nil {
		return
	}

	user, err := service.CacheService.Get(di.Db(), "t_users", "user_id", userId)
	if err != nil {
		ginx.InternalError(c)
		return
	}
	if 0 == len(user) {
		ginx.Error(c, 404, "UserNotFound", "用户不存在")
		return
	}

	ginx.Success(c, 200, user)
}

func (accountController) PostUsers(c *gin.Context) {
	//userName := fmt.Sprintf("QU%d", gox.RandInt64(111111, 999999))
	//if err := service.QueueService.EnqueueIn("user:AddUser", map[string]interface{}{"user_name": userName}, 5*time.Second); err != nil {
	//	ginx.InternalError(c)
	//	return
	//}
	//ginx.Success(c, 201, gin.H{"user_name": userName})

	//userId := gox.RandInt64(111111, 999999)
	//if err := service.QueueService.LowEnqueue("user:AddUserCounts", map[string]interface{}{"user_id": userId}); err != nil {
	//	ginx.InternalError(c)
	//	return
	//}
	//ginx.Success(c, 201, gin.H{"user_id": userId})

	jsonBody, err := ginx.GetJsonBody(c, []string{"counts:数量:+int:*"})
	if err != nil {
		return
	}

	var counts = 100
	if _, ok := jsonBody["counts"]; ok {
		counts = int(jsonBody["counts"].(int64))
	}

	perPage := 120
	doneCounts := 0
	for {
		if doneCounts >= counts {
			break
		}
		wpg := di.WorkerPool().Group()
		for i := 0; i < perPage; i++ {
			if doneCounts >= counts {
				break
			}
			// 多线程写
			doneCounts++
			wpg.Submit(func() {
				db := di.Db()
				if err := db.Begin(); err != nil {
					zap.L().Error(err.Error())
					return
				}

				userName := fmt.Sprintf("U%d", gox.RandInt64(111111111, 999999999))
				user, err := dbx.FetchOne(db, "SELECT user_id FROM t_users WHERE user_name=?", userName)
				if err != nil {
					_ = db.Rollback()
					return
				}
				var userId int64
				if len(user) > 0 { // 记录存在
					userId = user["user_id"].(int64)
				} else { // 记录不存在
					userId, err = dbx.Insert(db, "t_users", map[string]interface{}{"user_name": userName})
					if err != nil {
						_ = db.Rollback()
						return
					}
				}
				sql := "INSERT INTO t_user_counts(user_id,counts) VALUES(?,?) ON DUPLICATE KEY UPDATE counts = counts + 1"
				if _, err = dbx.Execute(db, sql, userId, gox.RandInt64(1, 9)); err != nil {
					_ = db.Rollback()
					return
				}
				if err := service.CacheService.Delete("t_user_counts", userId); err != nil {
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
		wpg.Wait()
	}

	ginx.Success(c, 201, gin.H{"counts": counts})
}
