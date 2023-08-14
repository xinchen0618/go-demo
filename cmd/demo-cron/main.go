package main

import (
	"time"

	"go-demo/config/di"
	"go-demo/internal/cron"

	"github.com/go-co-op/gocron"
)

func main() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		di.Logger().Error(err.Error())
		return
	}
	s := gocron.NewScheduler(loc)

	// 计划任务路由 DEMO
	if _, err = s.Cron("* * * * *").Do(cron.User.DeleteUsers, 10); err != nil {
		di.Logger().Error(err.Error())
	}

	// starts the scheduler and blocks current execution path
	s.StartBlocking()
}
