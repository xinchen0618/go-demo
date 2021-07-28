package config

// Redis key统一在此定义避免冲突
const (
	RedisResourceInfo = "resource:%s:%d" // 资源缓存
	RedisRoleAuth     = "role:%d:auth"   // 角色权限
)
