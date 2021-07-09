package di

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gohouse/gorose/v2"
	"github.com/spf13/viper"
)

// 先定义私有变量存放实例(保证实例不能被外部修改), 然后在init()中初始化实例(保证实例只初始化一次), 最后定义公共方法获取实例使用
var dbEngine *gorose.Engin
var cacheRedis *redis.Client
var jwtRedis *redis.Client

func init() {
	/* Log */
	logFile, err := os.OpenFile("/var/log/golang_error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.Llongfile | log.Lmicroseconds | log.Ldate)

	/* 配置 */
	viper.SetConfigName("config")   // name of config file (without extension)
	viper.SetConfigType("yaml")     // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("./config") // path to look for the config file in
	err = viper.ReadInConfig()      // Find and read the config file
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
}

func Db() gorose.IOrm {
	return dbEngine.NewOrm()
}

func Ctx() context.Context {
	return context.Background()
}

func CacheRedis() *redis.Client {
	return cacheRedis
}

func JwtRedis() *redis.Client {
	return jwtRedis
}
