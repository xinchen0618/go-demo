// Package config 配置实现
package config

import "github.com/samber/lo"

func init() {
	const env = "testing" // 测试环境配置

	if !lo.Contains([]string{RuntimeEnv(), "common", "local"}, env) {
		return
	}
	if _, ok := configure[env]; !ok {
		configure[env] = map[string]any{}
	}
	for k, v := range map[string]any{
		/************ 配置项 START **************/

		// 运行端口
		"server_port": 8080,

		// JWT 密钥, JWT 配套有白名单功能不必担心秘钥泄露的问题
		"jwt_secret": "Xx4KJQ2AguFL5gWurcRJvVfDC5a2itLi53vFJN9wthYkrxtQbdeRDkWTHzAjnn5n",

		/************ 配置项 END *****************/
	} {
		configure[env][k] = v
	}
}
