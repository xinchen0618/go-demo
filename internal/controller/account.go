package controller

import (
	"fmt"

	"go-demo/config/consts"
	"go-demo/config/di"
	"go-demo/internal/service"
	"go-demo/pkg/dbx"
	"go-demo/pkg/ginx"
	"go-demo/pkg/gox"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// 用户相关控制器 DEMO 这里定义一个空结构体用于为大量的controller方法做分类
type account struct{}

// Account 这里仅需结构体零值
var Account account

func (account) PostUserLogin(c *gin.Context) {
	jsonBody, err := ginx.GetJsonBody(c, []string{"user_name:用户名:string:+", "password:密码:string:+"})
	if err != nil {
		return
	}

	// 校验密码
	// language=SQL
	sql := "SELECT user_id,user_name,password FROM t_users WHERE user_name=? LIMIT 1"
	user, err := dbx.FetchOne(di.DemoDb(), sql, jsonBody["user_name"])
	if err != nil {
		ginx.InternalError(c, nil)
		return
	}
	if 0 == len(user) || !gox.PasswordVerify(jsonBody["password"].(string), user["password"].(string)) {
		ginx.Error(c, 400, "UserInvalid", "用户名或密码不正确")
		return
	}

	// JWT登录
	token, err := service.Auth.JwtLogin(consts.UserJwt, user["user_id"].(int64), user["user_name"].(string))
	if err != nil {
		ginx.InternalError(c, nil)
		return
	}

	ginx.Success(c, 200, gin.H{"user_id": user["user_id"], "token": token})
}

func (account) DeleteUserLogout(c *gin.Context) {
	userId := c.GetInt64("userId")
	token := c.Request.Header.Get("Authorization")[7:]
	if err := service.Auth.JwtLogout(consts.UserJwt, token, userId); err != nil {
		ginx.InternalError(c, nil)
		return
	}

	ginx.Success(c, 204, nil)
}

func (account) GetUsers(c *gin.Context) {
	// 假设需要分页并可以按名称搜索
	queries, err := ginx.GetQueries(c, []string{`user_name:用户名:string:""`})
	if err != nil {
		return
	}

	where := "1"
	bindParams := []any{}

	userName := queries["user_name"].(string)
	if userName != "" {
		where += " AND user_name LIKE ?"
		bindParams = append(bindParams, fmt.Sprintf("%%%s%%", userName))
	}

	pageItems, err := ginx.GetPageItems(c, ginx.PageQuery{
		Db:         di.DemoDb(),
		Select:     "user_id,user_name,money,created_at,updated_at",
		From:       "t_users",
		Where:      where,
		BindParams: bindParams,
		OrderBy:    "user_id DESC",
	})
	if err != nil {
		return
	}

	ginx.Success(c, 200, pageItems)
}

func (account) GetUsersById(c *gin.Context) {
	userId, err := ginx.FilterParam(c, "用户id", c.Param("user_id"), "+int", false)
	if err != nil {
		return
	}

	user, err := service.Cache.Get(di.DemoDb(), "t_users", "user_id", userId)
	if err != nil {
		ginx.InternalError(c, nil)
		return
	}
	if 0 == len(user) {
		ginx.Error(c, 404, "UserNotFound", "用户不存在")
		return
	}

	ginx.Success(c, 200, user)
}

func (account) PostUsers(c *gin.Context) {
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
			userData := map[string]any{
				"user_name": fmt.Sprintf("U%d", gox.RandInt64(111111111, 999999999)),
				"password":  gox.PasswordHash("111111"),
			}
			_, _ = service.User.CreateUser(userData)
		})
	}
	wpsg.Wait()

	ginx.Success(c, 201, gin.H{"counts": counts})
}

func (account) PutUsersById(c *gin.Context) {
	userId, err := ginx.FilterParam(c, "用户id", c.Param("user_id"), "+int", false)
	if err != nil {
		return
	}
	jsonBody, err := ginx.GetJsonBody(c, []string{"user_name:用户名:string:?", "money:金额:decimal:*"})
	if err != nil {
		return
	}
	if 0 == len(jsonBody) {
		ginx.Error(c, 400, "ParamError", "请至少传递一个参数")
		return
	}

	user, err := service.Cache.Get(di.DemoDb(), "t_users", "user_id", userId)
	if err != nil {
		ginx.InternalError(c, nil)
		return
	}
	if 0 == len(user) {
		ginx.Error(c, 404, "UserNotFound", "用户不存在")
		return
	}

	if _, ok := jsonBody["user_name"]; ok {
		// language=SQL
		sql := "SELECT user_id FROM t_users WHERE user_name = ? AND user_id != ?"
		userConflict, err := dbx.FetchOne(di.DemoDb(), sql, jsonBody["user_name"], userId)
		if err != nil {
			ginx.InternalError(c, nil)
			return
		}
		if len(userConflict) > 0 {
			ginx.Error(c, 400, "UserConflict", "用户名已存在")
			return
		}
	}

	if _, err := dbx.Update(di.DemoDb(), "t_users", jsonBody, "user_id = ?", userId); err != nil {
		ginx.InternalError(c, nil)
		return
	}
	_ = service.Cache.Delete("t_users", userId)

	ginx.Success(c, 204, nil)
}
