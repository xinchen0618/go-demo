package main

import (
	"context"
	"log"
	"time"

	"go-demo/config/di"
	"go-demo/internal/task"

	"github.com/hibiken/asynq"
)

func loggingMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		start := time.Now()
		log.Printf("Start processing %q --- %s", t.Type(), t.Payload())

		if err := h.ProcessTask(ctx, t); err != nil {
			log.Printf("Processing err %q --- %s, %v", t.Type(), t.Payload(), err)
			return err
		}

		log.Printf("Finished processing %q --- %s, Elapsed Time = %v", t.Type(), t.Payload(), time.Since(start))
		return nil
	})
}

func main() {
	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.Use(loggingMiddleware)

	// 注册队列任务 DEMO
	mux.HandleFunc("User:AddUser", task.User.AddUser)

	if err := di.QueueServer().Run(mux); err != nil {
		panic(err)
	}
}
