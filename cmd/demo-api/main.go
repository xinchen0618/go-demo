package main

import (
	"fmt"

	"go-demo/config"
	"go-demo/config/di"
	"go-demo/internal/middleware"
	"go-demo/internal/router"
	"go-demo/pkg/ginx"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

func main() {
	// 实例化gin
	if lo.Contains([]string{"prod", "stage"}, config.GetRuntimeEnv()) {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	r.Use(
		middleware.Recovery(),                           // panic处理
		middleware.Cors(),                               // 跨域处理
		middleware.QpsLimit(config.GetInt("qps_limit")), // 限流
	)

	// 加载路由 DEMO
	router.Account(r)

	// 未知路由处理
	r.NoRoute(func(c *gin.Context) {
		ginx.Error(c, 404, "ResourceNotFound", "您请求的资源不存在")
	})

	// Run gin
	addr := fmt.Sprintf(":%d", config.Get("server_port"))
	if err := endless.ListenAndServe(addr, r); err != nil {
		di.Logger().Error(err.Error())
		return
	}
}
