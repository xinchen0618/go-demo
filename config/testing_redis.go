// Package config 配置实现
package config

func init() {
	const env = "testing" // 测试环境配置
	if !EnvCheck(env) {
		return
	}

	if _, ok := configure[env]; !ok {
		configure[env] = map[string]any{}
	}
	for k, v := range map[string]any{
		/************ 配置项 START **************/

		// Redis DEMO
		"redis_host":          "127.0.0.1",
		"redis_port":          6379,
		"redis_auth":          "",
		"redis_index_cache":   0, // 缓存
		"redis_index_jwt":     1, // JWT
		"redis_index_storage": 2, // 存储
		"redis_index_queue":   3, // 消息队列

		/************ 配置项 END *****************/
	} {
		configure[env][k] = v
	}
}
