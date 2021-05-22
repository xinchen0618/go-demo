package routers

import (
	"github.com/gin-gonic/gin"
	"go-test/controllers"
)

func LoadUser(e *gin.Engine) {
	userGroup := e.Group("/user")
	{
		userGroup.GET("/v1/users", controllers.GetUsers)
	}
}
