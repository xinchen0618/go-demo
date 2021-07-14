package config

// Redis key统一在此定义避免冲突
const (
	RedisUser     = "user:%d"      // 用户信息
	RedisRoleAuth = "role:%d:auth" // 角色权限
)
