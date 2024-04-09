// 计划任务入口
package main

import (
	"time"

	"go-demo/config/di"
	"go-demo/internal/cron"

	"github.com/go-co-op/gocron/v2"
)

func main() {
	// create a scheduler
	s, err := gocron.NewScheduler()
	if err != nil {
		di.Logger().Error(err.Error())
		return
	}
	defer func() {
		// when you're done, shut it down
		if err := s.Shutdown(); err != nil {
			di.Logger().Error(err.Error())
		}
	}()

	// add a job to the scheduler
	if _, err := s.NewJob(
		gocron.DurationJob(10*time.Second),
		gocron.NewTask(cron.User.DeleteUsers, 10),
	); err != nil {
		di.Logger().Error(err.Error())
	}

	// start the scheduler
	s.Start()

	// block until you are ready to shut down
	select {}
}
