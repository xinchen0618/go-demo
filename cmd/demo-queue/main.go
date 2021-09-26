package main

import (
	"go-demo/config"
	"go-demo/config/di"
	"go-demo/internal/task"
)

func main() {
	// Register tasks
	tasksMap := map[string]interface{}{
		"AddUser": task.UserTask.AddUser,
	}
	if err := di.QueueServer().RegisterTasks(tasksMap); err != nil {
		panic(err)
	}

	worker := di.QueueServer().NewWorker("queue_worker", config.GetInt("queue_worker"))
	if err := worker.Launch(); err != nil {
		panic(err)
	}
}
