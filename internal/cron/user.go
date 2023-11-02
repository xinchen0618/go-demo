// Package cron 计划任务
package cron

import (
	"go.uber.org/zap"

	"go-demo/config/di"
	"go-demo/internal/model"
)

// 用户相关计划任务 DEMO 这里定义一个空结构体用于为大量的 cron 方法做分类
type user struct{}

// User 这里仅需结构体零值
var User user

// DeleteUsers 批量删除用户
//
//	userCount 为需要删除的数量.
func (user) DeleteUsers(userCount int) {
	userIDs := make([]int64, 0)
	if err := di.DemoDB().Model(&model.TUsers{}).Select("user_id").Limit(userCount).Find(&userIDs).Error; err != nil {
		return
	}
	if err := di.DemoDB().Where("user_id IN ?", userIDs).Delete(&model.TUsers{}).Error; err != nil {
		zap.L().Error(err.Error())
	}
}
