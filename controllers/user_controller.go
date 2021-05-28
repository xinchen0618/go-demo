package controllers

import (
	"github.com/gin-gonic/gin"
	"go-test/di"
	"go-test/services"
	"go-test/utils"
)

func GetUsers(c *gin.Context) {
	// 登录校验
	if _, err := services.CheckUserLogin(c); err != nil {
		return
	}

	res, err := utils.GetPageItems(map[string]interface{}{
		"ginContext": c,
		"db":         di.Db(),
		"select":     "u.user_id,u.user_name,u.money,u.created_at,u.updated_at,uc.counts",
		"from":       "t_users AS u JOIN t_user_counts AS uc ON u.user_id = uc.user_id",
		"where":      "u.user_id > ?",
		"bindParams": []interface{}{5},
		"orderBy":    "user_id DESC",
	})
	if err != nil {
		return
	}
	c.JSON(200, res)
}
