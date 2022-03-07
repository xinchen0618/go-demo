package cron

import (
	"fmt"
	"go-demo/config/di"
	"go-demo/pkg/dbx"

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
	userIds, err := dbx.FetchColumn(di.DemoDb(), "SELECT user_id FROM t_users WHERE is_vip=0 LIMIT ?", counts)
	if err != nil {
		return
	}
	where := fmt.Sprintf("user_id IN(%s)", dbx.Slice2in(userIds))
	if _, err := dbx.Update(di.DemoDb(), "t_users", map[string]interface{}{"is_vip": 1}, where); err != nil {
		zap.L().Error(err.Error())
	}
}
