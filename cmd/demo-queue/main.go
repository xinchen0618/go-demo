package main

import (
	"go-demo/config/di"
	"go-demo/internal/task"

	"github.com/hibiken/asynq"
)

func main() {
	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc("user:AddUser", task.AddUser)
	mux.HandleFunc("user:AddUserCounts", task.AddUserCounts)

	if err := di.QueueServer().Run(mux); err != nil {
		panic(err)
	}
}
