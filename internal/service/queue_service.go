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

// Enqueue
//	@receiver queueService
//	@param taskName string
//	@param payload map[string]interface{}
//	@return error
func (queueService) Enqueue(taskName string, payload map[string]interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error(err.Error())
	}
	signature := &tasks.Signature{
		Name: taskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: string(payloadBytes),
			},
		},
		RetryCount: 23,
	}

	if _, err := di.QueueServer().SendTask(signature); err != nil {
		return err
	}

	return nil
}

// EnqueueIn
//	@receiver queueService
//	@param taskName string
//	@param payload map[string]interface{}
//	@param delay int64
//	@return error
func (queueService) EnqueueIn(taskName string, payload map[string]interface{}, delay int64) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error(err.Error())
	}
	eta := time.Now().UTC().Add(time.Second * time.Duration(delay))
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

	if _, err := di.QueueServer().SendTask(signature); err != nil {
		return err
	}

	return nil
}
