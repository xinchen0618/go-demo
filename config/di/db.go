package di

import (
	"fmt"
	"os"
	"time"

	"go-demo/config"
	"go-demo/pkg/gox"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gohouse/gorose/v2"
	"github.com/golang-module/carbon/v2"
	"go.uber.org/zap"
)

// SQL log
type sqlLogger struct{}

func (sqlLogger) Sql(sqlStr string, runtime time.Duration) {
	f, err := os.OpenFile(config.GetString("sql_log"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		zap.L().Error(err.Error())
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			zap.L().Error(err.Error())
		}
	}(f)

	if _, err := fmt.Fprintf(f, "[SQL] [%s] %s --- %s\n", carbon.Now().ToDateTimeString(), runtime.String(), sqlStr); err != nil {
		zap.L().Error(err.Error())
	}
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

// MySQL DEMO, 从这里开始定义项目中的DB
var (
	demoDbEngine *gorose.Engin
	demoDbOnce   gox.Once
)

func DemoDb() gorose.IOrm {
	_ = demoDbOnce.Do(func() error {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
			config.Get("mysql_username"),
			config.Get("mysql_password"),
			config.Get("mysql_host"),
			config.Get("mysql_port"),
			config.Get("mysql_dbname"),
			config.Get("mysql_charset"),
		)
		var err error
		demoDbEngine, err = gorose.Open(&gorose.Config{
			Driver:          "mysql",
			Dsn:             dsn,
			SetMaxOpenConns: config.GetInt("mysql_max_open_conns"),
			SetMaxIdleConns: config.GetInt("mysql_max_idle_conns"),
		})
		if err != nil {
			panic(err) // 即便这里不panic, 调用者在nil指针上调用db方法也会panic
		}

		if config.GetString("sql_log") != "" {
			demoDbEngine.SetLogger(sqlLogger{})
		}

		return nil
	})

	return demoDbEngine.NewOrm()
}
