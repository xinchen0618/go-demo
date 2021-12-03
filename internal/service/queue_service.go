package service

import (
	"encoding/json"
	"go-demo/config/di"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type queueService struct{}

var QueueService queueService

// Enqueue 发送及时任务
//	@receiver queueService
//	@param taskName string
//	@param payload map[string]interface{}
//	@return error
func (queueService) Enqueue(taskName string, payload map[string]interface{}) error {
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
//	@receiver queueService
//	@param taskName string
//	@param payload map[string]interface{}
//	@return error
func (queueService) LowEnqueue(taskName string, payload map[string]interface{}) error {
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
//	@receiver queueService
//	@param taskName string
//	@param payload map[string]interface{}
//	@param delay time.Duration
//	@return error
func (queueService) EnqueueIn(taskName string, payload map[string]interface{}, delay time.Duration) error {
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
//	@receiver queueService
//	@param taskName string
//	@param payload map[string]interface{}
//	@param delay time.Duration
//	@return error
func (queueService) LowEnqueueIn(taskName string, payload map[string]interface{}, delay time.Duration) error {
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
