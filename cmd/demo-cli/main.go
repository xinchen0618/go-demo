// 命令行入口
package main

import (
	"os"

	"go-demo/config/di"
	"go-demo/internal/action"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{ // cli 路由
			// DEMO
			{
				Name:  "user",
				Usage: "用户相关",
				Subcommands: []*cli.Command{
					{
						Name:   "add-user",
						Usage:  "创建一个用户",
						Action: action.User.AddUser,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		di.Logger().Error(err.Error())
		return
	}
}
