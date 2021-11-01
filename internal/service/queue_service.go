package service

import (
	"encoding/json"
	"go-demo/config/di"
	"time"

	"github.com/hibiken/asynq"
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
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	_, err = di.QueueClient().Enqueue(task)
	return err
}

// LowEnqueue 发送低优先级及时任务
//	@receiver queueService
//	@param taskName string
//	@param payload map[string]interface{}
//	@return error
func (queueService) LowEnqueue(taskName string, payload map[string]interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	_, err = di.QueueClient().Enqueue(task, asynq.Queue("low"))
	return err
}

// EnqueueIn 发送延时任务
//	@receiver queueService
//	@param taskName string
//	@param payload map[string]interface{}
//	@param delaySeconds int64
//	@return error
func (queueService) EnqueueIn(taskName string, payload map[string]interface{}, delaySeconds int64) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	_, err = di.QueueClient().Enqueue(task, asynq.ProcessIn(time.Second*time.Duration(delaySeconds)))
	return err
}

// LowEnqueueIn 发送低优先级延时任务
//	@receiver queueService
//	@param taskName string
//	@param payload map[string]interface{}
//	@param delaySeconds int64
//	@return error
func (queueService) LowEnqueueIn(taskName string, payload map[string]interface{}, delaySeconds int64) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	_, err = di.QueueClient().Enqueue(task, asynq.Queue("low"), asynq.ProcessIn(time.Second*time.Duration(delaySeconds)))
	return err
}
