// RESTful API 入口
package main

import (
	"fmt"
	"time"

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
	// 实例化 Gin
	if lo.Contains([]string{"prod", "stage"}, config.GetRuntimeEnv()) {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	r.Use(
		middleware.Recovery(),                           // panic 处理
		middleware.CORS(),                               // 跨域处理
		middleware.QPSLimit(config.GetInt("qps_limit")), // 限流
		middleware.Timeout(time.Duration(config.GetInt("timeout"))*time.Second), // 超时控制
	)

	// 加载路由 DEMO
	router.Account(r)

	// 未知路由处理
	r.NoRoute(func(c *gin.Context) {
		ginx.Error(c, 404, "ResourceNotFound", "您请求的资源不存在")
	})

	// Run Gin
	addr := fmt.Sprintf(":%d", config.GetInt("server_port"))
	if err := endless.ListenAndServe(addr, r); err != nil {
		di.Logger().Error(err.Error())
		return
	}
}
