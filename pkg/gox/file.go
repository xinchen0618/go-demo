// Package gox Golang 增强函数
package gox

import (
	"encoding/csv"
	"os"

	"go.uber.org/zap"
)

// PutCSV 创建 CSV 文件
//
//	文件不存在会创建, 文件存在会覆盖写入.
func PutCSV(name string, data [][]string) error {
	f, err := os.Create(name) // 创建文件
	if err != nil {
		zap.L().Error(err.Error())
		return err
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			zap.L().Error(err.Error())
		}
	}(f)

	if _, err := f.WriteString("\xEF\xBB\xBF"); err != nil { // 写入 UTF-8 BOM
		zap.L().Error(err.Error())
		return err
	}

	w := csv.NewWriter(f)                    // 创建一个新的写入文件流
	if err := w.WriteAll(data); err != nil { // 写入数据
		zap.L().Error(err.Error())
		return err
	}
	w.Flush()

	return nil
}
