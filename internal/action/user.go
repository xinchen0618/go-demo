// Package action 命令行 action
package action

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"go-demo/config/di"
	"go-demo/internal/model"
)

// 用户相关命令行 DEMO 这里定义一个空结构体用于为大量的 action 方法做分类
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

	if err := di.DemoDB().Model(&model.TUsers{}).Create(map[string]any{"user_name": userName}).Error; err != nil {
		return err
	}
	fmt.Println("处理完毕")

	return nil
}
