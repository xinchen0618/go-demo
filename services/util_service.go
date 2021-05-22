package services

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// GenToken 生成一个token字符串
func GenToken() string {
	seed := strconv.FormatInt(time.Now().UnixNano(), 10) + strconv.Itoa(rand.Int())

	return fmt.Sprintf("%x", md5.Sum([]byte(seed)))
}
