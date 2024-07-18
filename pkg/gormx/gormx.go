package gormx

import (
	"fmt"

	gormcache "github.com/asjdf/gorm-cache/cache"
	gormcacheconfig "github.com/asjdf/gorm-cache/config"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type NewDBReq struct {
	LogLevel     string
	UserName     string
	Password     string
	Host         string
	Port         int
	DBName       string
	CharSet      string
	MaxIdleConns int
	MaxOpenConns int
}

// NewDB 创建数据库链接
func NewDB(req NewDBReq) (*gorm.DB, error) {
	// 日志
	loggerConfig := logger.Config{
		SlowThreshold:             0,
		IgnoreRecordNotFoundError: true,
		LogLevel:                  logger.Info,
	}
	switch req.LogLevel {
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
		req.UserName,
		req.Password,
		req.Host,
		req.Port,
		req.DBName,
		req.CharSet,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 NewLogger(loggerConfig),
	})
	if err != nil {
		zap.L().Error(err.Error())
		return nil, err
	}

	// 缓存
	cache, err := gormcache.NewGorm2Cache(&gormcacheconfig.CacheConfig{
		CacheLevel: gormcacheconfig.CacheLevelOff, // 缓存有 key 冲突的 BUG, 这里仅使用它的 singleflight
	})
	if err != nil {
		zap.L().Error(err.Error())
		return nil, err
	}
	if err := db.Use(cache); err != nil {
		zap.L().Error(err.Error())
		return nil, err
	}

	// 连接池
	sqlDB, err := db.DB()
	if err != nil {
		zap.L().Error(err.Error())
		return nil, err
	}
	sqlDB.SetMaxIdleConns(req.MaxIdleConns)
	sqlDB.SetMaxOpenConns(req.MaxOpenConns)

	return db, nil
}
