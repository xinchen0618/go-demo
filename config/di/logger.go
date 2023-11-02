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
	logFile, err := os.OpenFile(config.GetString("error_log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o664)
	if err != nil {
		panic(err)
	}
	fileSyncer := zapcore.AddSync(logFile)      // 输出到文件
	consoleSyncer := zapcore.AddSync(os.Stdout) // 输出到 console
	// 创建编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 彩色输出. json 格式输出时不需要
	encoder := zapcore.NewConsoleEncoder(encoderConfig)          // console 格式输出. json 格式输出为 NewJSONEncoder()
	// 创建 Core
	zapCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(fileSyncer, consoleSyncer), zapcore.DebugLevel) // 允许记录所有级别日志
	// 创建 Logger
	zapLogger = zap.New(zapCore, zap.AddStacktrace(zapcore.ErrorLevel)) // 错误日志记录栈信息
	// 替换 zap 包中全局的 zapLogger 实例, 后续在其他包中只需使用 zap.L() 调用即可
	zap.ReplaceGlobals(zapLogger)
}

// Logger 日志
func Logger() *zap.Logger {
	return zapLogger
}
