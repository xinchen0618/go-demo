package main

import (
	"go-demo/internal/cron"
	"time"

	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
)

func main() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		zap.L().Error(err.Error())
	}
	s := gocron.NewScheduler(loc)

	/* 计划任务路由 */
	if _, err = s.Cron("*/1 * * * *").Do(cron.User.InitVip, 10); err != nil {
		zap.L().Error(err.Error())
	}

	// starts the scheduler asynchronously
	s.StartAsync()
	// starts the scheduler and blocks current execution path
	s.StartBlocking()

}
