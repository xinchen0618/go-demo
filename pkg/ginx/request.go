// Package ginx gin增强方法
//
//	此包中出现error会向客户端输出4xx/500错误, 调用时捕获到error直接结束业务逻辑即可
package ginx

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"go-demo/pkg/dbx"
	"go-demo/pkg/gox"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/gohouse/gorose/v2"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

// PageQuery 分页参数
type PageQuery struct {
	Db         gorose.IOrm
	Select     string
	From       string
	Where      string
	BindParams []any
	GroupBy    string
	Having     string
	OrderBy    string
}

// PageItems 分页结果
type PageItems struct {
	Page        int64            `json:"page"`
	PerPage     int64            `json:"per_page"`
	TotalPages  int64            `json:"total_pages"`
	TotalCounts int64            `json:"total_counts"`
	Items       []map[string]any `json:"items"`
}

// GetJsonBody 获取Json参数
//
//	patterns 模式格式 ["paramKey:paramName:paramType:paramPattern"]
//	  paramType: 类型. 详情见FilterParam()方法paramType参数.
//	  paramPattern: 传值模式. + 表示字段必传,值不可为空; * 表示字段选传,值可为空; ? 表示字段选传,值不可为空.
func GetJsonBody(c *gin.Context, patterns []string) (map[string]any, error) {
	jsonBody := make(map[string]any)
	_ = c.ShouldBindJSON(&jsonBody) // 这里的error不要处理, 因为空body会报error

	result := make(map[string]any)
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
				Error(c, 400, "ParamEmpty", fmt.Sprintf("%s不得为空", patternAtoms[1]))
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
//
//	patterns 模式格式 ["paramKey:paramName:paramType:defaultValue"]
//	  paramType: 类型. 详情见FilterParam()方法paramType参数.
//	  defaultValue: 默认值. required 表示参数必填, "" 表示空字符串; 字符串不需要引号.
func GetQueries(c *gin.Context, patterns []string) (map[string]any, error) {
	result := make(map[string]any)
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
				Error(c, 400, "ParamEmpty", fmt.Sprintf("%s不得为空", patternAtoms[1]))
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
//
//	paramType 参数类型:
//		integer 整型64位,
//		+integer 正整型64位,
//		!-integer 非负整型64位,
//		string 字符串,
//		float.%d 浮点数,
//		decimal.%d 精度小数,
//		[] 枚举(支持数字float64与字符串string混合枚举),
//		array 数组,
//		[]integer 整型64位数组,
//		[]string 字符串数组.
func FilterParam(c *gin.Context, paramName string, paramValue any, paramType string, allowEmpty bool) (any, error) {
	valueType := reflect.TypeOf(paramValue).String() // 用户输入值类型

	// 整型64位
	if "integer" == paramType {
		valueStr, err := FilterParam(c, paramName, paramValue, "string", allowEmpty) // 先统一转字符串再转整型, 这样小数就不允许输入了
		if err != nil {
			return nil, err
		}
		if "" == valueStr.(string) {
			return int64(0), nil
		}
		valueInt, err := strconv.ParseInt(cast.ToString(valueStr), 10, 64) // 解决前导0被识别为8进制的问题
		if err != nil {
			Error(c, 400, "ParamInvalid", fmt.Sprintf("%s不正确", paramName))
			return nil, errors.New("ParamInvalid")
		}
		return valueInt, nil
	}

	// 正整数64位
	if "+integer" == paramType {
		valueInt, err := FilterParam(c, paramName, paramValue, "integer", allowEmpty)
		if err != nil {
			return nil, err
		}
		if valueInt.(int64) <= 0 {
			Error(c, 400, "ParamInvalid", fmt.Sprintf("%s不正确", paramName))
			return nil, errors.New("ParamInvalid")
		}
		return valueInt, nil
	}

	// 非负整数64位
	if "!-integer" == paramType {
		valueInt, err := FilterParam(c, paramName, paramValue, "integer", allowEmpty)
		if err != nil {
			return nil, err
		}
		if valueInt.(int64) < 0 {
			Error(c, 400, "ParamInvalid", fmt.Sprintf("%s不正确", paramName))
			return nil, errors.New("ParamInvalid")
		}
		return valueInt, nil
	}

	// 字符串, 去首尾空格
	if "string" == paramType {
		valueStr, err := cast.ToStringE(paramValue)
		if err != nil {
			Error(c, 400, "ParamInvalid", fmt.Sprintf("%s不正确", paramName))
			return nil, errors.New("ParamInvalid")
		}
		valueStr = strings.TrimSpace(valueStr)
		if "" == valueStr && !allowEmpty {
			Error(c, 400, "ParamEmpty", fmt.Sprintf("%s不得为空", paramName))
			return nil, errors.New("ParamEmpty")
		}

		return valueStr, nil
	}

	// 浮点数, float.%d, 数字表示精度(没有后补零), 超过精度四舍五入, 点号同数字可省略, 表示无限制, 返回类型为float64
	if "float" == lo.Substring(paramType, 0, 5) {
		valueStr, err := FilterParam(c, paramName, paramValue, "string", allowEmpty)
		if err != nil {
			return nil, err
		}

		if "" == valueStr.(string) && allowEmpty {
			return 0.0, nil
		}

		valueFloat, err := cast.ToFloat64E(valueStr)
		if err != nil {
			Error(c, 400, "ParamInvalid", fmt.Sprintf("%s不正确", paramName))
			return nil, errors.New("ParamInvalid")
		}

		prec := -1
		precStr := lo.Substring(paramType, 6, math.MaxInt)
		if precStr != "" {
			prec, err = cast.ToIntE(precStr)
			if err != nil {
				Error(c, 400, "ParamTypeError", fmt.Sprintf("数据类型错误: %s", paramName))
				return nil, errors.New("ParamTypeError")
			}
		}
		if -1 == prec {
			return valueFloat, nil
		}

		return gox.Round(valueFloat, prec), nil
	}

	// 精度小数, decimal.%d, 数字表示精度(有后补零), 超过精度四舍五入, 点号同数字可省略, 默认为2位小数, 返回类型为字符串
	if "decimal" == lo.Substring(paramType, 0, 7) {
		prec := 2
		precStr := lo.Substring(paramType, 8, math.MaxInt)
		if precStr != "" {
			var err error
			prec, err = cast.ToIntE(precStr)
			if err != nil {
				Error(c, 400, "ParamTypeError", fmt.Sprintf("数据类型错误: %s", paramName))
				return nil, errors.New("ParamTypeError")
			}
		}
		valueFloat, err := FilterParam(c, paramName, paramValue, fmt.Sprintf("float.%d", prec), allowEmpty)
		if err != nil {
			return nil, err
		}

		return strconv.FormatFloat(valueFloat.(float64), 'f', prec, 64), nil // 这里不会有精度问题, 精度在float递归时已经处理了
	}

	// 枚举, 支持数字float64与字符串string混合枚举
	if "[" == paramType[0:1] && "]" != paramType[1:2] {
		var enum []any
		if err := json.Unmarshal([]byte(paramType), &enum); err != nil { // 候选值解析到切片
			Error(c, 400, "ParamInvalid", fmt.Sprintf("%s不正确", paramName))
			return nil, errors.New("ParamInvalid")
		}
		for _, value := range enum { // 用户输入与候选值逐个比较
			enumType := reflect.TypeOf(value).String() // 候选值类型
			// 用户输入类型与候选类型一致
			if enumType == valueType && paramValue == value {
				return value, nil
			}
			// 用户输入类型与候选类型不一致
			if "float64" == valueType {
				valueStr := cast.ToString(paramValue)
				if valueStr == value {
					return value, nil
				}
			} else if "string" == valueType {
				valueFloat, err := cast.ToFloat64E(paramValue)
				if err != nil {
					Error(c, 400, "ParamInvalid", fmt.Sprintf("%s不正确", paramName))
					return nil, errors.New("ParamInvalid")
				}
				if valueFloat == value {
					return valueFloat, nil
				}
			} else {
				Error(c, 400, "ParamInvalid", fmt.Sprintf("%s不正确", paramName))
				return nil, errors.New("ParamInvalid")
			}
		}
		Error(c, 400, "ParamInvalid", fmt.Sprintf("%s不正确", paramName))
		return nil, errors.New("ParamInvalid")
	}

	// 数组
	if "array" == paramType {
		if "[]interface {}" == valueType {
			if !allowEmpty && 0 == len(paramValue.([]any)) {
				Error(c, 400, "ParamEmpty", fmt.Sprintf("%s不得为空", paramName))
				return nil, errors.New("ParamEmpty")
			}
			return paramValue, nil
		}
		Error(c, 400, "ParamInvalid", fmt.Sprintf("%s不正确", paramName))
		return nil, errors.New("ParamInvalid")
	}

	// int64数组
	if "[]integer" == paramType {
		valueArr, err := FilterParam(c, paramName, paramValue, "array", allowEmpty)
		if err != nil {
			return nil, err
		}
		intSlice := make([]int64, 0)
		for _, item := range valueArr.([]any) {
			itemAny, err := FilterParam(c, paramName, item, "integer", false)
			if err != nil {
				return nil, err
			}
			intSlice = append(intSlice, itemAny.(int64))
		}
		return intSlice, nil
	}

	// string数组
	if "[]string" == paramType {
		arrayValue, err := FilterParam(c, paramName, paramValue, "array", allowEmpty)
		if err != nil {
			return nil, err
		}
		stringSlice := make([]string, 0)
		for _, item := range arrayValue.([]any) {
			itemAny, err := FilterParam(c, paramName, item, "string", false)
			if err != nil {
				return nil, err
			}
			stringSlice = append(stringSlice, itemAny.(string))
		}
		return stringSlice, nil
	}

	Error(c, 400, "ParamTypeUndefined", fmt.Sprintf("未知数据类型: %s", paramName))
	return nil, errors.New("ParamTypeUndefined")
}

// GetPageItems 获取分页数据
func GetPageItems(c *gin.Context, pageQuery PageQuery) (PageItems, error) {
	queries, err := GetQueries(c, []string{"page:页码:+integer:1", "per_page:页大小:+integer:12"})
	if err != nil {
		return PageItems{}, err
	}
	page := queries["page"].(int64)
	perPage := queries["per_page"].(int64)

	bindParams := make([]any, 0)
	if pageQuery.BindParams != nil {
		bindParams = pageQuery.BindParams
	}

	where := pageQuery.Where
	if "" == where {
		where = "1"
	}

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
	countsData, err := dbx.FetchValue(pageQuery.Db, countSql, bindParams...) // 计算总记录数
	if err != nil {
		InternalError(c, nil)
		return PageItems{}, errors.New("InternalError")
	}
	counts := countsData.(int64)
	if 0 == counts { // 没有数据
		result := PageItems{
			Page:        page,
			PerPage:     perPage,
			TotalPages:  0,
			TotalCounts: 0,
			Items:       []map[string]any{},
		}
		return result, nil
	}

	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", pageQuery.Select, pageQuery.From, where)
	if pageQuery.OrderBy != "" {
		sql += fmt.Sprintf(" ORDER BY %s", pageQuery.OrderBy)
	}
	offset := (page - 1) * perPage
	sql += fmt.Sprintf(" LIMIT %d, %d", offset, perPage)
	items, err := dbx.FetchAll(pageQuery.Db, sql, bindParams...)
	if err != nil {
		InternalError(c, nil)
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
