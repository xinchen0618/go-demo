package routers

import (
	"github.com/gin-gonic/gin"
	"go-test/controllers"
)

func LoadAccount(e *gin.Engine) {
	userGroup := e.Group("/account")
	{
		userGroup.POST("/v1/login", controllers.PostUserLogin)
		userGroup.DELETE("/v1/logout", controllers.DeleteUserLogout)
	}
}
