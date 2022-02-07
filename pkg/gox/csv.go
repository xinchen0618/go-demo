package gox

import (
	"encoding/csv"
	"os"

	"go.uber.org/zap"
)

// PutCsv 创建CSV文件
//	文件不存在会创建, 文件存在会覆盖写入
//  @param name string
//  @param data [][]string
//  @return error
func PutCsv(name string, data [][]string) error {
	f, err := os.Create(name) //创建文件
	if err != nil {
		zap.L().Error(err.Error())
		return err
	}
	defer f.Close()

	if _, err := f.WriteString("\xEF\xBB\xBF"); err != nil { // 写入UTF-8 BOM
		zap.L().Error(err.Error())
		return err
	}

	w := csv.NewWriter(f)                    //创建一个新的写入文件流
	if err := w.WriteAll(data); err != nil { //写入数据
		zap.L().Error(err.Error())
		return err
	}
	w.Flush()

	return nil
}
