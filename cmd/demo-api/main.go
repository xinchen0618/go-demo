package main

import (
	"errors"
	"fmt"
	"go-demo/config"
	"go-demo/internal/router"
	"go-demo/pkg/ginx"
	"go-demo/pkg/gox"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
)

// recovery 主goroutine中panic兜底处理
//	@return gin.HandlerFunc
func recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				ginx.InternalError(c, errors.New(fmt.Sprint(err)))
			}
		}()
		c.Next()
	}
}

// cors 跨域处理
//	@return gin.HandlerFunc
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("Origin") != "" {
			c.Header("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
			c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language, DNT, Keep-Alive, User-Agent, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			c.Header("Access-Control-Max-Age", "1728000")
			c.Header("Access-Control-Allow-Credentials", "false")
			if "OPTIONS" == c.Request.Method {
				ginx.Success(c, 200)
			}
		}
		c.Next()
	}
}

func main() {
	// 实例化gin
	if gox.InSlice(config.GetRuntimeEnv(), []string{"prod", "stage"}) {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Panic处理
	r.Use(recovery())
	// 跨域处理
	r.Use(cors())

	// 加载路由
	router.Account(r)

	// Run gin
	addr := fmt.Sprintf(":%d", config.Get("server_port"))
	if err := endless.ListenAndServe(addr, r); err != nil {
		panic(err)
	}
}
