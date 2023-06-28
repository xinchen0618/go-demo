package service

import (
	"go-demo/config/di"
	"go-demo/pkg/dbx"
)

// 用户相关原子级操作 DEMO
type user struct{}

var User user

// CreateUser 创建用户
//
//	userData 用户信息键值对.
//	成功返回用户id.
func (user) CreateUser(userData map[string]any) (int64, error) {
	user, err := dbx.FetchOne(di.DemoDB(), "SELECT user_id FROM t_users WHERE user_name=?", userData["user_name"])
	if err != nil {
		return 0, err
	}
	if len(user) > 0 { // 记录存在
		return user["user_id"].(int64), nil
	}

	return dbx.Insert(di.DemoDB(), "t_users", userData)
}
