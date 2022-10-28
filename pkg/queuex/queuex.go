package queuex

import (
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Enqueue 发送及时任务
//
//	@param client *asynq.Client
//	@param taskName string
//	@param payload map[string]any
//	@return error
func Enqueue(client *asynq.Client, taskName string, payload map[string]any) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error(err.Error())
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	if _, err := client.Enqueue(task); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}

// LowEnqueue 发送低优先级及时任务
//
//	@param client *asynq.Client
//	@param taskName string
//	@param payload map[string]any
//	@return error
func LowEnqueue(client *asynq.Client, taskName string, payload map[string]any) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error(err.Error())
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	if _, err := client.Enqueue(task, asynq.Queue("low")); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}

// EnqueueIn 发送延时任务
//
//	@param client *asynq.Client
//	@param taskName string
//	@param payload map[string]any
//	@param delay time.Duration
//	@return error
func EnqueueIn(client *asynq.Client, taskName string, payload map[string]any, delay time.Duration) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error(err.Error())
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	if _, err := client.Enqueue(task, asynq.ProcessIn(delay)); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}

// EnqueueAt 发送定时任务
//
//	@param client *asynq.Client
//	@param taskName string
//	@param payload map[string]any
//	@param timeAt time.Time
//	@return error
func EnqueueAt(client *asynq.Client, taskName string, payload map[string]any, timeAt time.Time) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error(err.Error())
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	if _, err := client.Enqueue(task, asynq.ProcessAt(timeAt)); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}

// LowEnqueueIn 发送低优先级延时任务
//
//	@param client *asynq.Client
//	@param taskName string
//	@param payload map[string]any
//	@param delay time.Duration
//	@return error
func LowEnqueueIn(client *asynq.Client, taskName string, payload map[string]any, delay time.Duration) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error(err.Error())
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	if _, err := client.Enqueue(task, asynq.Queue("low"), asynq.ProcessIn(delay)); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}

// LowEnqueueAt 发送低优先级定时任务
//
//	@param client *asynq.Client
//	@param taskName string
//	@param payload map[string]any
//	@param timeAt time.Time
//	@return error
func LowEnqueueAt(client *asynq.Client, taskName string, payload map[string]any, timeAt time.Time) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error(err.Error())
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	if _, err := client.Enqueue(task, asynq.Queue("low"), asynq.ProcessAt(timeAt)); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}

// Payload 从Task中解析Payload
//
//	@param t *asynq.Task
//	@param p any 接收结果的指针, map指针或者struct指针皆可
//	@return error 解析失败返回的是SkipRetry的包裹, task方法中返回这个error将不再重试
func Payload(t *asynq.Task, p any) error {
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		zap.L().Error(err.Error())
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	return nil
}
