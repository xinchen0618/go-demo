package main

import (
	"errors"
	"fmt"
	"go-demo/config"
	"go-demo/internal/router"
	"go-demo/pkg/ginx"
	"go-demo/pkg/gox"
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
)

// recovery 主goroutine中panic兜底处理
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

// rateLimitMiddleware 限流
//	@param fillInterval time.Duration
//	@param quantum int64
//	@return gin.HandlerFunc
func rateLimit(fillInterval time.Duration, quantum int64) gin.HandlerFunc {
	bucket := ratelimit.NewBucketWithQuantum(fillInterval, quantum, quantum)
	return func(c *gin.Context) {
		if bucket.TakeAvailable(1) < 1 {
			ginx.Error(c, 429, "TooManyRequests", "服务繁忙, 请稍后重试")
			return
		}
		c.Next()
	}
}

// cors 跨域处理
//	@return gin.HandlerFunc
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("Origin") != "" {
			c.Header("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
			c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language, DNT, Keep-Alive, User-Agent, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			c.Header("Access-Control-Max-Age", "1728000")
			c.Header("Access-Control-Allow-Credentials", "false")
			if "OPTIONS" == c.Request.Method {
				ginx.Success(c, 200)
			}
		}
		c.Next()
	}
}

func main() {
	// 实例化gin
	if gox.InSlice(config.GetRuntimeEnv(), []string{"prod", "stage"}) {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Panic处理
	r.Use(recovery())
	// 限流
	r.Use(rateLimit(time.Second, int64(config.GetInt("rate_limit"))))
	// 跨域处理
	r.Use(cors())

	// 加载路由
	router.AccountRouter(r)

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
