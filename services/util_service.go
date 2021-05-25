package services

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// GenToken 生成一个token字符串
func GenToken() string {
	seed := strconv.FormatInt(time.Now().UnixNano(), 10) + strconv.Itoa(rand.Int())

	return fmt.Sprintf("%x", md5.Sum([]byte(seed)))
}

// GetJsonBody 获取Json参数
// pattern paramKey:paramName:paramType:paramPattern
// valuePattern +必填不可为空, *选填可以为空, ?选填不可为空
func GetJsonBody(c *gin.Context, patterns []string) (res map[string]interface{}, resErr error) {
	jsonBody := make(map[string]interface{})
	_ = c.ShouldBindJSON(&jsonBody) // 这里的error不要处理, 因为空body会报error
	res = make(map[string]interface{})

	for _, pattern := range patterns {
		patternAtoms := strings.Split(pattern, ":")
		required := true
		allowEmpty := false
		if "+" == patternAtoms[3] {
			required = true
			allowEmpty = false
		} else if "*" == patternAtoms[3] {
			required = false
			allowEmpty = true
		} else if "?" == patternAtoms[3] {
			required = false
			allowEmpty = false
		}

		paramValue, ok := jsonBody[patternAtoms[0]]
		if !ok {
			if required {
				c.JSON(400, gin.H{"status": "emptyParam", "message": fmt.Sprintf("%s不得为空", patternAtoms[1])})
				resErr = errors.New("emptyParam")
				return
			} else {
				continue
			}
		}

		res[patternAtoms[0]], resErr = FilterParam(c, patternAtoms[1], paramValue, patternAtoms[2], allowEmpty)
		if resErr != nil {
			return
		}
	}

	return
}

// GetQueries 获取Query参数
// pattern paramKey:paramName:paramType:defaultValue defaultValue为nil时参数必填
func GetQueries(c *gin.Context, patterns []string) (res map[string]interface{}, resErr error) {
	res = make(map[string]interface{})

	for _, pattern := range patterns {
		patternAtoms := strings.Split(pattern, ":")
		allowEmpty := false
		if `""` == patternAtoms[3] { // 默认值""表示空字符串
			patternAtoms[3] = ""
			allowEmpty = true
		}
		paramValue := c.Query(patternAtoms[0])
		if "" == paramValue {
			if "nil" == patternAtoms[3] { // 必填
				c.JSON(400, gin.H{"status": "emptyParam", "message": fmt.Sprintf("%s不得为空", patternAtoms[1])})
				resErr = errors.New("emptyParam")
				return
			} else {
				paramValue = patternAtoms[3]
			}
		}

		res[patternAtoms[0]], resErr = FilterParam(c, patternAtoms[1], paramValue, patternAtoms[2], allowEmpty)
		if resErr != nil {
			return
		}
	}

	return
}

// FilterParam 校验参数类型
// paramType int整型, +int正整型, !-int非负整型, string字符串, []枚举, array数组
func FilterParam(c *gin.Context, paramName string, paramValue interface{}, paramType string, allowEmpty bool) (resValue interface{}, resErr error) {
	valueType := reflect.TypeOf(paramValue).String()

	if "int" == paramType[len(paramType)-3:] { // 整型
		var intValue int
		var err error
		if "string" == valueType {
			paramValue = strings.TrimSpace(paramValue.(string))
			if "" == paramValue && !allowEmpty {
				c.JSON(400, gin.H{"status": "emptyParam", "message": fmt.Sprintf("%s不得为空", paramName)})
				resErr = errors.New("emptyParam")
				return
			}
			intValue, err = strconv.Atoi(paramValue.(string))
			if err != nil {
				c.JSON(400, gin.H{"status": "InvalidParam", "message": fmt.Sprintf("%s不正确", paramName)})
				resErr = errors.New("InvalidParam")
				return
			}
		} else if "float64" == valueType {
			stringValue := fmt.Sprintf("%v", paramValue)
			intValue, err = strconv.Atoi(stringValue)
			if err != nil {
				c.JSON(400, gin.H{"status": "InvalidParam", "message": fmt.Sprintf("%s不正确", paramName)})
				resErr = errors.New("InvalidParam")
				return
			}
		} else {
			c.JSON(400, gin.H{"status": "InvalidParam", "message": fmt.Sprintf("%s不正确", paramName)})
			resErr = errors.New("InvalidParam")
			return
		}
		if ("+int" == paramType && intValue <= 0) || ("!-int" == paramType && intValue < 0) {
			c.JSON(400, gin.H{"status": "InvalidParam", "message": fmt.Sprintf("%s不正确", paramName)})
			resErr = errors.New("InvalidParam")
			return
		}
		resValue = intValue
		return

	} else if "string" == paramType { // 字符串, 去首尾空格
		if "string" == valueType {
			paramValue = strings.TrimSpace(paramValue.(string))
			if "" == paramValue && !allowEmpty {
				c.JSON(400, gin.H{"status": "emptyParam", "message": fmt.Sprintf("%s不得为空", paramName)})
				resErr = errors.New("emptyParam")
				return
			}
			resValue = paramValue
			return
		} else if "float64" == valueType {
			stringValue := fmt.Sprintf("%v", paramValue)
			resValue = stringValue
			return
		} else {
			c.JSON(400, gin.H{"status": "InvalidParam", "message": fmt.Sprintf("%s不正确", paramName)})
			resErr = errors.New("InvalidParam")
			return
		}

	} else if EnumMark := paramType[0:1]; "[" == EnumMark { // 枚举
		var enum []interface{}
		if err := json.Unmarshal([]byte(paramType), &enum); err != nil {
			panic(err)
		}
		for _, value := range enum {
			enumType := reflect.TypeOf(enum[0]).String()
			if enumType == valueType && paramValue == value {
				resValue = value
				return
			}
			if "float64" == valueType {
				stringValue := fmt.Sprintf("%v", paramValue)
				if stringValue == value {
					resValue = value
					return
				}
			} else if "string" == valueType {
				floatValue, err := strconv.ParseFloat(paramValue.(string), 64)
				if err != nil {
					panic(err)
				}
				resValue = floatValue
				return
			} else {
				c.JSON(400, gin.H{"status": "InvalidParam", "message": fmt.Sprintf("%s不正确", paramName)})
				resErr = errors.New("InvalidParam")
				return
			}
		}
		c.JSON(400, gin.H{"status": "InvalidParam", "message": fmt.Sprintf("%s不正确", paramName)})
		resErr = errors.New("InvalidParam")
		return

	} else if "array" == paramType { // 数组
		if "[]interface {}" == valueType {
			resValue = paramValue
			return
		} else {
			c.JSON(400, gin.H{"status": "InvalidParam", "message": fmt.Sprintf("%s不正确", paramName)})
			resErr = errors.New("InvalidParam")
			return
		}
	}

	c.JSON(400, gin.H{"status": "UndefinedParamType", "message": fmt.Sprintf("未知数据类型: %s", paramName)})
	resErr = errors.New("UndefinedParamType")
	return
}
