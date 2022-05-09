package consts

// 鉴权
const (
	JwtLogin = "jwt:%s:%v:%s" // JWT登录凭证 jwt:<userType>:<userId>:<JwtSignature>
)

// 安全
const (
	SubmitLimit = "submit:limit:%s" // 提交频率限制, submit:limit:<md5(id|ip&&agent+method+path)>
)

// 缓存
const (
	CacheResource = "cache:resource:%s:%d" // 资源缓存 cache:resource:<table_name>:<primary_id>
)
