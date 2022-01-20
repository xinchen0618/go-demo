package consts

const (
	JwtUserLogin = "jwt:user:%v:%s"           // jwt用户登录凭证 jwt:user:<userId>:<JwtSignature>
	SubmitLimit  = "submit:limit:id:%v:%s:%s" // 提交频率限制, submit:limit:id:<id>:<method>:<path>

	CacheResource = "cache:resource:%s:%d" // 资源缓存 cache:resource:<table_name>:<primary_id>
	CacheUsers    = "cache:users:%s"       // 用户列表缓存 cache:users:<md5(queries)>
)
