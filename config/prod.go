package config

// 生产环境配置
func init() {
	if GetRuntimeEnv() != "prod" {
		return
	}
	if _, ok := configure["prod"]; !ok {
		configure["prod"] = map[string]any{}
	}

	for k, v := range map[string]any{
		/* 生产环境配置start */

		// 运行端口
		"server_port": 8090,

		// JWT密钥, JWT配套有白名单功能不必担心秘钥泄露的问题
		"jwt_secret": "btRZ5QHXX9VjfYhfGGHdCTcWiwQ6WFJXq9ZCwdqZwzk2ZfhceM9K3V5UGKsYLd9m",

		/* 生产环境配置end */
	} {
		configure["prod"][k] = v
	}
}
