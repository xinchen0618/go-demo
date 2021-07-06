package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"go-test/router"

	"github.com/gin-gonic/gin"
)

// Panic处理
// 与业务无关的错误, 使用panic, 记录错误日志并统一向客户端返回500错误
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

func main() {
	/* Run gin */
	runtimeEnv := os.Getenv("RUNTIME_ENV")
	if runtimeEnv == "" || runtimeEnv == "prod" || runtimeEnv == "stage" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Panic处理
	r.Use(recovery())

	// 加载路由
	router.LoadAccount(r)

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
