package task

import (
	"encoding/json"
	"go-demo/config/di"

	"github.com/gohouse/gorose/v2"
)

type userTask struct {
}

var UserTask userTask

func (userTask) AddUser(payload string) error {
	var payloadMap map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &payloadMap); err != nil {
		return err
	}
	_, err := di.Db().Table("t_users").Data(gorose.Data{"user_name": payloadMap["user_name"]}).Insert()
	return err
}
