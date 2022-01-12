package config

// 测试环境配置
func init() {
	if GetRuntimeEnv() != "testing" {
		return
	}

	configure["testing"] = map[string]interface{}{
		// 运行端口
		"server_port": 8080,

		// JWT密钥, JWT配套有白名单动能不必担心秘钥泄露的问题
		"jwt_secret": "Xx4KJQ2AguFL5gWurcRJvVfDC5a2itLi53vFJN9wthYkrxtQbdeRDkWTHzAjnn5n",

		// SQL log, 缺省或为空时不记录, 注意文件需要读写权限
		"sql_log": "/var/log/golang_sql.log",
	}
}
