package ginx

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Success 向客户端输出成功信息
//	@param c *gin.Context
//	@param httpCode int
//	@param body any 数据会json编码输出给客户端, nil表示无内容输出
func Success(c *gin.Context, httpCode int, body any) {
	if nil == body {
		body = gin.H{}
	}
	c.JSON(httpCode, body)
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
//	@param err error 错误记录日志, nil表示无需记录, 项目中方法的错误会就近记录, 无需重复记录
func InternalError(c *gin.Context, err error) {
	if err != nil {
		zap.L().Error(err.Error())
	}
	Error(c, 500, "InternalError", "服务异常, 请稍后重试")
}
