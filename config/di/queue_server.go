package di

import (
	"fmt"
	"go-demo/config"
	"sync"

	"github.com/RichardKnop/machinery/v2"
	redisbackend "github.com/RichardKnop/machinery/v2/backends/redis"
	redisbroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	redisconfig "github.com/RichardKnop/machinery/v2/config"
	eagerlock "github.com/RichardKnop/machinery/v2/locks/eager"
)

var (
	queueServer     *machinery.Server
	queueServerOnce sync.Once
)

// QueueServer
//	@return *machinery.Server
func QueueServer() *machinery.Server {
	queueServerOnce.Do(func() {
		cnf := &redisconfig.Config{
			DefaultQueue:    "default_queue",
			ResultsExpireIn: 3600,
			Redis: &redisconfig.RedisConfig{
				MaxIdle:                3,
				IdleTimeout:            240,
				ReadTimeout:            15,
				WriteTimeout:           15,
				ConnectTimeout:         15,
				NormalTasksPollPeriod:  1000,
				DelayedTasksPollPeriod: 500,
			},
		}

		// Create server instance
		broker := redisbroker.NewGR(cnf, []string{fmt.Sprintf("%s:%d", config.GetString("redis_host"), config.GetInt("redis_port"))}, config.GetInt("redis_index_queue"))
		backend := redisbackend.NewGR(cnf, []string{fmt.Sprintf("%s:%d", config.GetString("redis_host"), config.GetInt("redis_port"))}, config.GetInt("redis_index_queue"))
		lock := eagerlock.New()
		queueServer = machinery.NewServer(cnf, broker, backend, lock)
	})

	return queueServer
}

var (
	lowQueueServer     *machinery.Server
	lowQueueServerOnce sync.Once
)

func LowQueueServer() *machinery.Server {
	lowQueueServerOnce.Do(func() {
		cnf := &redisconfig.Config{
			DefaultQueue:    "low_queue",
			ResultsExpireIn: 3600,
			Redis: &redisconfig.RedisConfig{
				MaxIdle:                3,
				IdleTimeout:            240,
				ReadTimeout:            15,
				WriteTimeout:           15,
				ConnectTimeout:         15,
				NormalTasksPollPeriod:  1000,
				DelayedTasksPollPeriod: 500,
			},
		}

		// Create server instance
		broker := redisbroker.NewGR(cnf, []string{fmt.Sprintf("%s:%d", config.GetString("redis_host"), config.GetInt("redis_port"))}, config.GetInt("redis_index_queue"))
		backend := redisbackend.NewGR(cnf, []string{fmt.Sprintf("%s:%d", config.GetString("redis_host"), config.GetInt("redis_port"))}, config.GetInt("redis_index_queue"))
		lock := eagerlock.New()
		lowQueueServer = machinery.NewServer(cnf, broker, backend, lock)
	})

	return lowQueueServer
}
