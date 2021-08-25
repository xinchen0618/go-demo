// Package ginx gin增强方法
//	此包中出现error会向客户端返回4xx/500错误, 调用时捕获到error直接结束业务逻辑即可
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

// PageQuery 分页参数
type PageQuery struct {
	GinCtx     *gin.Context
	Db         gorose.IOrm
	Select     string
	From       string
	Where      string
	BindParams []interface{}
	GroupBy    string
	Having     string
	OrderBy    string
}

// PageItems 分页结果
type PageItems struct {
	Page        int64         `json:"page"`
	PerPage     int64         `json:"per_page"`
	TotalPages  int64         `json:"total_pages"`
	TotalCounts int64         `json:"total_counts"`
	Items       []gorose.Data `json:"items"`
}

// GetJsonBody 获取Json参数
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
				c.JSON(400, gin.H{"code": "ParamEmpty", "message": fmt.Sprintf("%s不得为空", patternAtoms[1])})
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
				c.JSON(400, gin.H{"code": "ParamEmpty", "message": fmt.Sprintf("%s不得为空", patternAtoms[1])})
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
// 	@param c *gin.Context
// 	@param paramName string
// 	@param paramValue interface{}
// 	@param paramType string int整型64位, +int正整型64位, !-int非负整型64位, string字符串, []枚举(支持数字float64与字符串string混合枚举), array数组
// 	@param allowEmpty bool
//	@return interface{}
//	@return error
func FilterParam(c *gin.Context, paramName string, paramValue interface{}, paramType string, allowEmpty bool) (interface{}, error) {
	valueType := reflect.TypeOf(paramValue).String()

	/* 整型64位 */
	if "int" == paramType {
		stringValue, err := FilterParam(c, paramName, paramValue, "string", allowEmpty) // 先统一转字符串再转整型, 这样小数就不允许输入了
		if err != nil {
			return nil, err
		}
		if "" == stringValue.(string) {
			return int64(0), nil
		}
		intValue, err := strconv.ParseInt(stringValue.(string), 10, 64) // 转整型64位
		if err != nil {
			c.JSON(400, gin.H{"code": "ParamInvalid", "message": fmt.Sprintf("%s不正确", paramName)})
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
			c.JSON(400, gin.H{"code": "ParamInvalid", "message": fmt.Sprintf("%s不正确", paramName)})
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
			c.JSON(400, gin.H{"code": "ParamInvalid", "message": fmt.Sprintf("%s不正确", paramName)})
			return nil, errors.New("ParamInvalid")
		}
		return intValue, nil
	}

	/* 字符串, 去首尾空格*/
	if "string" == paramType {
		if "string" == valueType {
			stringValue := strings.TrimSpace(paramValue.(string))
			if "" == stringValue && !allowEmpty {
				c.JSON(400, gin.H{"code": "ParamEmpty", "message": fmt.Sprintf("%s不得为空", paramName)})
				return nil, errors.New("ParamEmpty")
			}
			return stringValue, nil
		} else if "float64" == valueType {
			decimalValue, err := decimal.NewFromString(fmt.Sprintf("%v", paramValue)) // 解决6位以上数据被转科学记数法的问题
			if err != nil {
				c.JSON(400, gin.H{"code": "ParamInvalid", "message": fmt.Sprintf("%s不正确", paramName)})
				return nil, errors.New("ParamInvalid")
			}
			return decimalValue.String(), nil
		} else {
			c.JSON(400, gin.H{"code": "ParamInvalid", "message": fmt.Sprintf("%s不正确", paramName)})
			return nil, errors.New("ParamInvalid")
		}
	}

	/* 枚举, 支持数字float64与字符串string混合枚举 */
	if EnumMark := paramType[0:1]; "[" == EnumMark {
		var enum []interface{}
		if err := json.Unmarshal([]byte(paramType), &enum); err != nil {
			InternalError(c, err)
			return nil, errors.New("InternalError")
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
					InternalError(c, err)
					return nil, errors.New("InternalError")
				}
				return floatValue, nil
			} else {
				c.JSON(400, gin.H{"code": "ParamInvalid", "message": fmt.Sprintf("%s不正确", paramName)})
				return nil, errors.New("ParamInvalid")
			}
		}
		c.JSON(400, gin.H{"code": "ParamInvalid", "message": fmt.Sprintf("%s不正确", paramName)})
		return nil, errors.New("ParamInvalid")
	}

	/* 数组 */
	if "array" == paramType {
		if "[]interface {}" == valueType {
			return paramValue, nil
		} else {
			c.JSON(400, gin.H{"code": "ParamInvalid", "message": fmt.Sprintf("%s不正确", paramName)})
			return nil, errors.New("ParamInvalid")
		}
	}

	c.JSON(400, gin.H{"code": "ParamTypeUndefined", "message": fmt.Sprintf("未知数据类型: %s", paramName)})
	return nil, errors.New("ParamTypeUndefined")
}

