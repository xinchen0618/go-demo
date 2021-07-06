package routers

import (
	"github.com/gin-gonic/gin"
	"go-test/controllers"
)

func LoadAccount(e *gin.Engine) {
	accountGroup := e.Group("/account")
	{
		accountGroup.POST("/v1/login", controllers.PostUserLogin)
		accountGroup.DELETE("/v1/logout", controllers.DeleteUserLogout)
		accountGroup.GET("/v1/users", controllers.GetUsers)
	}
}
