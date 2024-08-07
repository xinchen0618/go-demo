// Package di 服务注入
package di

import (
	"os"

	"go-demo/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var zapLogger *zap.Logger

func init() { // 日志服务最为基础, 日志初始化失败, 程序不允许启动
	// 创建输出位置
	syncers := make([]zapcore.WriteSyncer, 0) // NewMultiWriteSyncer() 可以添加多个 syncer, 逗号分隔
	errorLog := config.GetString("error_log")
	if errorLog != "" { // 输出到文件
		logFile, err := os.OpenFile(errorLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o664)
		if err != nil {
			panic(err)
		}
		fileSyncer := zapcore.AddSync(logFile)
		syncers = append(syncers, fileSyncer)
	} else { // 输出到控制台
		consoleSyncer := zapcore.AddSync(os.Stdout)
		syncers = append(syncers, consoleSyncer)
	}
	// 创建编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05") // 自定义时间格式
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// 创建 Core
	logLevel := zapcore.DebugLevel
	switch config.GetString("error_log_level") {
	case "Debug":
		logLevel = zapcore.DebugLevel
	case "Info":
		logLevel = zapcore.InfoLevel
	case "Warn":
		logLevel = zapcore.WarnLevel
	case "Error":
		logLevel = zapcore.ErrorLevel
	}
	zapCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(syncers...), logLevel) // 允许记录所有级别日志
	// 创建 Logger
	zapLogger = zap.New(zapCore, zap.AddStacktrace(zapcore.ErrorLevel)) // 错误日志记录栈信息
	// 替换 zap 包中全局的 zapLogger 实例, 后续在其他包中只需使用 zap.L() 调用即可
	zap.ReplaceGlobals(zapLogger)
}

// Logger 日志
func Logger() *zap.Logger {
	return zapLogger
}
