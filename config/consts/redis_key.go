package consts

// 鉴权
const (
	JwtLogin = "jwt:%s:%v:%s" // JWT登录凭证 jwt:<userType>:<userId>:<JwtSignature>
)

// 安全
const (
	SubmitLimit = "submit:limit:id:%v:%s:%s" // 提交频率限制, submit:limit:id:<id>:<method>:<path>
)

// 缓存
const (
	CacheResource = "cache:resource:%s:%d" // 资源缓存 cache:resource:<table_name>:<primary_id>
	CacheUsers    = "cache:users:%s"       // 用户列表缓存 cache:users:<md5(queries)>
)
