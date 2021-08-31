package main

import (
	"go-demo/internal/cron"
	"time"

	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
)

func main() {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		zap.L().Error(err.Error())
	}
	s := gocron.NewScheduler(location)

	/* 计划任务路由 */
	if _, err = s.Cron("*/1 * * * *").Do(cron.UserCron.InitVip, 10); err != nil {
		zap.L().Error(err.Error())
	}

	// starts the scheduler asynchronously
	s.StartAsync()
	// starts the scheduler and blocks current execution path
	s.StartBlocking()

}
