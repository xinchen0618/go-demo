package di

import (
	"go-demo/config"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
)

// 日志服务最为基础, 日志初始化失败, 程序不允许启动
func init() {
	logFile, err := os.OpenFile(config.GetString("error_log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		panic(err)
	}
	writeSyncer := zapcore.AddSync(logFile)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	zapCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(writeSyncer, zapcore.AddSync(os.Stdout)), zapcore.DebugLevel) // 输出到console和文件
	logger = zap.New(zapCore, zap.AddStacktrace(zapcore.ErrorLevel))                                                              // 错误日志记录栈信息
}

func Logger() *zap.Logger {
	return logger
}
