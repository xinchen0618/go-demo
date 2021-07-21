package router

import (
	"go-demo/controller"

	"github.com/gin-gonic/gin"
)

func LoadAccount(r *gin.Engine) {
	accountGroup := r.Group("/account")
	{
		accountController := controller.AccountController{}

		accountGroup.POST("/v1/login", accountController.PostUserLogin)
		accountGroup.DELETE("/v1/logout", accountController.DeleteUserLogout)

		accountGroup.GET("/v1/users", accountController.GetUsers)
		accountGroup.GET("/v1/users/:user_id", accountController.GetUsersById)
		accountGroup.POST("/v1/users", accountController.PostUsers)
	}
}
