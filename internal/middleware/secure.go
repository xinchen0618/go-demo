// Package middleware Gin 中间件
package middleware

import (
	"errors"
	"fmt"

	"go-demo/pkg/ginx"

	"github.com/gin-gonic/gin"
)

// Recovery 主 Goroutine 中 panic 处理
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				ginx.InternalError(c, errors.New(fmt.Sprint(r)))
			}
		}()
		c.Next()
	}
}

// CORS 跨域处理
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("Origin") != "" {
			c.Header("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
			c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language, DNT, Keep-Alive, User-Agent, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			c.Header("Access-Control-Max-Age", "1728000")
			c.Header("Access-Control-Allow-Credentials", "false")
			if c.Request.Method == "OPTIONS" {
				ginx.Success(c, 200, nil)
				return
			}
		}
		c.Next()
	}
}
