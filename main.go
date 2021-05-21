package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-test/routers"
	"log"
	"runtime"
)

/* Panic处理 */
func recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log
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
	r := gin.Default()

	// Panic处理
	r.Use(recovery())

	// 加载路由
	routers.LoadUser(r)

	if err := r.Run(); err != nil {
		panic(err)
	}
}
