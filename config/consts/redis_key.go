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
	CacheDb = "cache:db:%s:%d" // DB缓存 cache:db:<table_name>:<primary_id>
)
