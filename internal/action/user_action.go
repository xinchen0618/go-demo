package action

import (
	"fmt"
	"go-demo/config/di"
	"go-demo/pkg/dbx"

	"github.com/urfave/cli/v2"
)

// 这里定义一个空结构体用于为大量的action方法做分类
type userAction struct{}

// UserAction 这里仅需结构体零值, cli通过action.XxxAction.Xxx的形式引用旗下定义的方法
var UserAction userAction

// InitPosition
//	@receiver *userAction
//	@param c *cli.Context
//	@return error
func (userAction) InitPosition(c *cli.Context) error {
	counts := c.Args().Get(0)
	if "" == counts {
		counts = "10"
	}

	users, err := dbx.FetchAll(di.Db(), "SELECT user_id FROM t_users WHERE position=0 ORDER BY user_id LIMIT ?", counts)
	if err != nil {
		return err
	}
	for _, user := range users {
		userId := user["user_id"].(int64)
		if _, err := dbx.Update(di.Db(), "t_users", map[string]interface{}{"position": 1024 * userId}, "user_id=?", userId); err != nil {
			return err
		}
	}
	fmt.Println("处理完毕")

	return nil
}
