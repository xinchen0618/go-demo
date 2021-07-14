package router

import (
	"go-test/controller"

	"github.com/gin-gonic/gin"
)

func LoadAccount(e *gin.Engine) {
	accountGroup := e.Group("/account")
	{
		accountGroup.POST("/v1/login", controller.PostUserLogin)
		accountGroup.DELETE("/v1/logout", controller.DeleteUserLogout)

		accountGroup.GET("/v1/users", controller.GetUsers)
		accountGroup.GET("/v1/users/:user_id", controller.GetUsersById)
		accountGroup.POST("/v1/users", controller.PostUsers)
	}
}
