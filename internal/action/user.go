package action

import (
	"fmt"

	"go-demo/internal/service"

	"github.com/urfave/cli/v2"
)

// 用户相关命令行 DEMO 这里定义一个空结构体用于为大量的action方法做分类
type user struct{}

// User 这里仅需结构体零值
var User user

// AddUser 添加一个用户
func (user) AddUser(c *cli.Context) error {
	userName := c.Args().Get(0)
	if userName == "" {
		fmt.Println("请输入用户名")
		return nil
	}

	userData := map[string]any{
		"user_name": userName,
	}
	_, err := service.User.CreateUser(userData)
	if err != nil {
		return err
	}
	fmt.Println("处理完毕")

	return nil
}
