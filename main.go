package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"go-test/di"
	"go-test/router"

	"github.com/gin-gonic/gin"
)

// recovery Panic处理
// 	主goroutine与业务无关的错误, 使用panic, 记录错误日志并统一向客户端返回500错误
//	@return gin.HandlerFunc
func recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录错误日志
				buf := make([]byte, 2048)
				n := runtime.Stack(buf, false)
				stackInfo := fmt.Sprintf("%s", buf[:n])
				log.Printf("%v\n%v\n", err, stackInfo)

				c.JSON(500, gin.H{"status": "InternalError", "message": "服务异常, 请稍后重试"})
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

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
