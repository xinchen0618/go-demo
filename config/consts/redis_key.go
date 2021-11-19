package consts

const (
	CacheResource = "cache:resource:%s:%d" // 资源缓存 cache:resource:<table_name>:<primary_id>
	CacheUsers    = "cache:users:%s"       // 用户列表缓存 cache:users:<md5(queries)>
)
