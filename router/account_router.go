package router

import (
	"go-demo/controller"

	"github.com/gin-gonic/gin"
)

// LoadAccount 这里定义路由, 然后在router.go中统一注册
func LoadAccount(r *gin.Engine) {
	accountGroup := r.Group("/account")
	{
		accountGroup.POST("/v1/login", controller.AccountController.PostUserLogin)
		accountGroup.DELETE("/v1/logout", controller.AccountController.DeleteUserLogout)

		accountGroup.GET("/v1/users", controller.AccountController.GetUsers)
		accountGroup.GET("/v1/users/:user_id", controller.AccountController.GetUsersById)
		accountGroup.POST("/v1/users", controller.AccountController.PostUsers)
	}
}
