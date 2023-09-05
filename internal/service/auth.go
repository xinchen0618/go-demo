// Package service 内部应用业务原子级服务
//
//	需要公共使用的业务逻辑在这里实现.
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-demo/config"
	"go-demo/config/consts"
	"go-demo/config/di"

	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cast"
)

type auth struct{}

var Auth auth

// JWTLogin JWT 登录
//
//	先生成 JWT, 再记录 redis 白名单.
//	userType 为 JWT 登录用户类型, 集中在 consts/auth.go 中定义. id 为用户 id.
//	返回字符串为 JWT token.
func (auth) JWTLogin(userType string, id int64, userName string) (string, error) {
	// JWT登录
	loginTtl := 30 * 24 * time.Hour  // 登录有效时长
	claims := &jwt.RegisteredClaims{ // **这样赋值并不符合JWT定义中的声明, 如此处理仅是为了方便**
		Issuer:    userType, // 角色
		Subject:   userName, // 用户名
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(loginTtl)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        cast.ToString(id), // ID
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.GetString("jwt_secret")))
	if err != nil {
		di.Logger().Error(err.Error())
		return "", err
	}
	// redis 登录白名单
	tokenAtoms := strings.Split(tokenString, ".")
	payload, err := json.Marshal(claims)
	if err != nil {
		di.Logger().Error(err.Error())
		return "", err
	}
	key := fmt.Sprintf(consts.JWTLogin, userType, claims.ID, tokenAtoms[2])
	if err := di.JWTRedis().Set(context.Background(), key, payload, loginTtl).Err(); err != nil {
		di.Logger().Error(err.Error())
		return "", err
	}

	return tokenString, nil
}

// JWTLogout JWT 登出
//
//	从 redis 白名单删除.
//	userType 为 JWT 登录用户类型, 集中在 consts/auth.go 中定义. token 为 JWT token. id 为用户 id.
func (auth) JWTLogout(userType, token string, id int64) error {
	tokenAtoms := strings.Split(token, ".")
	key := fmt.Sprintf(consts.JWTLogin, userType, id, tokenAtoms[2])
	if err := di.JWTRedis().Del(context.Background(), key).Err(); err != nil {
		di.Logger().Error(err.Error())
		return err
	}

	return nil
}
