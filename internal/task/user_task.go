package task

import (
	"context"
	"encoding/json"
	"fmt"
	"go-demo/config/di"
	"go-demo/pkg/dbx"

	"github.com/hibiken/asynq"
)

type userTask struct{}

var UserTask userTask

func (userTask) AddUser(ctx context.Context, t *asynq.Task) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	_, err := dbx.Insert(di.Db(), "t_users", map[string]interface{}{"user_name": payload["user_name"]})
	return err
}

func (userTask) AddUserCounts(ctx context.Context, t *asynq.Task) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	_, err := dbx.Insert(di.Db(), "t_user_counts", map[string]interface{}{"user_id": payload["user_id"]})
	return err
}
