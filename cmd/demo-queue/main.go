// 消息队列入口
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
			di.Logger().Error(err.Error())
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

	// register handler DEMO
	mux.HandleFunc("User:AddUser", task.User.AddUser)

	// run queue server
	if err := di.QueueServer().Run(mux); err != nil {
		di.Logger().Error(err.Error())
		return
	}
}
