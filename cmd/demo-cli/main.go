package main

import (
	"go-demo/internal/action"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{ // cli路由
			{
				Name:  "user",
				Usage: "用户相关",
				Subcommands: []*cli.Command{
					{
						Name:   "init-position",
						Usage:  "初始化用户position字段",
						Action: action.User.InitPosition,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
