package config

import (
	"os"

	"github.com/spf13/cast"
	"go.uber.org/zap"
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
	value, err := cast.ToIntE(Get(key))
	if err != nil {
		zap.L().Error(err.Error())
	}
	return value
}

func GetString(key string) string {
	value, err := cast.ToStringE(Get(key))
	if err != nil {
		zap.L().Error(err.Error())
	}
	return value
}

func GetStringSlice(key string) []string {
	value, err := cast.ToStringSliceE(Get(key))
	if err != nil {
		zap.L().Error(err.Error())
	}
	return value
}

func GetIntSlice(key string) []int {
	value, err := cast.ToIntSliceE(Get(key))
	if err != nil {
		zap.L().Error(err.Error())
	}
	return value
}
