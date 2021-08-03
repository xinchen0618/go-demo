package router

import "github.com/gin-gonic/gin"

// Init 注册路由
//	@param r *gin.Engine
func Init(r *gin.Engine) {
	LoadAccount(r)
}
