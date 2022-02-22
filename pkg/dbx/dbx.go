// Package dbx MySQL增/删/改/查/事务操作封装
//	MySQL=>Golang数据类型映射:
//		bigint/int/smallint/tinyint => int64,
//		float/double => float64,
//		varchar/char/longtext/text/mediumtext/tinytext/decimal/datetime/timestamp/date/time => string,
package dbx

import (
	"fmt"
	"go-demo/pkg/gox"
	"strings"

	"github.com/gohouse/gorose/v2"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

// FetchAll 获取多行记录
//  @param db gorose.IOrm
//  @param sql string
//  @param params ...interface{} 不支持切片
//  @return []map[string]interface{}
//  @return error
func FetchAll(db gorose.IOrm, sql string, params ...interface{}) ([]map[string]interface{}, error) {
	rows, err := db.Query(sql, params...)
	if err != nil {
		zap.L().Error(err.Error())
		return []map[string]interface{}{}, err
	}

	result := []map[string]interface{}{}
	for _, v := range rows {
		result = append(result, v)
	}

	return result, nil
}

// FetchOne 获取一行记录
//	查询时会自动添加限制LIMIT 1
//  @param db gorose.IOrm
//  @param sql string
//  @param params ...interface{}
//  @return map[string]interface{}
//  @return error
func FetchOne(db gorose.IOrm, sql string, params ...interface{}) (map[string]interface{}, error) {
	sql = strings.TrimSpace(sql)
	if !strings.HasSuffix(sql, "LIMIT 1") && !strings.HasSuffix(sql, "limit 1") {
		sql += " LIMIT 1"
	}

	rows, err := FetchAll(db, sql, params...)
	if err != nil {
		return map[string]interface{}{}, err
	}

	if 0 == len(rows) {
		return map[string]interface{}{}, nil
	}

	return rows[0], nil
}

// FetchValue 获取一个值
//  @param db gorose.IOrm
//  @param sql string
//  @param params ...interface{}
//  @return interface{}
//  @return error
func FetchValue(db gorose.IOrm, sql string, params ...interface{}) (interface{}, error) {
	row, err := FetchOne(db, sql, params...)
	if err != nil {
		return map[string]interface{}{}, err
	}

	for _, value := range row {
		return value, nil
	}

	// 0 == len(row)
	return nil, nil
}

// FetchColumn 获取一列值
//  @param db gorose.IOrm
//  @param sql string
//  @param params ...interface{}
//  @return []interface{}
//  @return error
func FetchColumn(db gorose.IOrm, sql string, params ...interface{}) ([]interface{}, error) {
	rows, err := FetchAll(db, sql, params...)
	if err != nil {
		return []interface{}{}, err
	}

	values := []interface{}{}
	for _, row := range rows {
		for _, value := range row {
			values = append(values, value)
			break
		}
	}

	return values, nil
}

// Slice2in Slice转IN条件
// 	Golang SQL驱动不支持IN(?)
//	使用fmt.Sprint("IN(%s)", Slice2in(s))
//	MySQL整型字段查询添加引号无影响
//  @param s interface{}
//  @return string
func Slice2in(s interface{}) string {
	stringSlice := cast.ToStringSlice(s)
	cleaned := []string{}
	for _, v := range stringSlice {
		cleaned = append(cleaned, gox.AddSlashes(v))
	}
	return "'" + strings.Join(cleaned, "','") + "'"
}

// Insert 新增记录
//  @param db gorose.IOrm
//  @param table string
//  @param data map[string]interface{}
//  @return id int64
//  @return err error
func Insert(db gorose.IOrm, table string, data map[string]interface{}) (id int64, err error) {
	id, err = db.Table(table).InsertGetId(data)
	if err != nil {
		zap.L().Error(err.Error())
		return 0, err
	}

	return id, nil
}

// Update 更新记录
//  @param db gorose.IOrm
//  @param table string
//  @param data map[string]interface{}
//  @param where string
//  @param params ...interface{}
//  @return affectedCounts int64
//  @return err error
func Update(db gorose.IOrm, table string, data map[string]interface{}, where string, params ...interface{}) (affectedCounts int64, err error) {
	dataPlaceholders := []string{}
	dataValues := []interface{}{}
	for k, v := range data {
		dataPlaceholder := fmt.Sprintf("%s=?", k)
		dataPlaceholders = append(dataPlaceholders, dataPlaceholder)
		dataValues = append(dataValues, v)
	}
	dataPlaceholdersStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(dataPlaceholders)), ","), "[]")

	allValues := dataValues
	for _, v := range params {
		allValues = append(allValues, v)
	}

	sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, dataPlaceholdersStr, where)
	affectedCounts, err = Execute(db, sql, allValues...)
	if err != nil {
		return 0, err
	}

	return affectedCounts, nil
}

// Delete 删除记录
//  @param db gorose.IOrm
//  @param table string
//  @param where string
//  @param params ...interface{}
//  @return affectedCounts int64
//  @return err error
func Delete(db gorose.IOrm, table string, where string, params ...interface{}) (affectedCounts int64, err error) {
	// DELETE FROM table WHERE aaa=?
	sql := fmt.Sprintf("DELETE FROM %s WHERE %s", table, where)
	affectedCounts, err = Execute(db, sql, params...)
	if err != nil {
		return 0, err
	}

	return affectedCounts, nil
}

// Execute 执行原生SQL
//  @param db gorose.IOrm
//  @param sql string
//  @param params ...interface{}
//  @return affectedCounts int64
//  @return err error
func Execute(db gorose.IOrm, sql string, params ...interface{}) (affectedCounts int64, err error) {
	affectedCounts, err = db.Execute(sql, params...)
	if err != nil {
		zap.L().Error(err.Error())
		return 0, err
	}

	return affectedCounts, nil
}

// Begin 开始事务
//  @param db gorose.IOrm
//  @return error
func Begin(db gorose.IOrm) error {
	if err := db.Begin(); err != nil {
		zap.L().Error(err.Error())
		return err
	}
	return nil
}

// Commit 提交事务
//	提交失败会自动回滚
//  @param db gorose.IOrm
//  @return error
func Commit(db gorose.IOrm) error {
	if err := db.Commit(); err != nil {
		zap.L().Error(err.Error())
		Rollback(db)
		return err
	}
	return nil
}

// Rollback 回滚事务
//  @param db gorose.IOrm
func Rollback(db gorose.IOrm) {
	if err := db.Rollback(); err != nil {
		zap.L().Error(err.Error())
	}
}
