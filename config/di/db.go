package di

import (
	"fmt"
	"log"
	"os"

	"go-demo/config"
	"go-demo/pkg/gox"

	gormcache "github.com/asjdf/gorm-cache/cache"
	gormcacheconfig "github.com/asjdf/gorm-cache/config"
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
		file := os.Stdout // 没有配置日志文件输出到 console
		sqlLog := config.GetString("sql_log")
		if sqlLog != "" {
			var err error
			file, err = os.OpenFile(sqlLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
			if err != nil {
				Logger().Error(err.Error())
				return err
			}
		}
		logLevel := logger.Info // Silent,Error,Warn,Info
		switch config.GetString("sql_log_level") {
		case "Silent":
			logLevel = logger.Silent
		case "Error":
			logLevel = logger.Error
		case "Warn":
			logLevel = logger.Warn
		case "Info":
			logLevel = logger.Info
		}
		colorful := config.GetBool("sql_log_colorful")

		gormLogger = logger.New(
			log.New(file, "\r\n", log.LstdFlags),
			logger.Config{
				LogLevel:                  logLevel,
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
			SkipDefaultTransaction: true,
			Logger:                 newGormLogger(),
		})
		if err != nil {
			Logger().Error(err.Error())
			return err
		}

		// 缓存
		cache, err := gormcache.NewGorm2Cache(&gormcacheconfig.CacheConfig{
			CacheLevel: gormcacheconfig.CacheLevelOff, // 缓存有 key 冲突的 BUG, 这里仅使用它的 singleflight
			// CacheStorage:         gormcachestorage.NewRedis(&gormcachestorage.RedisStoreConfig{Client: CacheRedis()}),
			// InvalidateWhenUpdate: true,            // when you create/update/delete objects, invalidate cache
			// CacheTTL:             6 * 3600 * 1000, // ms
			// CacheMaxItemCnt:      1,               // if length of objects retrieved one single time, exceeds this number, then don't cache
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
