package gormx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type gormZapLogger struct {
	logger.Config
}

// NewLogger SQL 日志记录到 zap
//
//	有效属性:
//		SlowThreshold
//		IgnoreRecordNotFoundError
//		LogLevel
func NewLogger(config logger.Config) logger.Interface {
	return &gormZapLogger{
		config,
	}
}

func (l *gormZapLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level

	return l
}

func (l *gormZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		zap.L().Info(fmt.Sprintf(msg, data...),
			zap.String("caller", utils.FileWithLineNum()),
		)
	}
}

func (l *gormZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		zap.L().Warn(fmt.Sprintf(msg, data...),
			zap.String("caller", utils.FileWithLineNum()),
		)
	}
}

func (l *gormZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		zap.L().Error(fmt.Sprintf(msg, data...),
			zap.String("caller", utils.FileWithLineNum()),
		)
	}
}

func (l *gormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		zap.L().Error("gorm", zap.Error(err), zap.String("sql", sql), zap.String("elapsed", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)), zap.Int64("rows", rows), zap.String("caller", utils.FileWithLineNum()))
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		zap.L().Warn("gorm", zap.String("slow sql", sql), zap.String("elapsed", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)), zap.Int64("rows", rows), zap.String("caller", utils.FileWithLineNum()))
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		zap.L().Info("gorm", zap.String("sql", sql), zap.String("elapsed", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)), zap.Int64("rows", rows), zap.String("caller", utils.FileWithLineNum()))
	}
}
