package main

import (
	"go-demo/config/di"
	"go-demo/internal/cron"
	"time"

	"github.com/go-co-op/gocron"
)

func main() {
	s := gocron.NewScheduler(time.UTC)

	/* 计划任务路由 */
	if _, err := s.Cron("*/1 * * * *").Do(cron.UserCron.InitVip, 10); err != nil {
		di.Logger().Error(err.Error())
	}

	// starts the scheduler asynchronously
	s.StartAsync()
	// starts the scheduler and blocks current execution path
	s.StartBlocking()

}
