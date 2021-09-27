package service

import (
	"encoding/json"
	"go-demo/config/di"
	"time"

	"github.com/hibiken/asynq"
)

type queueService struct {
}

var QueueService queueService

func (queueService) Enqueue(taskName string, payload map[string]interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	_, err = di.QueueClient().Enqueue(task)
	return err
}

func (queueService) LowEnqueue(taskName string, payload map[string]interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	_, err = di.QueueClient().Enqueue(task, asynq.Queue("low"))
	return err
}

func (queueService) EnqueueIn(taskName string, payload map[string]interface{}, delaySeconds int64) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	_, err = di.QueueClient().Enqueue(task, asynq.ProcessIn(time.Second*time.Duration(delaySeconds)))
	return err
}

func (queueService) LowEnqueueIn(taskName string, payload map[string]interface{}, delaySeconds int64) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskName, payloadBytes)
	_, err = di.QueueClient().Enqueue(task, asynq.Queue("low"), asynq.ProcessIn(time.Second*time.Duration(delaySeconds)))
	return err
}
