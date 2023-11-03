// Package config 配置实现
package config

func init() {
	// 公共配置
	const env = "common"

	if _, ok := configure[env]; !ok {
		configure[env] = map[string]any{}
	}
	for k, v := range map[string]any{
		/************ 配置 START ****************/

		// 日志路径
		"app_log": "/var/log/golang_app.log",
		// 彩色输出日志
		"colorful_log": true,
		// SQL 日志级别, 生产通常使用 Error, 开发使用 Info
		"sql_log_level": 4, // 1-Silent,2-Error,3-Warn,4-Info

		// 公共 Goroutine 池大小
		"worker_pool": 409600,

		// 限流 QPS
		"qps_limit": 40000,

		// 超时控制, 秒
		"timeout": 30,

		// DB DEMO
		"mysql_host":           "127.0.0.1",
		"mysql_port":           3306,
		"mysql_username":       "root",
		"mysql_password":       "cx654321",
		"mysql_dbname":         "test",
		"mysql_charset":        "utf8mb4",
		"mysql_max_open_conns": 140,
		"mysql_max_idle_conns": 30,

		// Redis DEMO
		"redis_host":          "127.0.0.1",
		"redis_port":          6379,
		"redis_auth":          "",
		"redis_index_cache":   0, // 缓存
		"redis_index_jwt":     1, // JWT
		"redis_index_storage": 2, // 存储
		"redis_index_queue":   3, // 消息队列

		/************ 配置 END ******************/
	} {
		configure[env][k] = v
	}
}