// GetPageItems 获取分页数据
// 	@param pageQuery PageQuery
//	@return PageItems
//	@return error
func GetPageItems(pageQuery PageQuery) (PageItems, error) {
	queries, err := GetQueries(pageQuery.GinCtx, []string{"page:页码:+int:1", "per_page:页大小:+int:12"})
	if err != nil {
		return PageItems{}, err
	}
	page := queries["page"].(int64)
	perPage := queries["per_page"].(int64)

	bindParams := []interface{}{}
	if pageQuery.BindParams != nil {
		bindParams = pageQuery.BindParams
	}

	where := pageQuery.Where

	var countSql string
	if pageQuery.GroupBy != "" { // GROUP BY存在总记录数计算方式会不同
		where += " GROUP BY " + pageQuery.GroupBy
		if pageQuery.Having != "" {
			where += " HAVING " + pageQuery.Having
		}
		countSql = fmt.Sprintf("SELECT COUNT(*) AS counts FROM (SELECT %s FROM %s WHERE %s) AS t", pageQuery.Select, pageQuery.From, where)
	} else {
		countSql = fmt.Sprintf("SELECT COUNT(*) AS counts FROM %s WHERE %s", pageQuery.From, where)
	}
	countsData, err := pageQuery.Db.Query(countSql, bindParams...) // 计算总记录数
	if err != nil {
		InternalError(pageQuery.GinCtx, err)
		return PageItems{}, errors.New("InternalError")
	}
	counts := countsData[0]["counts"].(int64)
	if 0 == counts { // 没有数据
		result := PageItems{
			Page:        page,
			PerPage:     perPage,
			TotalPages:  0,
			TotalCounts: 0,
			Items:       []gorose.Data{},
		}
		return result, nil
	}

	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", pageQuery.Select, pageQuery.From, where)
	if pageQuery.OrderBy != "" {
		sql += fmt.Sprintf(" ORDER BY %s", pageQuery.OrderBy)
	}
	offset := (page - 1) * perPage
	sql += fmt.Sprintf(" LIMIT %d, %d", offset, perPage)
	items, err := pageQuery.Db.Query(sql, bindParams...)
	if err != nil {
		InternalError(pageQuery.GinCtx, err)
		return PageItems{}, errors.New("InternalError")
	}
	result := PageItems{
		Page:        page,
		PerPage:     perPage,
		TotalPages:  int64(math.Ceil(float64(counts) / float64(perPage))),
		TotalCounts: counts,
		Items:       items,
	}
	return result, nil
}

// InternalError 服务异常
//	记录日志并向客户端返回500错误
//	@param c *gin.Context
//	@param err error
func InternalError(c *gin.Context, err error) {
	di.Logger().Error(err.Error())
	c.JSON(500, gin.H{"code": "InternalError", "message": "服务异常, 请稍后重试"})
}
