package di

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gohouse/gorose/v2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 先定义私有变量存放实例(保证实例不能被外部修改), 然后在Init()中初始化实例(once保证实例只初始化一次), 最后定义公共方法获取实例使用
var once sync.Once

var (
	dbEngine   *gorose.Engin
	cacheRedis *redis.Client
	jwtRedis   *redis.Client
	logger     *zap.Logger
)

func Init() {
	once.Do(func() {
		/* 配置 */
		viper.SetConfigName("config")   // name of config file (without extension)
		viper.SetConfigType("yaml")     // REQUIRED if the config file does not have the extension in the name
		viper.AddConfigPath("./config") // path to look for the config file in
		err := viper.ReadInConfig()     // Find and read the config file
		if err != nil {                 // Handle errors reading the config file
			panic(err)
		}
		runtimeEnv := os.Getenv("RUNTIME_ENV") // 多环境配置
		if runtimeEnv == "" {
			runtimeEnv = "prod"
		}
		viper.SetConfigName("config_" + runtimeEnv)
		viper.AddConfigPath("./config")
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
		zapCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(writeSyncer, zapcore.AddSync(os.Stdout)), zapcore.DebugLevel)
		logger = zap.New(zapCore, zap.AddStacktrace(zapcore.ErrorLevel)) // 错误日志记录栈信息

		/* mysql */
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
			viper.Get("mysql.username"), viper.Get("mysql.password"), viper.Get("mysql.host"),
			viper.Get("mysql.port"), viper.Get("mysql.dbname"), viper.Get("mysql.charset"))
		dbEngine, err = gorose.Open(&gorose.Config{Driver: "mysql", Dsn: dsn, SetMaxOpenConns: 100, SetMaxIdleConns: 100})
		if err != nil {
			panic(err)
		}

		/* redis */
		cacheRedis = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", viper.Get("redis.host"), viper.Get("redis.port")),
			Password: viper.GetString("redis.auth"),
			DB:       viper.GetInt("redis.index.cache"),
		})
		jwtRedis = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", viper.Get("redis.host"), viper.Get("redis.port")),
			Password: viper.GetString("redis.auth"),
			DB:       viper.GetInt("redis.index.jwt"),
		})
	})
}

func Db() gorose.IOrm {
	return dbEngine.NewOrm()
}

func CacheRedis() *redis.Client {
	return cacheRedis
}

func JwtRedis() *redis.Client {
	return jwtRedis
}

func Logger() *zap.Logger {
	return logger
}
