package config

import (
	"os"
)

// configure["common": interface{}, "<runtimeEnv>": interface{}]
var configure = map[string]interface{}{}

// GetRuntimeEnv 获取运行时环境
//	@return string
func GetRuntimeEnv() string {
	runtimeEnv := os.Getenv("RUNTIME_ENV")
	if "" == runtimeEnv { // 默认为生产环境
		runtimeEnv = "prod"
	}

	return runtimeEnv
}

func Get(key string) interface{} {
	runtimeEnv := GetRuntimeEnv()

	if envConfigure, ok := configure[runtimeEnv]; ok { // 环境配置
		if value, ok := envConfigure.(map[string]interface{})[key]; ok {
			return value
		}
	}
	if commonConfigure, ok := configure["common"]; ok { // 公共配置
		if value, ok := commonConfigure.(map[string]interface{})[key]; ok {
			return value
		}
	}

	return nil
}

func GetInt(key string) int {
	if value := Get(key); value != nil {
		return value.(int)
	} else {
		return 0
	}
}

func GetString(key string) string {
	if value := Get(key); value != nil {
		return value.(string)
	} else {
		return ""
	}
}

func GetStringSlice(key string) []string {
	if value := Get(key); value != nil {
		return value.([]string)
	} else {
		return []string{}
	}
}

func GetIntSlice(key string) []int {
	if value := Get(key); value != nil {
		return value.([]int)
	} else {
		return []int{}
	}
}
