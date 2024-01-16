// Package config 配置实现
package config

import (
	"os"

	"github.com/spf13/cast"
	"go.uber.org/zap"
)

// configure{"common": map[string]any, "<runtimeEnv>": map[string]any}
var configure = map[string]map[string]any{}

// RuntimeEnv 获取运行时环境
func RuntimeEnv() string {
	runtimeEnv := os.Getenv("RUNTIME_ENV")
	if runtimeEnv == "" { // 默认为生产环境
		runtimeEnv = "prod"
	}

	return runtimeEnv
}

func get(key string) any {
	runtimeEnv := RuntimeEnv()

	if envConfigure, ok := configure[runtimeEnv]; ok { // 环境配置
		if value, ok := envConfigure[key]; ok {
			return value
		}
	}
	if commonConfigure, ok := configure["common"]; ok { // 公共配置
		if value, ok := commonConfigure[key]; ok {
			return value
		}
	}

	return nil
}

func GetInt(key string) int {
	value, err := cast.ToIntE(get(key))
	if err != nil {
		zap.L().Error(err.Error())
	}
	return value
}

func GetString(key string) string {
	value, err := cast.ToStringE(get(key))
	if err != nil {
		zap.L().Error(err.Error())
	}
	return value
}

func GetBool(key string) bool {
	value, err := cast.ToBoolE(get(key))
	if err != nil {
		zap.L().Error(err.Error())
	}
	return value
}

func GetStringSlice(key string) []string {
	value, err := cast.ToStringSliceE(get(key))
	if err != nil {
		zap.L().Error(err.Error())
	}
	return value
}

func GetIntSlice(key string) []int {
	value, err := cast.ToIntSliceE(get(key))
	if err != nil {
		zap.L().Error(err.Error())
	}
	return value
}

func GetStringMapString(key string) map[string]string {
	value, err := cast.ToStringMapStringE(get(key))
	if err != nil {
		zap.L().Error(err.Error())
	}
	return value
}
