// Package controller API 控制器
package controller

import (
	"fmt"

	"go-demo/config/di"
	"go-demo/internal/consts"
	"go-demo/internal/model"
	"go-demo/internal/service"
	"go-demo/pkg/ginx"
	"go-demo/pkg/gox"

	"github.com/gin-gonic/gin"
	"github.com/golang-module/carbon/v2"
	"github.com/spf13/cast"
)

// 用户相关控制器 DEMO 这里定义一个空结构体用于为大量的 controller 方法做分类
type account struct{}

// Account 这里仅需结构体零值
var Account account

func (account) PostUserLogin(c *gin.Context) {
	jsonBody, err := ginx.GetJSONBody(c, []string{"user_name:用户名:string:+", "password:密码:string:+"})
	if err != nil {
		return
	}

	// 校验密码
	user := struct {
		UserID   int64  `json:"user_id"`
		UserName string `json:"user_name"`
		Password string `json:"password"`
	}{}
	if err := di.DemoDB().Model(&model.TUsers{}).Where("user_name = ?", jsonBody["user_name"]).Limit(1).Find(&user).Error; err != nil {
		ginx.InternalError(c, nil)
		return
	}
	if user.UserID == 0 || !gox.PasswordVerify(jsonBody["password"].(string), user.Password) {
		ginx.Error(c, 400, "UserInvalid", "用户名或密码不正确")
		return
	}

	// JWT 登录
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

	where := "1 = 1"
	bindParams := make([]any, 0)

	userName := queries["user_name"].(string)
	if userName != "" {
		where += " AND user_name LIKE ?"
		bindParams = append(bindParams, "%"+userName+"%")
	}

	items := make([]struct {
		UserID    int64  `json:"user_id"`
		UserName  string `json:"user_name"`
		CreatedAt string `json:"created_at"`
	}, 0)
	paging, err := ginx.Paginate(c, &items, ginx.PageQuery{
		DB:         di.DemoDB(),
		Model:      &model.TUsers{},
		Where:      where,
		BindParams: bindParams,
		OrderBy:    "user_id DESC",
	})
	if err != nil {
		return
	}

	ginx.PageSuccess(c, items, paging)
}

func (account) GetUsersByID(c *gin.Context) {
	userID, err := ginx.FilterParam(c, "用户id", c.Param("user_id"), "+integer", false)
	if err != nil {
		return
	}

	user := model.TUsers{}
	if err := di.DemoDB().Where("user_id = ?", userID).Find(&user).Error; err != nil {
		ginx.InternalError(c, nil)
		return
	}
	if user.UserID == 0 {
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

	// 多线程写 Demo
	ch := make(chan error, userCount)
	psg := di.PoolSeparate(100).Group()
	for i := 0; i < userCount; i++ {
		psg.Submit(func() {
			user := model.TUsers{
				UserName: fmt.Sprintf("U%d%d", carbon.Now().Timestamp(), gox.RandInt64(1111, 9999)),
				Password: gox.PasswordHash("111111"),
			}
			if err := di.DemoDB().Create(&user).Error; err != nil {
				ch <- err
			}
		})
	}
	psg.Wait()
	close(ch)
	okCount := userCount - len(ch)

	ginx.Success(c, 201, gin.H{"ok_count": okCount})
}

func (account) PutUsersByID(c *gin.Context) {
	userID, err := ginx.FilterParam(c, "用户id", c.Param("user_id"), "+integer", false)
	if err != nil {
		return
	}

	jsonBody, err := ginx.GetJSONBody(c, []string{"user_name:用户名:string:?", "password:密码:string:?", "is_vip:VIP身份:[0,1]:?"})
	if err != nil {
		return
	}
	if len(jsonBody) == 0 {
		ginx.Error(c, 400, "ParamError", "请至少传递一个参数")
		return
	}

	user := struct {
		UserID int64
	}{}
	if err := di.DemoDB().Model(&model.TUsers{}).Where("user_id = ?", userID).Find(&user).Error; err != nil {
		ginx.InternalError(c, nil)
		return
	}
	if user.UserID == 0 {
		ginx.Error(c, 404, "UserNotFound", "用户不存在")
		return
	}

	if _, ok := jsonBody["user_name"]; ok {
		conflictUser := struct {
			UserID int64
		}{}
		if err := di.DemoDB().Model(&model.TUsers{}).Where("user_name = ? AND user_id != ?", jsonBody["user_name"], userID).Find(&conflictUser).Error; err != nil {
			ginx.InternalError(c, nil)
			return
		}
		if conflictUser.UserID > 0 {
			ginx.Error(c, 400, "UserConflict", "用户名已存在")
			return
		}
	}
	if password, ok := jsonBody["password"].(string); ok {
		jsonBody["password"] = gox.PasswordHash(password)
	}

	if err := di.DemoDB().Model(&model.TUsers{}).Where("user_id = ?", userID).Updates(jsonBody).Error; err != nil {
		ginx.InternalError(c, nil)
		return
	}

	ginx.Success(c, 200, nil)
}
