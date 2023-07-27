package controller

import (
	"fmt"

	"go-demo/config/consts"
	"go-demo/config/di"
	"go-demo/internal/service"
	"go-demo/pkg/dbcache"
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
	jsonBody, err := ginx.GetJSONBody(c, []string{"user_name:用户名:string:+", "password:密码:string:+"})
	if err != nil {
		return
	}

	// 校验密码
	var user struct {
		UserID   int64  `json:"user_id"`
		UserName string `json:"user_name"`
		Password string `json:"password"`
	}
	sql := "SELECT user_id,user_name,password FROM t_users WHERE user_name=? LIMIT 1"
	if err := dbx.TakeOne(&user, di.DemoDB(), sql, jsonBody["user_name"]); err != nil {
		ginx.InternalError(c, nil)
		return
	}
	if user.UserID == 0 || !gox.PasswordVerify(jsonBody["password"].(string), user.Password) {
		ginx.Error(c, 400, "UserInvalid", "用户名或密码不正确")
		return
	}

	// JWT登录
	token, err := service.Auth.JWTLogin(consts.UserJWT, user.UserID, user.UserName)
	if err != nil {
		ginx.InternalError(c, nil)
		return
	}

	ginx.Success(c, 200, gin.H{"user_id": user.UserID, "token": token})
}

func (account) DeleteUserLogout(c *gin.Context) {
	userID := c.GetInt64("userID")
	token := c.Request.Header.Get("Authorization")[7:]
	if err := service.Auth.JWTLogout(consts.UserJWT, token, userID); err != nil {
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
	bindParams := make([]any, 0)

	userName := queries["user_name"].(string)
	if userName != "" {
		where += " AND user_name LIKE ?"
		bindParams = append(bindParams, fmt.Sprintf("%%%s%%", gox.AddSlashes(userName)))
	}

	pageItems, err := ginx.GetPageItems(c, ginx.PageQuery{
		DB:         di.DemoDB(),
		Select:     "user_id,user_name,created_at",
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

func (account) GetUsersByID(c *gin.Context) {
	userID, err := ginx.FilterParam(c, "用户id", c.Param("user_id"), "+integer", false)
	if err != nil {
		return
	}

	user, err := dbcache.Get(di.CacheRedis(), di.DemoDB(), "t_users", userID)
	if err != nil {
		ginx.InternalError(c, nil)
		return
	}
	if len(user) == 0 {
		ginx.Error(c, 404, "UserNotFound", "用户不存在")
		return
	}

	ginx.Success(c, 200, user)
}

func (account) PostUsers(c *gin.Context) {
	jsonBody, err := ginx.GetJSONBody(c, []string{"user_count:数量:+integer:*"})
	if err != nil {
		return
	}

	userCount := 100
	if _, ok := jsonBody["user_count"]; ok {
		userCount = cast.ToInt(jsonBody["user_count"])
	}

	// 多线程写Demo
	ch := make(chan error, userCount)
	wpsg := di.WorkerPoolSeparate(100).Group()
	for i := 0; i < userCount; i++ {
		wpsg.Submit(func() {
			userData := map[string]any{
				"user_name": fmt.Sprintf("U%d", gox.RandInt64(111111111, 999999999)),
				"password":  gox.PasswordHash("111111"),
			}
			if _, err := service.User.CreateUser(userData); err != nil {
				ch <- err
			}
		})
	}
	wpsg.Wait()
	close(ch)
	okCount := userCount - len(ch)

	ginx.Success(c, 201, gin.H{"ok_count": okCount})
}

func (account) PutUsersByID(c *gin.Context) {
	userID, err := ginx.FilterParam(c, "用户id", c.Param("user_id"), "+integer", false)
	if err != nil {
		return
	}

	jsonBody, err := ginx.GetJSONBody(c, []string{"user_name:用户名:string:?", "password:密码:string:?"})
	if err != nil {
		return
	}
	if len(jsonBody) == 0 {
		ginx.Error(c, 400, "ParamError", "请至少传递一个参数")
		return
	}

	user, err := dbcache.Get(di.CacheRedis(), di.DemoDB(), "t_users", userID)
	if err != nil {
		ginx.InternalError(c, nil)
		return
	}
	if len(user) == 0 {
		ginx.Error(c, 404, "UserNotFound", "用户不存在")
		return
	}

	if _, ok := jsonBody["user_name"]; ok {
		sql := "SELECT user_id FROM t_users WHERE user_name = ? AND user_id != ?"
		userConflict, err := dbx.FetchOne(di.DemoDB(), sql, jsonBody["user_name"], userID)
		if err != nil {
			ginx.InternalError(c, nil)
			return
		}
		if len(userConflict) > 0 {
			ginx.Error(c, 400, "UserConflict", "用户名已存在")
			return
		}
	}
	if password, ok := jsonBody["password"].(string); ok {
		jsonBody["password"] = gox.PasswordHash(password)
	}

	if _, err := dbcache.Update(di.CacheRedis(), di.DemoDB(), "t_users", jsonBody, "user_id = ?", userID); err != nil {
		ginx.InternalError(c, nil)
		return
	}

	ginx.Success(c, 200, nil)
}
