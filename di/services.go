package di

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gohouse/gorose/v2"
	"github.com/spf13/viper"
	"log"
	"os"
)

var db gorose.IOrm
var ctx = context.Background()
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

	/* mysql */
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
		viper.Get("mysql.username"), viper.Get("mysql.password"), viper.Get("mysql.host"),
		viper.Get("mysql.port"), viper.Get("mysql.dbname"), viper.Get("mysql.charset"))
	engine, err := gorose.Open(&gorose.Config{Driver: "mysql", Dsn: dsn})
	if err != nil {
		panic(err)
	}
	db = engine.NewOrm()

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
	return db
}

func Ctx() context.Context {
	return ctx
}

func CacheRedis() *redis.Client {
	return cacheRedis
}

func JwtRedis() *redis.Client {
	return jwtRedis
}
