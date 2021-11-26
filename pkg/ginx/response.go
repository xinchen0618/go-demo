package ginx

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Success 成功
//	@param c *gin.Context
//	@param httpCode int
//	@param obj ...interface{}
func Success(c *gin.Context, httpCode int, obj ...interface{}) {
	if len(obj) > 0 {
		c.JSON(httpCode, obj[0])
	} else {
		c.JSON(httpCode, gin.H{})
	}
}

// Error 失败
//	@param c *gin.Context
//	@param httpCode int
//	@param code string
//	@param message string
func Error(c *gin.Context, httpCode int, code, message string) {
	c.AbortWithStatusJSON(httpCode, gin.H{"code": code, "message": message})
}

// InternalError 服务异常
//	记录日志并向客户端返回500错误
//	@param c *gin.Context
//	@param err ...error
func InternalError(c *gin.Context, err ...error) {
	if len(err) > 0 {
		zap.L().Error(err[0].Error())
	}
	Error(c, 500, "InternalError", "服务异常, 请稍后重试")
}
