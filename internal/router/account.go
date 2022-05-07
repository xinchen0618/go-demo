package router

import (
	"go-demo/internal/controller"
	"go-demo/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Account 账号模块 DEMO
func Account(r *gin.Engine) {
	accountGroup := r.Group("/account", middleware.UserJwtParse())
	{
		accountGroup.POST("/v1/login", controller.Account.PostUserLogin)                              // 用户登录
		accountGroup.DELETE("/v1/logout", middleware.UserAuth(), controller.Account.DeleteUserLogout) // 用户退出登录

		accountGroup.GET("/v1/users", controller.Account.GetUsers)              // 获取用户列表
		accountGroup.GET("/v1/users/:user_id", controller.Account.GetUsersById) // 获取用户详情
		accountGroup.POST("/v1/users", controller.Account.PostUsers)            // 新增用户
		accountGroup.PUT("/v1/users/:user_id", controller.Account.PutUsersById) // 修改用户信息
	}
}
