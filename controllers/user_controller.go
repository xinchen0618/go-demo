package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gohouse/gorose/v2"
	"go-test/di"
	"time"
)

func GetUsers(c *gin.Context) {
	key := "redis:users"
	var res []gorose.Data
	resCache, err := di.Cache.Get(di.Ctx, key).Result()
	if err == nil { // 缓存存在
		err = json.Unmarshal([]byte(resCache), &res)
		if err != nil {
			panic(err)
		}

	} else { // 缓存不存在
		res, err = di.Db.Query("SELECT * FROM t_users LIMIT 12")
		if err != nil {
			panic(err)
		}

		data, err := json.Marshal(res)
		if err != nil {
			panic(err)
		}
		err = di.Cache.Set(di.Ctx, key, data, 10*time.Second).Err()
		if err != nil {
			panic(err)
		}
	}

	c.JSON(200, gin.H{
		"page":         1,
		"per_page":     12,
		"total_pages":  1,
		"total_counts": 1,
		"items":        res,
	})
}
