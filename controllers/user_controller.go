package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gohouse/gorose/v2"
	"go-test/config"
	"log"
	"time"
)

func GetUsers(c *gin.Context) {
	key := "redis:users"
	var res []gorose.Data
	resCache, err := config.Cache.Get(config.Ctx, key).Result()
	if err == nil {    // 缓存存在
		err = json.Unmarshal([]byte(resCache), &res)
		if err != nil {
			log.Println(err)
			c.JSON(500, gin.H{"status": "InternalError", "message": "服务异常, 请稍后重试"})
			return
		}

	} else {    // 缓存不存在
		res, err = config.Db.Query("SELECT * FROM t_users LIMIT 12")
		if err != nil {
			log.Println(err)
			c.JSON(500, gin.H{"status": "InternalError", "message": "服务异常, 请稍后重试"})
			return
		}

		data, err := json.Marshal(res)
		if err != nil {
			log.Println(err)
			c.JSON(500, gin.H{"status": "InternalError", "message": "服务异常, 请稍后重试"})
			return
		}
		err = config.Cache.Set(config.Ctx, key, data, 10 * time.Second).Err()
		if err != nil {
			log.Println(err)
			c.JSON(500, gin.H{"status": "InternalError", "message": "服务异常, 请稍后重试"})
			return
		}
	}

	c.JSON(200, gin.H{
		"page": 1,
		"per_page": 12,
		"total_pages": 1,
		"total_counts": 1,
		"items": res,
	})
}