// Package router API 路由
package router

import (
	"go-demo/config/consts"
	"go-demo/internal/controller"
	"go-demo/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Account 账号模块 DEMO
func Account(r *gin.Engine) {
	accountGroup := r.Group("/account/v1", middleware.JWTParse(consts.UserJWT))
	{
		accountGroup.POST("/login", middleware.SubmitLimit(), controller.Account.PostUserLogin)    // 用户登录
		accountGroup.DELETE("/logout", middleware.UserAuth(), controller.Account.DeleteUserLogout) // 用户退出登录

		accountGroup.GET("/users", controller.Account.GetUsers)                             // 获取用户列表
		accountGroup.GET("/users/:user_id", controller.Account.GetUsersByID)                // 获取用户详情
		accountGroup.POST("/users", middleware.SubmitLimit(), controller.Account.PostUsers) // 新增用户
		accountGroup.PUT("/users/:user_id", controller.Account.PutUsersByID)                // 修改用户信息
	}
}
