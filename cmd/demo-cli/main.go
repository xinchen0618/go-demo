package main

import (
	"log"
	"os"

	"go-demo/internal/action"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{ // cli路由
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
		log.Fatal(err)
	}
}
