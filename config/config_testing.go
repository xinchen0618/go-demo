package config

// 测试环境配置
func init() {
	configure["testing"] = map[string]interface{}{
		// 运行端口
		"server_port": 8080,

		// JWT密钥, JWT配套有白名单动能不必担心秘钥泄露的问题
		"jwt_secret": "Xx4KJQ2AguFL5gWurcRJvVfDC5a2itLi53vFJN9wthYkrxtQbdeRDkWTHzAjnn5n",
	}
}