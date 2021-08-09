package main

import (
	"errors"
	"fmt"
	"go-demo/config"
	"go-demo/internal/router"
	"go-demo/pkg/ginx"
	"net/http"
	"os"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
)

// recovery 主routine中panic兜底处理
// 	除程序初始化可以使用panic, 其他地方必须避免出现panic
//	goroutine中的panic这里是捕获不到的, 要自行recover
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
	return func(context *gin.Context) {
		if context.Request.Header.Get("Origin") != "" {
			context.Header("Access-Control-Allow-Origin", context.Request.Header.Get("Origin"))
			context.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			context.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language, DNT, Keep-Alive, User-Agent, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			context.Header("Access-Control-Max-Age", "1728000")
			context.Header("Access-Control-Allow-Credentials", "false")

			if "OPTIONS" == context.Request.Method {
				context.JSON(http.StatusOK, gin.H{})
			}
		}

		//处理请求
		context.Next()
	}
}

func main() {
	// 实例化gin
	runtimeEnv := os.Getenv("RUNTIME_ENV")
	if runtimeEnv == "" || runtimeEnv == "prod" || runtimeEnv == "stage" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Panic处理
	r.Use(recovery())
	// 跨域处理
	r.Use(cors())
	// 加载路由
	router.Init(r)

	// Run gin
	if err := endless.ListenAndServe(fmt.Sprintf(":%d", config.Get("server_port")), r); err != nil {
		panic(err)
	}
}
