package main

import (
	"fmt"
	"net/http"
	"os"

	"go-demo/di"
	"go-demo/router"
	"go-demo/util"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// recovery Panic处理
// 	程序初始化可以使用panic, 其他地方必须避免出现panic
//	@return gin.HandlerFunc
func recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				util.InternalError(c, err.(error))
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
			context.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			context.Header("Access-Control-Allow-Origin", "*") // 设置允许访问所有域
			context.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			context.Header("Access-Control-Allow-Headers", "X-Token, Authorization, Content-Length, Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language, DNT, Keep-Alive, User-Agent, If-Modified-Since, Cache-Control, Content-Type, Pragma")
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
	// 初始化Di
	di.Init()

	// Run gin
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

	if err := r.Run(fmt.Sprintf(":%d", viper.GetInt64("serverPort"))); err != nil {
		panic(err)
	}
}
