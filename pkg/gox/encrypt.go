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
func Md5(str string) string {
	h := md5.New()
	if _, err := io.WriteString(h, str); err != nil {
		zap.L().Error(err.Error())
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Md5Interface 非String/Number数据md5
//	先json.Marshal再Md5
//	@param i interface{}
//	@return string
func Md5Interface(i interface{}) string {
	iBytes, err := json.Marshal(i)
	if err != nil {
		zap.L().Error(err.Error())
		return ""
	}
	return Md5(string(iBytes))
}
