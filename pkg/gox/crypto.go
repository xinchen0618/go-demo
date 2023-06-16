package gox

import (
	"crypto/md5"
	"fmt"
	"strconv"

	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

// Md5 字符串md5
func Md5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// Md5x 非String数据md5
//
//	数据会先json.Marshal再Md5.
func Md5x(a any) (string, error) {
	iBytes, err := json.Marshal(a)
	if err != nil {
		zap.L().Error(err.Error())
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(iBytes)), nil
}

// PasswordHash 创建密码的散列
//
//	返回38位16进制字符串.
func PasswordHash(password string) string {
	salt := strconv.FormatInt(RandInt64(0x100000, 0xffffff), 16) // 6位16进制字符串
	return salt + Md5(password+Md5(password+salt)+salt)
}

// PasswordVerify 验证密码与散列是否匹配
func PasswordVerify(password, passwordHash string) bool {
	salt := passwordHash[0:6]
	return passwordHash == salt+Md5(password+Md5(password+salt)+salt)
}
