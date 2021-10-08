package main

import (
	"context"
	"go-demo/config/di"
	"go-demo/internal/task"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

func loggingMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		start := time.Now()
		log.Printf("Start processing %q --- %s", t.Type(), t.Payload())
		err := h.ProcessTask(ctx, t)
		if err != nil {
			log.Printf("Processing err %q --- %s, %v", t.Type(), t.Payload(), err)
			return err
		}
		log.Printf("Finished processing %q: Elapsed Time = %v", t.Type(), time.Since(start))
		return nil
	})
}

func main() {
	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.Use(loggingMiddleware)
	mux.HandleFunc("user:AddUser", task.AddUser)
	mux.HandleFunc("user:AddUserCounts", task.AddUserCounts)

	if err := di.QueueServer().Run(mux); err != nil {
		panic(err)
	}
}
