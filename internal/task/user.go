// Package task 消息队列任务
package task

import (
	"context"
	"fmt"

	"go-demo/config/di"
	"go-demo/internal/model"
	"go-demo/pkg/gox"
	"go-demo/pkg/queuex"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// 用户相关消息队列 DEMO
type user struct{}

var User user

func (user) AddUser(ctx context.Context, t *asynq.Task) error {
	// 解析 payload
	var user struct {
		UserName string `json:"user_name"`
	}
	if err := queuex.Payload(t, &user); err != nil {
		return err
	}

	// 参数校验
	if user.UserName == "" {
		return fmt.Errorf("用户名不得为空. %w", asynq.SkipRetry)
	}

	// 业务处理
	userData := map[string]any{}
	if err := gox.Cast(user, &userData); err != nil {
		return err
	}
	if err := di.DemoDB().Model(&model.TUsers{}).Create(userData).Error; err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}
