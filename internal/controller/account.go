package controller

import (
	"fmt"
	"go-demo/config/consts"
	"go-demo/config/di"
	"go-demo/internal/service"
	"go-demo/pkg/dbx"
	"go-demo/pkg/ginx"
	"go-demo/pkg/gox"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// 这里定义一个空结构体用于为大量的controller方法做分类
type account struct{}

// Account 这里仅需结构体零值
var Account account

func (account) PostUserLogin(c *gin.Context) { // 先生成JWT, 再记录redis白名单
	jsonBody, err := ginx.GetJsonBody(c, []string{"user_name:用户名:string:+", "password:密码:string:+"})
	if err != nil {
		return
	}

	// 校验密码, 实际密码应使用gox.PasswordHash()创建散列
	sql := "SELECT user_id,user_name FROM t_users WHERE user_name=? AND password=? LIMIT 1"
	user, err := dbx.FetchOne(di.DemoDb(), sql, jsonBody["user_name"], jsonBody["password"])
	if err != nil {
		ginx.InternalError(c)
		return
	}
	if 0 == len(user) {
		ginx.Error(c, 400, "UserInvalid", "用户名或密码不正确")
		return
	}

	// JWT登录
	token, err := service.Auth.JwtLogin(consts.UserJwt, user["user_id"].(int64), user["user_name"].(string))
	if err != nil {
		ginx.InternalError(c)
		return
	}

	ginx.Success(c, 200, gin.H{"user_id": user["user_id"], "token": token})
}

func (account) DeleteUserLogout(c *gin.Context) {
	userId := c.GetInt64("userId")
	token := c.Request.Header.Get("Authorization")[7:]
	if err := service.Auth.JwtLogout(consts.UserJwt, token, userId); err != nil {
		ginx.InternalError(c)
		return
	}

	ginx.Success(c, 204)
}

func (account) GetUsers(c *gin.Context) {
	// Cache Demo
	queries, err := ginx.GetQueries(c, []string{`user_name:用户名:string:""`, "page:页码:+int:1", "per_page:页大小:+int:12"})
	if err != nil {
		return
	}

	key, err := gox.Md5x(queries)
	if err != nil {
		ginx.InternalError(c)
		return
	}

	pageItemsCache, err := ginx.GetOrSetCache(c, key, 3*time.Second, func() (interface{}, error) {
		where := "1"
		bindParams := []interface{}{}

		userName := queries["user_name"].(string)
		if userName != "" {
			where += " AND user_name LIKE ?"
			bindParams = append(bindParams, fmt.Sprintf("%%%s%%", userName))
		}

		return ginx.GetPageItems(c, ginx.PageQuery{
			Db:         di.DemoDb(),
			Select:     "user_id,user_name,money,created_at,updated_at",
			From:       "t_users",
			Where:      where,
			BindParams: bindParams,
			OrderBy:    "user_id DESC",
		})
	})
	if err != nil {
		return
	}

	ginx.Success(c, 200, pageItemsCache)
}

func (account) GetUsersById(c *gin.Context) {
	userId, err := ginx.FilterParam(c, "用户id", c.Param("user_id"), "+int", false)
	if err != nil {
		return
	}

	user, err := service.Cache.Get(di.DemoDb(), "t_users", "user_id", userId)
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

func (account) PostUsers(c *gin.Context) {
	// 延时队列Demo
	//userName := fmt.Sprintf("QU%d", gox.RandInt64(111111, 999999))
	//if err := service.Queue.EnqueueIn("user:AddUser", map[string]interface{}{"user_name": userName}, 5*time.Second); err != nil {
	//	ginx.InternalError(c)
	//	return
	//}
	//ginx.Success(c, 201, gin.H{"user_name": userName})

	jsonBody, err := ginx.GetJsonBody(c, []string{"counts:数量:+int:*"})
	if err != nil {
		return
	}

	var counts = 100
	if _, ok := jsonBody["counts"]; ok {
		counts = cast.ToInt(jsonBody["counts"])
	}

	// 多线程写Demo
	wpsg := di.WorkerPoolSeparate(100).Group()
	for i := 0; i < counts; i++ {
		wpsg.Submit(func() {
			userData := map[string]interface{}{
				"user_name": fmt.Sprintf("U%d", gox.RandInt64(111111111, 999999999)),
			}
			_, _ = service.User.CreateUser(userData)
		})
	}
	wpsg.Wait()

	ginx.Success(c, 201, gin.H{"counts": counts})
}
