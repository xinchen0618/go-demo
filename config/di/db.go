package di

import (
	"go-demo/config"
	"go-demo/pkg/gormx"
	"go-demo/pkg/gox"

	"gorm.io/gorm"
)

/**************** DEMO DB *************************************************/
var (
	demoDB     *gorm.DB
	demoDBOnce gox.Once
)

func DemoDB() *gorm.DB {
	_ = demoDBOnce.Do(func() (err error) {
		demoDB, err = gormx.NewDB(gormx.NewDBReq{
			LogLevel:     config.GetString("error_log_level"),
			UserName:     config.GetString("mysql_username"),
			Password:     config.GetString("mysql_password"),
			Host:         config.GetString("mysql_host"),
			Port:         config.GetInt("mysql_port"),
			DBName:       config.GetString("mysql_dbname"),
			Charset:      config.GetString("mysql_charset"),
			MaxIdleConns: config.GetInt("mysql_max_idle_conns"),
			MaxOpenConns: config.GetInt("mysql_max_open_conns"),
		})

		return
	})

	return demoDB
}
