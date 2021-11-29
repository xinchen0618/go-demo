package gox

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
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
	return Md5(string(iBytes))
}

// GobEncode Gob编码
//	@param v interface{}
//	@return []byte
//	@return error
func GobEncode(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		zap.L().Error(err.Error())
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

// GobDecode Gob解码
//	@param b []byte
//	@param result interface{}
//	@return error
func GobDecode(b []byte, result interface{}) error {
	buf := bytes.NewBuffer(b)
	enc := gob.NewDecoder(buf)

	if err := enc.Decode(result); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}
