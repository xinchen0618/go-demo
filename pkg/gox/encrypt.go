package gox

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"

	"go.uber.org/zap"
)

// Md5 字符串md5
//	@param str string
//	@return string
//	@return error
func Md5(s string) (string, error) {
	h := md5.New()
	if _, err := io.WriteString(h, s); err != nil {
		zap.L().Error(err.Error())
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// Md5x 非String/Number数据md5
//	@param i interface{}
//	@return string
//	@return error
func Md5x(i interface{}) (string, error) {
	iBytes, err := json.Marshal(i)
	if err != nil {
		zap.L().Error(err.Error())
		return "", err
	}
	return Md5(string(iBytes))
}
