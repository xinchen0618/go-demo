// Package config 配置实现
package config

func init() {
	// 生产环境配置
	const env = "prod"

	if RuntimeEnv() != env {
		return
	}
	if _, ok := configure[env]; !ok {
		configure[env] = map[string]any{}
	}
	for k, v := range map[string]any{
		/************ 配置 START **************/

		// 运行端口
		"server_port": 8090,

		// JWT 密钥, JWT 配套有白名单功能不必担心秘钥泄露的问题
		"jwt_secret": "btRZ5QHXX9VjfYhfGGHdCTcWiwQ6WFJXq9ZCwdqZwzk2ZfhceM9K3V5UGKsYLd9m",

		// 日志
		"error_log_level":   "Error",
		"error_log_encoder": "JSON",

		/************ 配置 END ****************/
	} {
		configure[env][k] = v
	}
}
