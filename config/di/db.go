package di

import (
	"fmt"

	"go-demo/config"
	"go-demo/pkg/gormx"
	"go-demo/pkg/gox"

	gormcache "github.com/asjdf/gorm-cache/cache"
	gormcacheconfig "github.com/asjdf/gorm-cache/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

/**************** DEMO DB *************************************************/
var (
	demoDB     *gorm.DB
	demoDBOnce gox.Once
)

func DemoDB() *gorm.DB {
	_ = demoDBOnce.Do(func() error {
		// 日志
		loggerConfig := logger.Config{
			SlowThreshold:             0,
			IgnoreRecordNotFoundError: true,
			LogLevel:                  logger.Info,
		}
		switch config.GetString("error_log_level") {
		case "Info":
			loggerConfig.LogLevel = logger.Info
		case "Warn":
			loggerConfig.LogLevel = logger.Warn
		case "Error":
			loggerConfig.LogLevel = logger.Error
		}
		// 连接
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
			SkipDefaultTransaction: true,
			Logger:                 gormx.NewLogger(loggerConfig),
		})
		if err != nil {
			Logger().Error(err.Error())
			return err
		}

		// 缓存
		cache, err := gormcache.NewGorm2Cache(&gormcacheconfig.CacheConfig{
			CacheLevel: gormcacheconfig.CacheLevelOff, // 缓存有 key 冲突的 BUG, 这里仅使用它的 singleflight
		})
		if err != nil {
			Logger().Error(err.Error())
			return err
		}
		if err := demoDB.Use(cache); err != nil {
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
