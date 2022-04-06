// Package ginx gin增强方法
//	此包中出现error会向客户端输出4xx/500错误, 调用时捕获到error直接结束业务逻辑即可
package ginx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go-demo/config/di"
	"go-demo/pkg/dbx"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gohouse/gorose/v2"
	"github.com/spf13/cast"
	"golang.org/x/sync/singleflight"
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

var (
	cacheSg singleflight.Group
)

// GetJsonBody 获取Json参数
// 	@param c *gin.Context
// 	@param patterns []string ["paramKey:paramName:paramType:paramPattern"] paramPattern +必填不可为空, *选填可为空, ?选填不可为空
//	@return map[string]any
//	@return error
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
// 	@param c *gin.Context
// 	@param patterns []string ["paramKey:paramName:paramType:defaultValue"] defaultValue为required时参数必填
//	@return map[string]any
//	@return error
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
// 	@param c *gin.Context
// 	@param paramName string
// 	@param paramValue any
// 	@param paramType string int整型64位, +int正整型64位, !-int非负整型64位, string字符串, money金额, []枚举(支持数字float64与字符串string混合枚举), array数组, []int整型64位数组, []string字符串数组
// 	@param allowEmpty bool
//	@return any
//	@return error
func FilterParam(c *gin.Context, paramName string, paramValue any, paramType string, allowEmpty bool) (any, error) {
	valueType := reflect.TypeOf(paramValue).String() // 用户输入值类型

	// 整型64位
	if "int" == paramType {
		valueStr, err := FilterParam(c, paramName, paramValue, "string", allowEmpty) // 先统一转字符串再转整型, 这样小数就不允许输入了
		if err != nil {
			return nil, err
		}
		if "" == valueStr.(string) {
			return int64(0), nil
		}
		valueInt, err := cast.ToInt64E(valueStr)
		if err != nil {
			Error(c, 400, "ParamInvalid", fmt.Sprintf("%s不正确", paramName))
			return nil, errors.New("ParamInvalid")
		}
		return valueInt, nil
	}

	// 正整数64位
	if "+int" == paramType {
		valueInt, err := FilterParam(c, paramName, paramValue, "int", allowEmpty)
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
	if "!-int" == paramType {
		valueInt, err := FilterParam(c, paramName, paramValue, "int", allowEmpty)
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

	// 金额
	if "money" == paramType {
		valueStr, err := FilterParam(c, paramName, paramValue, "string", allowEmpty)
		if err != nil {
			return nil, err
		}
		if "" == valueStr.(string) {
			return "0.00", nil
		}
		valueFloat, err := cast.ToFloat64E(valueStr)
		if err != nil {
			Error(c, 400, "ParamInvalid", fmt.Sprintf("%s不正确", paramName))
			return nil, errors.New("ParamInvalid")
		}

		valueMoney := strconv.FormatFloat(valueFloat, 'f', 2, 64)
		if valueMoney != valueStr {
			Error(c, 400, "ParamInvalid", fmt.Sprintf("%s不正确", paramName))
			return nil, errors.New("ParamInvalid")
		}

		return valueMoney, nil
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
	if "[]int" == paramType {
		valueArr, err := FilterParam(c, paramName, paramValue, "array", allowEmpty)
		if err != nil {
			return nil, err
		}
		intSlice := []int64{}
		for _, item := range valueArr.([]any) {
			itemAny, err := FilterParam(c, paramName, item, "int", false)
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
		stringSlice := []string{}
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
//  @param c *gin.Context
//  @param pageQuery PageQuery
//  @return PageItems
//  @return error
func GetPageItems(c *gin.Context, pageQuery PageQuery) (PageItems, error) {
	queries, err := GetQueries(c, []string{"page:页码:+int:1", "per_page:页大小:+int:12"})
	if err != nil {
		return PageItems{}, err
	}
	page := queries["page"].(int64)
	perPage := queries["per_page"].(int64)

	bindParams := []any{}
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
		InternalError(c)
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
		InternalError(c)
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

// GetOrSetCache 获取或者设置业务缓存
//	方法返回的是json.Unmarshal的数据
//	@receiver cacheService
//	@param key string
//	@param ttl time.Duration 缓存时长
//	@param f func() (any, error)
//	@return any
//	@return error
func GetOrSetCache(c *gin.Context, key string, ttl time.Duration, f func() (any, error)) (any, error) {
	result, err, _ := cacheSg.Do(key, func() (any, error) {
		var resultCache string
		resultCache, err := di.CacheRedis().Get(context.Background(), key).Result()
		if err != nil {
			if err != redis.Nil {
				InternalError(c, err)
				return nil, err
			}

			// 缓存不存在
			result, err := f()
			if err != nil {
				return nil, err
			}
			resultBytes, err := json.Marshal(result)
			if err != nil {
				InternalError(c, err)
				return nil, err
			}
			if err := di.CacheRedis().Set(context.Background(), key, resultBytes, ttl).Err(); err != nil {
				InternalError(c, err)
				return nil, err
			}
			resultCache = string(resultBytes)
		}

		var resultAny any
		if err := json.Unmarshal([]byte(resultCache), &resultAny); err != nil {
			InternalError(c, err)
			return nil, err
		}
		return resultAny, nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
