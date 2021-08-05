// Package ginx gin增强方法
package ginx

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-demo/config/di"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gohouse/gorose/v2"
	"github.com/shopspring/decimal"
)

// GetJsonBody 获取Json参数
//	参数异常时方法会向客户端返回4xx错误, 调用方法时捕获到error直接结束业务逻辑即可
// 	@param c *gin.Context
// 	@param patterns []string ["paramKey:paramName:paramType:paramPattern"] paramPattern +必填不可为空, *选填可为空, ?选填不可为空
//	@return map[string]interface{}
//	@return error
func GetJsonBody(c *gin.Context, patterns []string) (map[string]interface{}, error) {
	jsonBody := make(map[string]interface{})
	_ = c.ShouldBindJSON(&jsonBody) // 这里的error不要处理, 因为空body会报error

	result := make(map[string]interface{})
	var err error
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
				c.JSON(400, gin.H{"status": "ParamEmpty", "message": fmt.Sprintf("%s不得为空", patternAtoms[1])})
				return nil, errors.New("ParamEmpty")
			} else {
				continue
			}
		}

		result[patternAtoms[0]], err = FilterParam(c, patternAtoms[1], paramValue, patternAtoms[2], allowEmpty)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// GetQueries 获取Query参数
//	参数异常时方法会向客户端返回4xx错误, 调用方法时捕获到error直接结束业务逻辑即可
// 	@param c *gin.Context
// 	@param patterns []string ["paramKey:paramName:paramType:defaultValue"] defaultValue为required时参数必填
//	@return map[string]interface{}
//	@return error
func GetQueries(c *gin.Context, patterns []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var err error
	for _, pattern := range patterns {
		patternAtoms := strings.Split(pattern, ":")
		allowEmpty := false
		if `""` == patternAtoms[3] { // 默认值""表示空字符串
			patternAtoms[3] = ""
			allowEmpty = true
		}
		paramValue := c.Query(patternAtoms[0])
		if "" == paramValue {
			if "required" == patternAtoms[3] { // 必填
				c.JSON(400, gin.H{"status": "ParamEmpty", "message": fmt.Sprintf("%s不得为空", patternAtoms[1])})
				return nil, errors.New("ParamEmpty")
			} else {
				paramValue = patternAtoms[3]
			}
		}

		result[patternAtoms[0]], err = FilterParam(c, patternAtoms[1], paramValue, patternAtoms[2], allowEmpty)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// FilterParam 校验参数类型
//	参数异常时方法会向客户端返回4xx错误, 调用方法时捕获到error直接结束业务逻辑即可
// 	@param c *gin.Context
// 	@param paramName string
// 	@param paramValue interface{}
// 	@param paramType string int整型64位, +int正整型64位, !-int非负整型64位, string字符串, []枚举(支持数字float64与字符串string混合枚举), array数组
// 	@param allowEmpty bool
//	@return interface{}
//	@return error
func FilterParam(c *gin.Context, paramName string, paramValue interface{}, paramType string, allowEmpty bool) (interface{}, error) {
	valueType := reflect.TypeOf(paramValue).String()

	/* 整型 */
	if "int" == paramType {
		stringValue, err := FilterParam(c, paramName, paramValue, "string", allowEmpty) // 先统一转字符串再转整型, 这样小数就不允许输入了
		if err != nil {
			return nil, err
		}
		if "" == stringValue.(string) {
			return 0, nil
		}
		intValue, err := strconv.ParseInt(stringValue.(string), 10, 64) // 转整型64位
		if err != nil {
			c.JSON(400, gin.H{"status": "ParamInvalid", "message": fmt.Sprintf("%s不正确", paramName)})
			return nil, errors.New("ParamInvalid")
		}
		return intValue, nil
	}

	/* 正整数64位 */
	if "+int" == paramType {
		intValue, err := FilterParam(c, paramName, paramValue, "int", allowEmpty)
		if err != nil {
			return nil, err
		}
		if intValue.(int64) <= 0 {
			c.JSON(400, gin.H{"status": "ParamInvalid", "message": fmt.Sprintf("%s不正确", paramName)})
			return nil, errors.New("ParamInvalid")
		}
		return intValue, nil
	}

	/* 非负整数64位 */
	if "!-int" == paramType {
		intValue, err := FilterParam(c, paramName, paramValue, "int", allowEmpty)
		if err != nil {
			return nil, err
		}
		if intValue.(int64) < 0 {
			c.JSON(400, gin.H{"status": "ParamInvalid", "message": fmt.Sprintf("%s不正确", paramName)})
			return nil, errors.New("ParamInvalid")
		}
		return intValue, nil
	}

	/* 字符串, 去首尾空格*/
	if "string" == paramType {
		if "string" == valueType {
			stringValue := strings.TrimSpace(paramValue.(string))
			if "" == stringValue && !allowEmpty {
				c.JSON(400, gin.H{"status": "ParamEmpty", "message": fmt.Sprintf("%s不得为空", paramName)})
				return nil, errors.New("ParamEmpty")
			}
			return stringValue, nil
		} else if "float64" == valueType {
			decimalValue, err := decimal.NewFromString(fmt.Sprintf("%v", paramValue)) // 解决6位以上数据被转科学记数法的问题
			if err != nil {
				c.JSON(400, gin.H{"status": "ParamInvalid", "message": fmt.Sprintf("%s不正确", paramName)})
				return nil, errors.New("ParamInvalid")
			}
			return decimalValue.String(), nil
		} else {
			c.JSON(400, gin.H{"status": "ParamInvalid", "message": fmt.Sprintf("%s不正确", paramName)})
			return nil, errors.New("ParamInvalid")
		}
	}

	/* 枚举, 支持数字float64与字符串string混合枚举 */
	if EnumMark := paramType[0:1]; "[" == EnumMark {
		var enum []interface{}
		if err := json.Unmarshal([]byte(paramType), &enum); err != nil {
			panic(err)
		}
		for _, value := range enum {
			enumType := reflect.TypeOf(enum[0]).String()
			if enumType == valueType && paramValue == value {
				return value, nil
			}
			if "float64" == valueType {
				stringValue := fmt.Sprintf("%v", paramValue)
				if stringValue == value {
					return value, nil
				}
			} else if "string" == valueType {
				floatValue, err := strconv.ParseFloat(paramValue.(string), 64)
				if err != nil {
					panic(err)
				}
				return floatValue, nil
			} else {
				c.JSON(400, gin.H{"status": "ParamInvalid", "message": fmt.Sprintf("%s不正确", paramName)})
				return nil, errors.New("ParamInvalid")
			}
		}
		c.JSON(400, gin.H{"status": "ParamInvalid", "message": fmt.Sprintf("%s不正确", paramName)})
		return nil, errors.New("ParamInvalid")
	}

	/* 数组 */
	if "array" == paramType {
		if "[]interface {}" == valueType {
			return paramValue, nil
		} else {
			c.JSON(400, gin.H{"status": "ParamInvalid", "message": fmt.Sprintf("%s不正确", paramName)})
			return nil, errors.New("ParamInvalid")
		}
	}

	c.JSON(400, gin.H{"status": "ParamTypeUndefined", "message": fmt.Sprintf("未知数据类型: %s", paramName)})
	return nil, errors.New("ParamTypeUndefined")
}

// GetPageItems 获取分页数据
// 	出现异常时方法会向客户端返回4xx错误, 调用方法捕获到error直接结束业务逻辑即可
// 	@param query map[string]interface{} {"ginContext": *gin.Context, "db": gorose.IOrm, "select": string, "from": string, "where": string, "groupBy" => string, "having" => string, "orderBy": string}
//	@return map[string]interface{} {"page": int64, "per_page": int64, "total_page": int64, "total_counts": int64, "items": []gorose.Data}
//	@return error
func GetPageItems(query map[string]interface{}) (map[string]interface{}, error) {
	queries, err := GetQueries(query["ginContext"].(*gin.Context), []string{"page:页码:+int:1", "per_page:页大小:+int:12"})
	if err != nil {
		return nil, err
	}

	bindParams, ok := query["bindParams"].([]interface{}) // 参数绑定
	if !ok {
		bindParams = []interface{}{}
	}

	where := query["where"].(string)

	var countSql string
	groupBy, ok := query["groupBy"].(string) // GROUP BY存在总记录数计算方式会不同
	if ok {
		where += " GROUP BY " + groupBy
		having, ok := query["having"].(string)
		if ok {
			where += " HAVING " + having
		}
		countSql = fmt.Sprintf("SELECT COUNT(*) AS counts FROM (SELECT %s FROM %s WHERE %s) AS t", query["select"], query["from"], where)
	} else {
		countSql = fmt.Sprintf("SELECT COUNT(*) AS counts FROM %s WHERE %s", query["from"], where)
	}
	counts, err := query["db"].(gorose.IOrm).Query(countSql, bindParams...) // 计算总记录数
	if err != nil {
		panic(err)
	}
	if 0 == counts[0]["counts"].(int64) { // 没有数据
		result := map[string]interface{}{
			"page":         queries["page"],
			"per_page":     queries["per_page"],
			"total_pages":  0,
			"total_counts": 0,
			"items":        []gorose.Data{},
		}
		return result, nil
	}

	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", query["select"], query["from"], where)
	orderBy, ok := query["orderBy"]
	if ok {
		sql += fmt.Sprintf(" ORDER BY %s", orderBy)
	}
	offset := (queries["page"].(int64) - 1) * queries["per_page"].(int64)
	sql += fmt.Sprintf(" LIMIT %d, %d", offset, queries["per_page"])
	items, err := query["db"].(gorose.IOrm).Query(sql, bindParams...)
	if err != nil {
		panic(err)
	}
	result := map[string]interface{}{
		"page":         queries["page"],
		"per_page":     queries["per_page"],
		"total_pages":  math.Ceil(float64(counts[0]["counts"].(int64)) / float64(queries["per_page"].(int64))),
		"total_counts": counts[0]["counts"],
		"items":        items,
	}
	return result, nil
}

// InternalError 服务异常
//	记录日志并向客户端返回500错误
//	@param c *gin.Context
//	@param err error
func InternalError(c *gin.Context, err error) {
	di.Logger().Error(err.Error())
	c.JSON(500, gin.H{"status": "InternalError", "message": "服务异常, 请稍后重试"})
}
