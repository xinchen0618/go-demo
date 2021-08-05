package action

import (
	"fmt"
	"go-demo/config/di"
	"strconv"

	"github.com/gohouse/gorose/v2"
	"github.com/urfave/cli/v2"
)

// 这里定义一个空结构体用于为大量的action方法做分类
type userAction struct {
}

// UserAction 这里不需要实例化, cli通过action.XxxAction.Xxx的形式引用旗下定义的方法
var UserAction *userAction

// InitPosition
//	@receiver *userAction
//	@param c *cli.Context
//	@return error
func (*userAction) InitPosition(c *cli.Context) error {
	counts := c.Args().Get(0)
	if "" == counts {
		counts = "10"
	}
	countsInt, err := strconv.Atoi(counts)
	if err != nil {
		di.Logger().Error(err.Error())
		return err
	}

	users, err := di.Db().Table("t_users").Fields("user_id").Limit(countsInt).Order("user_id").Get()
	if err != nil {
		di.Logger().Error(err.Error())
		return err
	}
	for key, user := range users {
		_, err = di.Db().Table("t_users").Where(gorose.Data{"user_id": user["user_id"]}).Data(gorose.Data{"position": 1024 * (key + 1)}).Update()
		if err != nil {
			di.Logger().Error(err.Error())
			return err
		}
	}
	fmt.Println("处理完毕")

	return nil
}
