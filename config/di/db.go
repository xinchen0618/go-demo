package di

import (
	"fmt"
	"go-demo/config"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gohouse/gorose/v2"
)

// mysql
var (
	dbEngine *gorose.Engin
	dbOnce   sync.Once
)

func Db() gorose.IOrm {
	dbOnce.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
			config.Get("mysql_username"), config.Get("mysql_password"), config.Get("mysql_host"),
			config.Get("mysql_port"), config.Get("mysql_dbname"), config.Get("mysql_charset"))
		var err error
		dbEngine, err = gorose.Open(&gorose.Config{Driver: "mysql", Dsn: dsn, SetMaxOpenConns: 100, SetMaxIdleConns: 100})
		if err != nil {
			Logger().Error(err.Error())
		}
	})

	return dbEngine.NewOrm()
}
