package main

import (
	"github.com/gin-gonic/gin"
	"go-test/routers"
	"log"
)

func main() {
	/* Recover */
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	/* Run gin */
	r := gin.Default()

	// 加载路由
	routers.LoadUser(r)

	if err := r.Run(); err != nil {
		panic(err)
	}
}