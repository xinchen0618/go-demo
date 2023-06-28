package consts

// 鉴权
const (
	JWTLogin = "jwt:%s:%v:%s" // JWT登录凭证 jwt:<userType>:<userId>:<jwtSignature>
)

// 安全
const (
	SubmitLimit = "submit:limit:%s" // 提交频率限制, submit:limit:<md5(id|ip&&agent+method+path)>
)
