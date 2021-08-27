package di

import (
	"fmt"
	"go-demo/config"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gohouse/gorose/v2"
	"github.com/golang-module/carbon"
	"github.com/matryer/resync"
)

// mysql
var (
	dbEngine *gorose.Engin
	dbOnce   resync.Once
	dbError  error
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
	dbOnce.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
			config.Get("mysql_username"), config.Get("mysql_password"), config.Get("mysql_host"),
			config.Get("mysql_port"), config.Get("mysql_dbname"), config.Get("mysql_charset"))
		dbEngine, dbError = gorose.Open(&gorose.Config{Driver: "mysql", Dsn: dsn, SetMaxOpenConns: 100, SetMaxIdleConns: 100})
		if dbError != nil {
			Logger().Error(dbError.Error())
			return
		}

		// print SQL to console
		if "dev" == os.Getenv("RUNTIME_ENV") || "testing" == os.Getenv("RUNTIME_ENV") {
			dbEngine.SetLogger(sqlLogger{})
		}
	})
	if dbError != nil {
		dbOnce.Reset()
	}

	return dbEngine.NewOrm()
}
