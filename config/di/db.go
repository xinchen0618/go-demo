package di

import (
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gohouse/gorose/v2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// mysql
var (
	dbEngine *gorose.Engin
	dbOnce   sync.Once
)

func Db() gorose.IOrm {
	dbOnce.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
			viper.Get("mysql.username"), viper.Get("mysql.password"), viper.Get("mysql.host"),
			viper.Get("mysql.port"), viper.Get("mysql.dbname"), viper.Get("mysql.charset"))
		var err error
		dbEngine, err = gorose.Open(&gorose.Config{Driver: "mysql", Dsn: dsn, SetMaxOpenConns: 100, SetMaxIdleConns: 100})
		if err != nil {
			zap.L().Error(err.Error())
		}
	})

	return dbEngine.NewOrm()
}
