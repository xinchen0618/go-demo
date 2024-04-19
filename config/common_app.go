// Package config 配置实现
package config

func init() {
	const env = "common" // 公共配置
	if !EnvCheck(env) {
		return
	}

	if _, ok := configure[env]; !ok {
		configure[env] = map[string]any{}
	}
	for k, v := range map[string]any{
		/************ 配置项 START ****************/

		// ERROR 日志路径
		"error_log": "/var/log/golang_app.log",
		// ERROR 日志级别
		"error_log_level": "Debug", // Debug, Info, Warn, Error
		// 日志编码格式
		"error_log_encoder": "JSON", // Console, JSON
		// 彩色输出 ERROR 日志
		"error_log_colorful": false, // Console 格式有效

		// 公共 Goroutine 池大小
		"worker_pool": 409600,

		// 限流 QPS
		"qps_limit": 40000,

		// 超时控制, 秒
		"timeout": 30,

		/************ 配置项 END ******************/
	} {
		configure[env][k] = v
	}
}
