package config

// 测试环境配置
func init() {
	if GetRuntimeEnv() != "testing" {
		return
	}
	if _, ok := configure["testing"]; !ok {
		configure["testing"] = map[string]any{}
	}

	for k, v := range map[string]any{
		// 运行端口
		"server_port": 8080,

		// JWT密钥, JWT配套有白名单功能不必担心秘钥泄露的问题
		"jwt_secret": "Xx4KJQ2AguFL5gWurcRJvVfDC5a2itLi53vFJN9wthYkrxtQbdeRDkWTHzAjnn5n",

		// SQL日志
		"sql_log": "/var/log/golang_sql.log",
	} {
		configure["testing"][k] = v
	}
}
