package gox

import (
	"crypto/md5"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// Md5 字符串md5
//	@param str string
//	@return string
func Md5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// Md5x 非String数据md5
//	数据会先json.Marshal再Md5
//	@param i interface{}
//	@return string
//	@return error
func Md5x(i interface{}) (string, error) {
	iBytes, err := json.Marshal(i)
	if err != nil {
		zap.L().Error(err.Error())
		return "", err
	}
	return Md5(string(iBytes)), nil
}
