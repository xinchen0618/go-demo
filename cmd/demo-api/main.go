package main

import (
	"fmt"

	"go-demo/config"
	"go-demo/internal/middleware"
	"go-demo/internal/router"
	"go-demo/pkg/ginx"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slices"
)

func main() {
	// 实例化gin
	if slices.Contains([]string{"prod", "stage"}, config.GetRuntimeEnv()) {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// panic处理/跨域处理/限流
	r.Use(middleware.Recovery(), middleware.Cors(), middleware.QpsLimit(config.GetInt("qps_limit")))

	// 加载路由 DEMO
	router.Account(r)

	// 未知路由处理
	r.NoRoute(func(c *gin.Context) {
		ginx.Error(c, 404, "ResourceNotFound", "您请求的资源不存在")
		return
	})

	// Run gin
	addr := fmt.Sprintf(":%d", config.Get("server_port"))
	if err := endless.ListenAndServe(addr, r); err != nil {
		panic(err)
	}
}
