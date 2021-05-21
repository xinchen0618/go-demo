package main

import (
	"github.com/gin-gonic/gin"
	"go-test/routers"
	"log"
)

/* Panic处理 */
func recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
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
