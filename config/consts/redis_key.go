package consts

const (
	JwtUserLogin = "jwt:%v:%s" // jwt用户登录凭证 jwt:<userId>:<JwtSignature>

	CacheResource = "cache:resource:%s:%d" // 资源缓存 cache:resource:<table_name>:<primary_id>
	CacheUsers    = "cache:users:%s"       // 用户列表缓存 cache:users:<md5(queries)>
)
