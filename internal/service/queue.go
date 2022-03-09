package service

import (
	"encoding/json"
	"time"

	"go-demo/config/di"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type queue struct{}

var Queue queue

// Enqueue 发送及时任务
//	@receiver queue
//	@param taskName string
//	@param payload map[string]interface{}
//	@return error
func (queue) Enqueue(taskName string, payload map[string]interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error(err.Error())
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	if _, err := di.QueueClient().Enqueue(task); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}

// LowEnqueue 发送低优先级及时任务
//	@receiver queue
//	@param taskName string
//	@param payload map[string]interface{}
//	@return error
func (queue) LowEnqueue(taskName string, payload map[string]interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error(err.Error())
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	if _, err := di.QueueClient().Enqueue(task, asynq.Queue("low")); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}

// EnqueueIn 发送延时任务
//	@receiver queue
//	@param taskName string
//	@param payload map[string]interface{}
//	@param delay time.Duration
//	@return error
func (queue) EnqueueIn(taskName string, payload map[string]interface{}, delay time.Duration) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error(err.Error())
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	if _, err := di.QueueClient().Enqueue(task, asynq.ProcessIn(delay)); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}

// LowEnqueueIn 发送低优先级延时任务
//	@receiver queue
//	@param taskName string
//	@param payload map[string]interface{}
//	@param delay time.Duration
//	@return error
func (queue) LowEnqueueIn(taskName string, payload map[string]interface{}, delay time.Duration) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error(err.Error())
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	if _, err := di.QueueClient().Enqueue(task, asynq.Queue("low"), asynq.ProcessIn(delay)); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}
