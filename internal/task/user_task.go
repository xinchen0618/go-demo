package task

import (
	"context"
	"encoding/json"
	"fmt"
	"go-demo/config/di"

	"github.com/gohouse/gorose/v2"
	"github.com/hibiken/asynq"
)

func AddUser(ctx context.Context, t *asynq.Task) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	_, err := di.Db().Table("t_users").Data(gorose.Data{"user_name": payload["user_name"]}).Insert()
	return err
}
