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
	countsInt, err := strconv.Atoi(counts)
	if err != nil {
		di.Logger().Error(err.Error())
		return err
	}

	users, err := di.Db().Table("t_users").Fields("user_id").Where(gorose.Data{"position": 0}).Limit(countsInt).Order("user_id").Get()
	if err != nil {
		di.Logger().Error(err.Error())
		return err
	}
	for _, user := range users {
		userId := user["user_id"].(int64)
		_, err = di.Db().Table("t_users").Where(gorose.Data{"user_id": userId}).Data(gorose.Data{"position": 1024 * userId}).Update()
		if err != nil {
			di.Logger().Error(err.Error())
			return err
		}
	}
	fmt.Println("处理完毕")

	return nil
}
