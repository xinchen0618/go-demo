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

		// DB DEMO
		"mysql_host":           "127.0.0.1",
		"mysql_port":           3306,
		"mysql_username":       "root",
		"mysql_password":       "cx654321",
		"mysql_dbname":         "test",
		"mysql_charset":        "utf8mb4",
		"mysql_max_open_conns": 140,
		"mysql_max_idle_conns": 30,

		/************ 配置项 END *****************/
	} {
		configure[env][k] = v
	}
}
