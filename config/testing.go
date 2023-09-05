// Package config 配置实现
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
		/************ 测试环境配置 START **************/

		// 运行端口
		"server_port": 8080,

		// JWT 密钥, JWT 配套有白名单功能不必担心秘钥泄露的问题
		"jwt_secret": "Xx4KJQ2AguFL5gWurcRJvVfDC5a2itLi53vFJN9wthYkrxtQbdeRDkWTHzAjnn5n",

		/************ 测试环境配置 END *****************/
	} {
		configure["testing"][k] = v
	}
}
