package router

import (
	"go-demo/internal/controller"
	"go-demo/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Account 账号模块
func Account(r *gin.Engine) {
	accountGroup := r.Group("/account")
	{
		accountGroup.POST("/v1/login", controller.AccountController.PostUserLogin)
		accountGroup.DELETE("/v1/logout", middleware.UserAuth(), controller.AccountController.DeleteUserLogout)

		accountGroup.GET("/v1/users", controller.AccountController.GetUsers)
		accountGroup.GET("/v1/users/:user_id", controller.AccountController.GetUsersById)
		accountGroup.POST("/v1/users", controller.AccountController.PostUsers)
	}
}
