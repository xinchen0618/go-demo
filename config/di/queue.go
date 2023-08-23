// Package di 服务注入
package di

import (
	"fmt"
	"sync"

	"go-demo/config"

	"github.com/hibiken/asynq"
)

var (
	queueClient     *asynq.Client
	queueClientOnce sync.Once
)

// QueueClient 消息队列 client
func QueueClient() *asynq.Client {
	queueClientOnce.Do(func() {
		queueClient = asynq.NewClient(asynq.RedisClientOpt{
			Addr:     fmt.Sprintf("%s:%d", config.GetString("redis_host"), config.GetInt("redis_port")),
			DB:       config.GetInt("redis_index_queue"),
			Password: config.GetString("redis_auth"),
		})
	})

	return queueClient
}

var (
	queueServer     *asynq.Server
	queueServerOnce sync.Once
)

// QueueServer 消息队列 server
func QueueServer() *asynq.Server {
	queueServerOnce.Do(func() {
		queueServer = asynq.NewServer(
			asynq.RedisClientOpt{
				Addr:     fmt.Sprintf("%s:%d", config.GetString("redis_host"), config.GetInt("redis_port")),
				DB:       config.GetInt("redis_index_queue"),
				Password: config.GetString("redis_auth"),
			},
			asynq.Config{
				// Specify how many concurrent workers to use
				Concurrency: 100,
				// Optionally specify multiple queues with different priority.
				Queues: map[string]int{
					"default": 9,
					"low":     1,
				},
				// See the godoc for other configuration options
			},
		)
	})

	return queueServer
}
