package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 初始化配置
var configOnce sync.Once

func Init() {
	configOnce.Do(func() {
		/* 配置 */
		viper.SetConfigName("config") // name of config file (without extension)
		viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		configPath := filepath.Dir(filepath.Dir(wd)) + "/config"
		viper.AddConfigPath(configPath) // path to look for the config file in
		err = viper.ReadInConfig()      // Find and read the config file
		if err != nil {                 // Handle errors reading the config file
			panic(err)
		}
		runtimeEnv := os.Getenv("RUNTIME_ENV") // 多环境配置
		if runtimeEnv == "" {
			runtimeEnv = "prod"
		}
		viper.SetConfigName("config_" + runtimeEnv)
		viper.AddConfigPath(configPath)
		err = viper.MergeInConfig()
		if err != nil {
			panic(err)
		}

		/* zap Log */
		logFile, err := os.OpenFile(viper.GetString("errorLog"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
		if err != nil {
			panic(err)
		}
		writeSyncer := zapcore.AddSync(logFile)
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		encoder := zapcore.NewConsoleEncoder(encoderConfig)
		zapCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(writeSyncer, zapcore.AddSync(os.Stdout)), zapcore.DebugLevel) // 输出到console和文件
		lg := zap.New(zapCore, zap.AddStacktrace(zapcore.ErrorLevel))                                                                 // 错误日志记录栈信息
		zap.ReplaceGlobals(lg)                                                                                                        // 替换zap包中全局的logger实例，后续在其他包中只需使用zap.L()调用即可
	})
}
