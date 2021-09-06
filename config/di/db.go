package di

import (
	"fmt"
	"go-demo/config"
	"go-demo/pkg/gox"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gohouse/gorose/v2"
	"github.com/golang-module/carbon"
)

// mysql, 成功则仅初始化一次, 失败允许再次初始化
var (
	dbEngine *gorose.Engin
	dbOnce   gox.Once
)

// print SQL
type sqlLogger struct {
}

func (sqlLogger) Sql(sqlStr string, runtime time.Duration) {
	fmt.Printf("[SQL] [%s] %s --- %s\n", carbon.Now().ToDateTimeString(), runtime.String(), sqlStr)
}
func (sqlLogger) Slow(sqlStr string, runtime time.Duration) {
}
func (sqlLogger) Error(msg string) {
}
func (sqlLogger) EnableSqlLog() bool {
	return true
}
func (sqlLogger) EnableErrorLog() bool {
	return false
}
func (sqlLogger) EnableSlowLog() float64 {
	return 0
}

func Db() gorose.IOrm {
	_ = dbOnce.Do(func() error {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
			config.Get("mysql_username"), config.Get("mysql_password"), config.Get("mysql_host"),
			config.Get("mysql_port"), config.Get("mysql_dbname"), config.Get("mysql_charset"))
		var err error
		dbEngine, err = gorose.Open(&gorose.Config{Driver: "mysql", Dsn: dsn, SetMaxOpenConns: 100, SetMaxIdleConns: 100})
		if err != nil {
			panic(err) // 即便这里不panic, 调用者在nil指针上调用db方法也会panic
		}
		if gox.InSlice(config.GetRuntimeEnv(), []string{"dev", "testing"}) { // print SQL to console
			dbEngine.SetLogger(sqlLogger{})
		}

		return nil
	})

	return dbEngine.NewOrm()
}
