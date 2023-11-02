// Package ginx Gin 增强函数
//
//	此包中出现 error 会向客户端输出4xx/500错误, 调用时捕获到 error 直接结束业务逻辑即可.
package ginx

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"go-demo/pkg/gox"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

// GetJSONBody 获取 JSON 参数
//
//	patterns 模式格式 ["paramKey:paramName:paramType:paramPattern"]
//	  paramType: 类型. 详情见 FilterParam() 方法 paramType 参数.
//	  paramPattern: 传值模式. + 表示字段必传,值不可为空; * 表示字段选传,值可为空; ? 表示字段选传,值不可为空.
func GetJSONBody(c *gin.Context, patterns []string) (map[string]any, error) {
	// body
	jsonBody := make(map[string]any)
	_ = c.ShouldBindJSON(&jsonBody) // 这里的 error 不要处理, 因为空 body 会报 error
	// 逐字段校验
	result := make(map[string]any)
	var err error
	for _, pattern := range patterns {
		// pattern
		patternAtoms := strings.Split(pattern, ":")
		required := true
		allowEmpty := false
		if patternAtoms[3] == "+" {
			required = true
			allowEmpty = false
		} else if patternAtoms[3] == "*" {
			required = false
			allowEmpty = true
		} else if patternAtoms[3] == "?" {
			required = false
			allowEmpty = false
		}
		// key
		paramValue, ok := jsonBody[patternAtoms[0]]
		if !ok || paramValue == nil {
			if required {
				Error(c, 400, "ParamEmpty", patternAtoms[1]+"不得为空")
				return nil, errors.New("ParamEmpty")
			} else {
				continue
			}
		}
		// 类型值
		result[patternAtoms[0]], err = FilterParam(c, patternAtoms[1], paramValue, patternAtoms[2], allowEmpty)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// GetQueries 获取 Query 参数
//
//	patterns 模式格式 ["paramKey:paramName:paramType:defaultValue"]
//	  paramType: 类型. 详情见 FilterParam() 方法 paramType 参数.
//	  defaultValue: 默认值. required 表示参数必填, "" 表示空字符串; 字符串不需要引号.
func GetQueries(c *gin.Context, patterns []string) (map[string]any, error) {
	// 逐字段校验
	result := make(map[string]any)
	var err error
	for _, pattern := range patterns {
		patternAtoms := strings.Split(pattern, ":")
		// default
		allowEmpty := false
		if patternAtoms[3] == `""` { // 默认值""表示空字符串
			patternAtoms[3] = ""
			allowEmpty = true
		}
		// key
		paramValue := c.Query(patternAtoms[0])
		if paramValue == "" {
			if patternAtoms[3] == "required" { // 必填
				Error(c, 400, "ParamEmpty", patternAtoms[1]+"不得为空")
				return nil, errors.New("ParamEmpty")
			} else {
				paramValue = patternAtoms[3]
			}
		}
		// 类型值
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
//		integer 整型64位;
//		+integer 正整型64位;
//		!-integer 非负整型64位;
//		string 字符串, 去首尾空格;
//		float.%d 浮点数, 数字表示精度(没有后补零), 超过精度四舍五入, 点号同数字可省略, 表示无限制, 返回类型为 float64;
//		decimal.%d 精度小数, 数字表示精度(有后补零), 超过精度四舍五入, 点号同数字可省略, 默认为2位小数, 返回类型为字符串;
//		[] 枚举, 支持数字 float64 与字符串 string 混合枚举;
//		array 数组;
//		[]integer 整型64位数组;
//		[]string 字符串数组;
func FilterParam(c *gin.Context, paramName string, paramValue any, paramType string, allowEmpty bool) (any, error) {
	valueType := reflect.TypeOf(paramValue).String() // 用户输入值类型

	// 整型64位
	if paramType == "integer" {
		valueStr, err := FilterParam(c, paramName, paramValue, "string", allowEmpty) // 先统一转字符串再转整型, 这样小数就不允许输入了
		if err != nil {
			return nil, err
		}
		if valueStr.(string) == "" {
			return int64(0), nil
		}
		valueInt, err := strconv.ParseInt(cast.ToString(valueStr), 10, 64) // 解决前导0被识别为8进制的问题
		if err != nil {
			Error(c, 400, "ParamInvalid", paramName+"不正确")
			return nil, errors.New("ParamInvalid")
		}
		return valueInt, nil
	}

	// 正整型64位
	if paramType == "+integer" {
		valueInt, err := FilterParam(c, paramName, paramValue, "integer", allowEmpty)
		if err != nil {
			return nil, err
		}
		if valueInt.(int64) <= 0 {
			Error(c, 400, "ParamInvalid", paramName+"不正确")
			return nil, errors.New("ParamInvalid")
		}
		return valueInt, nil
	}

	// 非负整型64位
	if paramType == "!-integer" {
		valueInt, err := FilterParam(c, paramName, paramValue, "integer", allowEmpty)
		if err != nil {
			return nil, err
		}
		if valueInt.(int64) < 0 {
			Error(c, 400, "ParamInvalid", paramName+"不正确")
			return nil, errors.New("ParamInvalid")
		}
		return valueInt, nil
	}

	// 字符串, 去首尾空格
	if paramType == "string" {
		valueStr, err := cast.ToStringE(paramValue)
		if err != nil {
			Error(c, 400, "ParamInvalid", paramName+"不正确")
			return nil, errors.New("ParamInvalid")
		}
		valueStr = strings.TrimSpace(valueStr)
		if valueStr == "" && !allowEmpty {
			Error(c, 400, "ParamEmpty", paramName+"不得为空")
			return nil, errors.New("ParamEmpty")
		}

		return valueStr, nil
	}

	// 浮点数, float.%d, 数字表示精度(没有后补零), 超过精度四舍五入, 点号同数字可省略, 表示无限制, 返回类型为 float64
	if lo.Substring(paramType, 0, 5) == "float" {
		// 值
		valueStr, err := FilterParam(c, paramName, paramValue, "string", allowEmpty)
		if err != nil {
			return nil, err
		}
		// 空值
		if valueStr.(string) == "" && allowEmpty {
			return 0.0, nil
		}
		// float
		valueFloat, err := cast.ToFloat64E(valueStr)
		if err != nil {
			Error(c, 400, "ParamInvalid", paramName+"不正确")
			return nil, errors.New("ParamInvalid")
		}
		// 精度
		prec := -1
		precStr := lo.Substring(paramType, 6, math.MaxUint)
		if precStr != "" {
			prec, err = cast.ToIntE(precStr)
			if err != nil {
				Error(c, 400, "ParamTypeError", "数据类型错误: "+paramName)
				return nil, errors.New("ParamTypeError")
			}
		}
		if prec == -1 {
			return valueFloat, nil
		}

		return gox.Round(valueFloat, prec), nil
	}

	// 精度小数, decimal.%d, 数字表示精度(有后补零), 超过精度四舍五入, 点号同数字可省略, 默认为2位小数, 返回类型为字符串
	if lo.Substring(paramType, 0, 7) == "decimal" {
		prec := 2
		precStr := lo.Substring(paramType, 8, math.MaxUint)
		if precStr != "" {
			var err error
			prec, err = cast.ToIntE(precStr)
			if err != nil {
				Error(c, 400, "ParamTypeError", "数据类型错误: "+paramName)
				return nil, errors.New("ParamTypeError")
			}
		}
		valueFloat, err := FilterParam(c, paramName, paramValue, fmt.Sprintf("float.%d", prec), allowEmpty)
		if err != nil {
			return nil, err
		}

		return strconv.FormatFloat(valueFloat.(float64), 'f', prec, 64), nil // 这里不会有精度问题, 精度在float递归时已经处理了
	}

	// 枚举, 支持数字 float64 与字符串 string 混合枚举
	if paramType[0:1] == "[" && paramType[1:2] != "]" {
		var enum []any
		if err := json.Unmarshal([]byte(paramType), &enum); err != nil { // 候选值解析到切片
			Error(c, 400, "ParamInvalid", paramName+"不正确")
			return nil, errors.New("ParamInvalid")
		}
		for _, value := range enum { // 用户输入与候选值逐个比较
			enumType := reflect.TypeOf(value).String() // 候选值类型
			// 用户输入类型与候选类型一致
			if enumType == valueType && paramValue == value {
				return value, nil
			}
			// 用户输入类型与候选类型不一致
			if valueType == "float64" {
				valueStr := cast.ToString(paramValue)
				if valueStr == value {
					return value, nil
				}
			} else if valueType == "string" {
				valueFloat, err := cast.ToFloat64E(paramValue)
				if err != nil {
					Error(c, 400, "ParamInvalid", paramName+"不正确")
					return nil, errors.New("ParamInvalid")
				}
				if valueFloat == value {
					return valueFloat, nil
				}
			} else {
				Error(c, 400, "ParamInvalid", paramName+"不正确")
				return nil, errors.New("ParamInvalid")
			}
		}
		Error(c, 400, "ParamInvalid", paramName+"不正确")
		return nil, errors.New("ParamInvalid")
	}

	// 数组
	if paramType == "array" {
		if valueType == "[]interface {}" {
			if !allowEmpty && len(paramValue.([]any)) == 0 {
				Error(c, 400, "ParamEmpty", paramName+"不得为空")
				return nil, errors.New("ParamEmpty")
			}
			return paramValue, nil
		}
		Error(c, 400, "ParamInvalid", paramName+"不正确")
		return nil, errors.New("ParamInvalid")
	}

	// int64 数组
	if paramType == "[]integer" {
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

	// string 数组
	if paramType == "[]string" {
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

	Error(c, 400, "ParamTypeUndefined", "未知数据类型: %s"+paramName)
	return nil, errors.New("ParamTypeUndefined")
}

// PageQuery 分页参数
type PageQuery struct {
	DB         *gorm.DB
	Model      any // 表 model 指针
	Where      string
	BindParams []any
	OrderBy    string
}

// Paging 分页结果
type Paging struct {
	Page         int64 `json:"page"`          // 页码
	PerPage      int64 `json:"per_page"`      // 页大小
	TotalPages   int64 `json:"total_pages"`   // 总页数
	TotalResults int64 `json:"total_results"` // 总记录数
}

// Paginate 获取分页数据
func Paginate(c *gin.Context, items any, pageQuery PageQuery) (Paging, error) {
	// 页码
	queries, err := GetQueries(c, []string{"page:页码:+integer:1", "per_page:页大小:+integer:12"})
	if err != nil {
		return Paging{}, err
	}
	page := queries["page"].(int64)
	perPage := queries["per_page"].(int64)
	// 总记录数
	var totalResults int64 // 计算总记录数
	if err := pageQuery.DB.Model(pageQuery.Model).Where(pageQuery.Where, pageQuery.BindParams...).Count(&totalResults).Error; err != nil {
		InternalError(c, err)
		return Paging{}, errors.New("InternalError")
	}
	if totalResults == 0 { // 没有数据
		result := Paging{
			Page:         page,
			PerPage:      perPage,
			TotalPages:   0,
			TotalResults: 0,
		}
		return result, nil
	}
	// items
	tx := pageQuery.DB.Model(pageQuery.Model).Where(pageQuery.Where, pageQuery.BindParams...)
	if pageQuery.OrderBy != "" {
		tx.Order(pageQuery.OrderBy)
	}
	offset := (page - 1) * perPage
	if err := tx.Offset(int(offset)).Limit(int(perPage)).Find(items).Error; err != nil {
		InternalError(c, err)
		return Paging{}, errors.New("InternalError")
	}
	result := Paging{
		Page:         page,
		PerPage:      perPage,
		TotalPages:   int64(math.Ceil(float64(totalResults) / float64(perPage))),
		TotalResults: totalResults,
	}
	return result, nil
}
