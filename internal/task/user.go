package task

import (
	"context"
	"fmt"

	"go-demo/internal/service"
	"go-demo/pkg/queuex"

	"github.com/hibiken/asynq"
)

// 用户相关消息队列 DEMO
type user struct{}

var User user

func (user) AddUser(ctx context.Context, t *asynq.Task) error {
	var userData struct {
		UserName string `json:"user_name"`
	}
	if err := queuex.Payload(&userData, t); err != nil {
		return err
	}
	if "" == userData.UserName {
		return fmt.Errorf("用户名不得为空. %w", asynq.SkipRetry)
	}

	user := map[string]any{
		"user_name": userData.UserName,
	}
	_, err := service.User.CreateUser(user)
	return err
}
