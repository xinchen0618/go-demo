package ginx

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InternalError 服务异常
//	记录日志并向客户端返回500错误
//	@param c *gin.Context
//	@param err error
func InternalError(c *gin.Context, err error) {
	zap.L().Error(err.Error())
	c.JSON(500, gin.H{"code": "InternalError", "message": "服务异常, 请稍后重试"})
}
