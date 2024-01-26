// Package consts 常量定义
package consts

// 鉴权
const (
	JWTLogin = "%s:%v:jwt:%s" // JWT 登录凭证 <userType>:<userID>:jwt:<md5(jwtToken)>
)

// 安全
const (
	SubmitLimit = "submit:limit:%s" // 提交频率限制, submit:limit:<md5(id|ip&&agent+method+path)>
)
