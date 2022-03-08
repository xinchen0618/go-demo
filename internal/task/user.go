package task

import (
	"context"
	"encoding/json"
	"fmt"
	"go-demo/internal/service"

	"github.com/hibiken/asynq"
)

type user struct{}

var User user

func (user) AddUser(ctx context.Context, t *asynq.Task) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	userData := map[string]interface{}{
		"user_name": payload["user_name"],
	}
	_, err := service.User.CreateUser(userData)
	return err
}
