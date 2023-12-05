// Package ginx Gin 增强函数
//
//	此包中出现 error 会向客户端输出 4xx/500 错误, 调用时捕获到 error 直接结束业务逻辑即可.
package ginx

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Success 输出成功信息
//
//	body 数据会 json 编码输出给客户端, nil 表示无内容输出.
func Success(c *gin.Context, httpCode int, body any) {
	if body == nil {
		body = gin.H{}
	}
	c.JSON(httpCode, body)
}

// PageSuccess 输出分页结果
//
//	items 列表数据
func PageSuccess(c *gin.Context, items any, paging Paging) {
	body := struct {
		Page         int64 `json:"page"`          // 页码
		PerPage      int64 `json:"per_page"`      // 页大小
		TotalPages   int64 `json:"total_pages"`   // 总页数
		TotalResults int64 `json:"total_results"` // 总记录数
		Items        any   `json:"items"`         // 列表
	}{
		paging.Page,
		paging.PerPage,
		paging.TotalPages,
		paging.TotalResults,
		items,
	}
	c.JSON(200, body)
}

// Error 输出失败信息
func Error(c *gin.Context, httpCode int, code, message string) {
	c.AbortWithStatusJSON(httpCode, gin.H{"code": code, "message": message})
}

// InternalError 输出500错误
//
//	err 记录错误日志, nil 表示无需记录, 项目中定义的方法错误会就近记录, 无需重复记录.
func InternalError(c *gin.Context, err error) {
	if err != nil {
		zap.L().Error(err.Error())
	}
	Error(c, 500, "InternalError", "服务异常, 请稍后重试")
}
