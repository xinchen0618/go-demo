package dbx

import (
	"fmt"
	"strings"

	"github.com/gohouse/gorose/v2"
	"go.uber.org/zap"
)

// FetchAll 获取多行记录
//  @param db gorose.IOrm
//  @param sql string
//  @param params ...interface{} 不支持切片
//  @return []gorose.Data
//  @return error
func FetchAll(db gorose.IOrm, sql string, params ...interface{}) ([]gorose.Data, error) {
	rows, err := db.Query(sql, params...)
	if err != nil {
		zap.L().Error(err.Error())
		return []gorose.Data{}, err
	}

	return rows, nil
}

// FetchOne 获取一行记录
//  @param db gorose.IOrm
//  @param sql string
//  @param params ...interface{}
//  @return gorose.Data
//  @return error
func FetchOne(db gorose.IOrm, sql string, params ...interface{}) (gorose.Data, error) {
	rows, err := FetchAll(db, sql, params...)
	if err != nil {
		return gorose.Data{}, err
	}

	if 0 == len(rows) {
		return gorose.Data{}, nil
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
		return gorose.Data{}, err
	}

	for _, value := range row {
		if nil == value {
			return "", nil
		}
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

// Execute 执行SQL语句
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
