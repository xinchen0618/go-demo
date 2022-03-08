package service

import (
	"go-demo/config/di"
	"go-demo/pkg/dbx"
	"go-demo/pkg/gox"

	"github.com/gohouse/gorose/v2"
)

type user struct{}

var User user

// CreateUser 创建用户
//  @receiver user
//  @return int64
//  @return error
func (user) CreateUser(userData map[string]interface{}) (int64, error) {
	var userId int64
	err := di.DemoDb().Transaction(func(db gorose.IOrm) error { // 事务DEMO
		user, err := dbx.FetchOne(db, "SELECT user_id FROM t_users WHERE user_name=?", userData["user_name"])
		if err != nil {
			return err
		}
		if len(user) > 0 { // 记录存在
			userId = user["user_id"].(int64)
		} else { // 记录不存在
			userId, err = dbx.Insert(db, "t_users", userData)
			if err != nil {
				return err
			}
		}
		sql := "INSERT INTO t_user_counts(user_id,counts) VALUES(?,?) ON DUPLICATE KEY UPDATE counts = counts + 1"
		if _, err = dbx.Execute(db, sql, userId, gox.RandInt64(1, 9)); err != nil {
			return err
		}
		if err := Cache.Delete("t_user_counts", userId); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return userId, nil
}

// DeleteUser 删除用户
//  @receiver user
//  @param userId int64
//  @return error
func (user) DeleteUser(userId int64) error {
	err := di.DemoDb().Transaction(func(db gorose.IOrm) error {
		if _, err := dbx.Delete(db, "t_users", "user_id=?", userId); err != nil {
			return err
		}
		if err := Cache.Delete("t_users", userId); err != nil {
			return err
		}

		if _, err := dbx.Delete(db, "t_user_counts", "user_id=?", userId); err != nil {
			return err
		}
		if err := Cache.Delete("t_user_counts", userId); err != nil {
			return err
		}

		return nil
	})

	return err
}
