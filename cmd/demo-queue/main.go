package main

import (
	"go-demo/config/di"
	"go-demo/internal/task"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

func main() {
	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc("AddUser", task.AddUser)
	mux.HandleFunc("AddUserCounts", task.AddUserCounts)

	if err := di.QueueServer().Run(mux); err != nil {
		zap.L().Error(err.Error())
		panic(err)
	}
}
