package di

import (
	"fmt"
	"go-demo/config"
	"sync"

	"github.com/hibiken/asynq"
)

var (
	queueClient     *asynq.Client
	queueClientOnce sync.Once
)

func QueueClient() *asynq.Client {
	queueClientOnce.Do(func() {
		redisAddr := fmt.Sprintf("%s:%d", config.GetString("redis_host"), config.GetInt("redis_port"))
		queueClient = asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr, DB: config.GetInt("redis_index_queue")})
	})

	return queueClient
}

var (
	queueServer     *asynq.Server
	queueServerOnce sync.Once
)

func QueueServer() *asynq.Server {
	queueServerOnce.Do(func() {
		redisAddr := fmt.Sprintf("%s:%d", config.GetString("redis_host"), config.GetInt("redis_port"))
		queueServer = asynq.NewServer(
			asynq.RedisClientOpt{Addr: redisAddr, DB: config.GetInt("redis_index_queue")},
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
