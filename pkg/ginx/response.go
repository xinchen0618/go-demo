// Package ginx Gin增强函数
//
//	此包中出现error会向客户端输出4xx/500错误, 调用时捕获到error直接结束业务逻辑即可.
package ginx

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Success 输出成功信息
//
//	body 数据会json编码输出给客户端, nil表示无内容输出.
func Success(c *gin.Context, httpCode int, body any) {
	if body == nil {
		body = gin.H{}
	}
	c.JSON(httpCode, body)
}

// Error 输出失败信息
func Error(c *gin.Context, httpCode int, code, message string) {
	c.AbortWithStatusJSON(httpCode, gin.H{"code": code, "message": message})
}

// InternalError 输出500错误
//
//	err 为nil时表示无需记录, 项目中定义的方法错误会就近记录, 无需重复记录.
func InternalError(c *gin.Context, err error) {
	if err != nil {
		zap.L().Error(err.Error())
	}
	Error(c, 500, "InternalError", "服务异常, 请稍后重试")
}
