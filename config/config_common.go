package config

// 公共配置
func init() {
	// 配置value支持string/int/[]string/[]int
	configure["common"] = map[string]interface{}{
		// 错误日志路径
		"error_log": "/var/log/golang_error.log",

		// Goroutine池大小
		"worker_pool": 40960,

		// DB
		"mysql_host":     "127.0.0.1",
		"mysql_port":     3306,
		"mysql_username": "root",
		"mysql_password": "cx654321",
		"mysql_dbname":   "test",
		"mysql_charset":  "utf8mb4",

		// Redis
		"redis_host":          "127.0.0.1",
		"redis_port":          6379,
		"redis_auth":          "",
		"redis_index_cache":   0, // 缓存
		"redis_index_jwt":     1, // JWT
		"redis_index_storage": 2, // 存储
		"redis_index_queue":   3, // 消息队列
	}
}
