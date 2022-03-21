package ginx

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Success 向客户端输出成功信息
//	@param c *gin.Context
//	@param httpCode int
//	@param obj ...any 选填, 数据会json编码输出给客户端
func Success(c *gin.Context, httpCode int, obj ...any) {
	if len(obj) > 0 {
		c.JSON(httpCode, obj[0])
	} else {
		c.JSON(httpCode, gin.H{})
	}
}

// Error 向客户端输出失败信息
//	@param c *gin.Context
//	@param httpCode int
//	@param code string
//	@param message string
func Error(c *gin.Context, httpCode int, code, message string) {
	c.AbortWithStatusJSON(httpCode, gin.H{"code": code, "message": message})
}

// InternalError 向客户端输出500错误
//	@param c *gin.Context
//	@param err ...error 选填, 记录错误日志
func InternalError(c *gin.Context, err ...error) {
	if len(err) > 0 {
		zap.L().Error(err[0].Error())
	}
	Error(c, 500, "InternalError", "服务异常, 请稍后重试")
}
