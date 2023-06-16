// Package dbx MySQL增删改查操作封装
//
//	MySQL=>Golang数据类型映射:
//		bigint/int/smallint/tinyint => int64,
//		float/double => float64,
//		varchar/char/longtext/text/mediumtext/tinytext/decimal/datetime/timestamp/date/time => string,
package dbx

import (
	"fmt"
	"math"
	"strings"

	"go-demo/pkg/gox"

	"github.com/gohouse/gorose/v2"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

// FetchAll 获取多行记录返回map
//
//	params 不支持切片类型.
func FetchAll(db gorose.IOrm, sql string, params ...any) ([]map[string]any, error) {
	rows, err := db.Query(sql, params...)
	if err != nil {
		zap.L().Error(err.Error())
		return nil, err
	}

	result := make([]map[string]any, 0)
	for _, v := range rows {
		result = append(result, v)
	}

	return result, nil
}

// TakeAll 获取多行记录至struct
//
//	p 为接收结果的指针, json tag是必须的.
func TakeAll(p any, db gorose.IOrm, sql string, params ...any) error {
	items, err := FetchAll(db, sql, params...)
	if err != nil {
		return err
	}
	if err := gox.TypeCast(items, p); err != nil {
		return err
	}

	return nil
}

// FetchOne 获取一行记录返回map
//
//	查询时会自动添加限制LIMIT 1.
func FetchOne(db gorose.IOrm, sql string, params ...any) (map[string]any, error) {
	sql = strings.TrimSpace(sql)
	if "SELECT" == strings.ToUpper(lo.Substring(sql, 0, 6)) && strings.ToUpper(lo.Substring(sql, -7, math.MaxInt)) != "LIMIT 1" {
		sql += " LIMIT 1"
	}

	rows, err := FetchAll(db, sql, params...)
	if err != nil {
		return nil, err
	}

	if 0 == len(rows) {
		return map[string]any{}, nil
	}

	return rows[0], nil
}

// TakeOne 获取一行记录至struct
//
//	p 为接收结果的指针, json tag是必须的.
func TakeOne(p any, db gorose.IOrm, sql string, params ...any) error {
	item, err := FetchOne(db, sql, params...)
	if err != nil {
		return err
	}
	if err := gox.TypeCast(item, p); err != nil {
		return err
	}

	return nil
}

// FetchValue 获取一个值
func FetchValue(db gorose.IOrm, sql string, params ...any) (any, error) {
	row, err := FetchOne(db, sql, params...)
	if err != nil {
		return nil, err
	}

	for _, value := range row {
		return value, nil
	}

	// 0 == len(row)
	return nil, nil
}

// TakeValue 获取一个值至指定类型
//
//	p 为接收结果的指针.
func TakeValue(p any, db gorose.IOrm, sql string, params ...any) error {
	value, err := FetchValue(db, sql, params...)
	if err != nil {
		return err
	}
	if err := gox.TypeCast(value, p); err != nil {
		return err
	}

	return nil
}

// FetchColumn 获取一列值
func FetchColumn(db gorose.IOrm, sql string, params ...any) ([]any, error) {
	rows, err := FetchAll(db, sql, params...)
	if err != nil {
		return nil, err
	}

	values := make([]any, 0)
	for _, row := range rows {
		for _, value := range row {
			values = append(values, value)
			break
		}
	}

	return values, nil
}

// TakeColumn 获取一列值至指定类型
//
//	p 为接收结果的指针.
func TakeColumn(p any, db gorose.IOrm, sql string, params ...any) error {
	values, err := FetchColumn(db, sql, params...)
	if err != nil {
		return err
	}
	if err := gox.TypeCast(values, p); err != nil {
		return err
	}

	return nil
}

// Slice2in Slice转IN条件
//
//	Golang SQL驱动不支持IN(?), 使用fmt.Sprint("IN(%s)", Slice2in(s)).
func Slice2in(s any) string {
	stringSlice := cast.ToStringSlice(s)
	cleaned := make([]string, 0)
	for _, v := range stringSlice {
		cleaned = append(cleaned, gox.AddSlashes(v))
	}
	return "'" + strings.Join(cleaned, "','") + "'"
}

// Insert 新增记录
func Insert(db gorose.IOrm, table string, data map[string]any) (id int64, err error) {
	id, err = db.Table(table).InsertGetId(data)
	if err != nil {
		zap.L().Error(err.Error())
		return 0, err
	}

	return id, nil
}

// InsertBatch 批量新增记录
func InsertBatch(db gorose.IOrm, table string, data []map[string]any) (affectedRows int64, err error) {
	affectedRows, err = db.Table(table).Insert(data)
	if err != nil {
		zap.L().Error(err.Error())
		return 0, err
	}

	return affectedRows, nil
}

// Update 更新记录
func Update(db gorose.IOrm, table string, data map[string]any, where string, params ...any) (affectedRows int64, err error) {
	dataPlaceholders := make([]string, 0)
	dataValues := make([]any, 0)
	for k, v := range data {
		dataPlaceholder := fmt.Sprintf("%s=?", k)
		dataPlaceholders = append(dataPlaceholders, dataPlaceholder)
		dataValues = append(dataValues, v)
	}
	dataPlaceholdersStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(dataPlaceholders)), ","), "[]")

	allValues := dataValues
	allValues = append(allValues, params...)

	sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, dataPlaceholdersStr, where)
	affectedRows, err = Execute(db, sql, allValues...)
	if err != nil {
		return 0, err
	}

	return affectedRows, nil
}

// Delete 删除记录
func Delete(db gorose.IOrm, table string, where string, params ...any) (affectedRows int64, err error) {
	sql := fmt.Sprintf("DELETE FROM %s WHERE %s", table, where)
	affectedRows, err = Execute(db, sql, params...)
	if err != nil {
		return 0, err
	}

	return affectedRows, nil
}

// Execute 执行原生SQL
func Execute(db gorose.IOrm, sql string, params ...any) (affectedRows int64, err error) {
	affectedRows, err = db.Execute(sql, params...)
	if err != nil {
		zap.L().Error(err.Error())
		return 0, err
	}

	return affectedRows, nil
}

// Begin 手动开始事务
func Begin(db gorose.IOrm) error {
	if err := db.Begin(); err != nil {
		zap.L().Error(err.Error())
		return err
	}
	return nil
}

// Commit 手动提交事务
//
//	提交失败会自动回滚.
func Commit(db gorose.IOrm) error {
	if err := db.Commit(); err != nil {
		zap.L().Error(err.Error())
		Rollback(db)
		return err
	}
	return nil
}

// Rollback 手动回滚事务
func Rollback(db gorose.IOrm) {
	if err := db.Rollback(); err != nil {
		zap.L().Error(err.Error())
	}
}
