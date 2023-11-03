package di

import (
	"fmt"
	"log"
	"os"

	"go-demo/config"
	"go-demo/pkg/gox"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

/****************** GORM 日志定义 ***********************************/
var (
	gormLogger     logger.Interface
	gormLoggerOnce gox.Once
)

func newGormLogger() logger.Interface {
	_ = gormLoggerOnce.Do(func() error {
		sqlLog := config.GetString("app_log")
		if sqlLog == "" {
			gormLogger = logger.Default.LogMode(logger.Silent) // 未设定日志路径不记录
			return nil
		}
		file, err := os.OpenFile(sqlLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
		if err != nil {
			Logger().Error(err.Error())
			return err
		}
		logLevel := config.GetInt("sql_log_level") // 1-Silent,2-Error,3-Warn,4-Info
		if logLevel == 0 {
			logLevel = 2 // 默认记录级别为 Error
		}
		colorful := config.GetBool("colorful_log")

		gormLogger = logger.New(
			log.New(file, "\r\n", log.LstdFlags),
			logger.Config{
				LogLevel:                  logger.LogLevel(logLevel),
				IgnoreRecordNotFoundError: true,
				ParameterizedQueries:      false, // Raw() 方法无效
				Colorful:                  colorful,
			},
		)

		return nil
	})

	return gormLogger
}

/**************** DEMO DB *************************************************/

var (
	demoDB     *gorm.DB
	demoDBOnce gox.Once
)

func DemoDB() *gorm.DB {
	_ = demoDBOnce.Do(func() error {
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			config.GetString("mysql_username"),
			config.GetString("mysql_password"),
			config.GetString("mysql_host"),
			config.GetInt("mysql_port"),
			config.GetString("mysql_dbname"),
			config.GetString("mysql_charset"),
		)
		var err error
		demoDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: newGormLogger(),
		})
		if err != nil {
			Logger().Error(err.Error())
			return err
		}

		// 连接池
		sqlDB, err := demoDB.DB()
		if err != nil {
			Logger().Error(err.Error())
			return err
		}
		sqlDB.SetMaxIdleConns(config.GetInt("mysql_max_idle_conns"))
		sqlDB.SetMaxOpenConns(config.GetInt("mysql_max_open_conns"))

		return nil
	})

	return demoDB
}
