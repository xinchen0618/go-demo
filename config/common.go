package config

// 公共配置
func init() {
	if _, ok := configure["common"]; !ok {
		configure["common"] = map[string]any{}
	}

	for k, v := range map[string]any{
		// 错误日志路径
		"error_log": "/var/log/golang_error.log",

		// SQL日志
		"sql_log": "/var/log/golang_sql.log",

		// 公共Goroutine池大小
		"worker_pool": 409600,

		// 限流QPS
		"qps_limit": 40000,

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
	} {
		configure["common"][k] = v
	}
}
