package cron

import (
	"go-demo/config/di"
	"go-demo/pkg/dbx"

	"github.com/gohouse/gorose/v2"
	"go.uber.org/zap"
)

// 这里定义一个空结构体用于为大量的cron方法做分类
type user struct{}

// User 这里仅需结构体零值
var User user

// InitVip
//	@receiver user
//	@param counts int
func (user) InitVip(counts int) {
	userIds, err := dbx.FetchColumn(di.Db(), "SELECT user_id FROM t_users WHERE is_vip=0 LIMIT ?", counts)
	if err != nil {
		return
	}
	if _, err = di.Db().Table("t_users").WhereIn("user_id", userIds).Update(gorose.Data{"is_vip": 1}); err != nil {
		zap.L().Error(err.Error())
	}
}
