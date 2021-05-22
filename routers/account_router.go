package routers

import (
	"go-test/controllers"

	"github.com/gin-gonic/gin"
)

func LoadAccount(e *gin.Engine) {

	userGroup := e.Group("/account")
	{
		userGroup.POST("/v1/login", controllers.PostUserLogin)
	}
}
