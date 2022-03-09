package router

import (
	"go-demo/internal/controller"
	"go-demo/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Account 账号模块 DEMO
func Account(r *gin.Engine) {
	accountGroup := r.Group("/account")
	{
		accountGroup.POST("/v1/login", controller.Account.PostUserLogin)
		accountGroup.DELETE("/v1/logout", middleware.UserAuth(), controller.Account.DeleteUserLogout)

		accountGroup.GET("/v1/users", controller.Account.GetUsers)
		accountGroup.GET("/v1/users/:user_id", controller.Account.GetUsersById)
		accountGroup.POST("/v1/users", controller.Account.PostUsers)
	}
}
