package main

import (
	"errors"
	"go-demo/config/di"
	"go-demo/internal/task"
	"os"
)

func main() {
	queueLevel := os.Getenv("QUEUE_LEVEL")
	if "" == queueLevel {
		queueLevel = "default"
	}

	// Register tasks
	tasksMap := map[string]interface{}{
		"AddUser":       task.UserTask.AddUser,
		"AddUserCounts": task.UserTask.AddUserCounts,
	}
	if "default" == queueLevel {
		if err := di.QueueServer().RegisterTasks(tasksMap); err != nil {
			panic(err)
		}
		worker := di.QueueServer().NewWorker("default_queue_worker", 1000)
		if err := worker.Launch(); err != nil {
			panic(err)
		}

	} else if "low" == queueLevel {
		if err := di.LowQueueServer().RegisterTasks(tasksMap); err != nil {
			panic(err)
		}
		worker := di.LowQueueServer().NewWorker("low_queue_worker", 3)
		if err := worker.Launch(); err != nil {
			panic(err)
		}

	} else {
		panic(errors.New("未知队列优先级"))
	}

}
