package cron

import (
	"go-demo/config/di"

	"github.com/gohouse/gorose/v2"
)

// 这里定义一个空结构体用于为大量的cron方法做分类
type userCron struct {
}

// UserCron 这里不需要实例化, 计划任务通过cron.XxxCron.Xxx的形式引用旗下定义的方法
var UserCron userCron

// InitVip
//	@receiver *userCron
//	@param counts int
func (userCron) InitVip(counts int) {
	userIds, err := di.Db().Table("t_users").Where(gorose.Data{"is_vip": 0}).OrderBy("user_id").Limit(counts).Pluck("user_id")
	if err != nil {
		di.Logger().Error(err.Error())
	}
	_, err = di.Db().Table("t_users").WhereIn("user_id", userIds.([]interface{})).Update(gorose.Data{"is_vip": 1})
	if err != nil {
		di.Logger().Error(err.Error())
	}
}