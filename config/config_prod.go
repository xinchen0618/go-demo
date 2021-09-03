package config

// 生产环境配置
func init() {
	if GetRuntimeEnv() != "prod" {
		return
	}

	configure["prod"] = map[string]interface{}{
		// 运行端口
		"server_port": 8081,

		// JWT密钥, JWT配套有白名单动能不必担心秘钥泄露的问题
		"jwt_secret": "btRZ5QHXX9VjfYhfGGHdCTcWiwQ6WFJXq9ZCwdqZwzk2ZfhceM9K3V5UGKsYLd9m",
	}
}
