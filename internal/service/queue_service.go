package service

import (
	"encoding/json"
	"go-demo/config/di"
	"time"

	"github.com/RichardKnop/machinery/v2/tasks"

	"go.uber.org/zap"
)

type queueService struct {
}

var QueueService queueService

// Enqueue 入队默认优先级队列
//	@receiver queueService
//	@param taskName string
//	@param payload map[string]interface{}
//	@return error
func (queueService) Enqueue(taskName string, payload map[string]interface{}) error {
	signature, err := QueueService.newSignature(taskName, payload, 0)
	if err != nil {
		return err
	}

	if _, err := di.QueueServer().SendTask(signature); err != nil {
		return err
	}

	return nil
}

// LowEnqueue 入队低优先级队列
//	@receiver queueService
//	@param taskName string
//	@param payload map[string]interface{}
//	@return error
func (queueService) LowEnqueue(taskName string, payload map[string]interface{}) error {
	signature, err := QueueService.newSignature(taskName, payload, 0)
	if err != nil {
		return err
	}

	if _, err := di.LowQueueServer().SendTask(signature); err != nil {
		return err
	}

	return nil
}

// EnqueueIn 延迟入队默认优先级队列
//	@receiver queueService
//	@param taskName string
//	@param payload map[string]interface{}
//	@param delaySeconds int64
//	@return error
func (queueService) EnqueueIn(taskName string, payload map[string]interface{}, delaySeconds int64) error {
	signature, err := QueueService.newSignature(taskName, payload, delaySeconds)
	if err != nil {
		return err
	}

	if _, err := di.QueueServer().SendTask(signature); err != nil {
		return err
	}

	return nil
}

// LowEnqueueIn 延迟入队低优先级队列
//	@receiver queueService
//	@param taskName string
//	@param payload map[string]interface{}
//	@param delaySeconds int64
//	@return error
func (queueService) LowEnqueueIn(taskName string, payload map[string]interface{}, delaySeconds int64) error {
	signature, err := QueueService.newSignature(taskName, payload, delaySeconds)
	if err != nil {
		return err
	}

	if _, err := di.LowQueueServer().SendTask(signature); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}

// newSignature
//	@receiver queueService
//	@param taskName string
//	@param payload map[string]interface{}
//	@param delaySeconds int64
//	@return *tasks.Signature
//	@return error
func (queueService) newSignature(taskName string, payload map[string]interface{}, delaySeconds int64) (*tasks.Signature, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error(err.Error())
		return nil, err
	}
	eta := time.Now().UTC().Add(time.Second * time.Duration(delaySeconds))
	signature := &tasks.Signature{
		Name: taskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: string(payloadBytes),
			},
		},
		ETA:        &eta,
		RetryCount: 23,
	}

	return signature, nil
}
