package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-demo/config"
	"go-demo/config/consts"
	"go-demo/config/di"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

type auth struct{}

var Auth auth

// JwtLogin JWT登录
//	先生成JWT, 再记录redis白名单
//  @receiver auth
//  @param userType string JWT登录用户类型
//  @param id int64 用户id
//  @param userName string
//  @return string jwt token
//  @return error
func (auth) JwtLogin(userType string, id int64, userName string) (string, error) {
	// JWT登录
	loginTtl := 30 * 24 * time.Hour // 登录有效时长
	claims := &jwt.StandardClaims{
		Audience:  userName,
		ExpiresAt: time.Now().Add(loginTtl).Unix(),
		Id:        cast.ToString(id),
		IssuedAt:  time.Now().Unix(),
		Issuer:    userType + " login",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.GetString("jwt_secret")))
	if err != nil {
		zap.L().Error(err.Error())
		return "", err
	}
	// redis登录白名单
	tokenAtoms := strings.Split(tokenString, ".")
	payload, err := json.Marshal(claims)
	if err != nil {
		zap.L().Error(err.Error())
		return "", err
	}
	key := fmt.Sprintf(consts.JwtLogin, userType, claims.Id, tokenAtoms[2])
	if err := di.JwtRedis().Set(context.Background(), key, payload, loginTtl).Err(); err != nil {
		zap.L().Error(err.Error())
		return "", err
	}

	return tokenString, nil
}

// JwtCheck JWT登录校验
//  @receiver auth
//  @param userType string
//  @param token string
//  @return int64 用户id, 0表示校验不通过
//  @return error
func (auth) JwtCheck(userType string, token string) (int64, error) {
	// JWT解析
	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return []byte(config.GetString("jwt_secret")), nil
	})
	if err != nil { // token无效
		return 0, nil
	}
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok || !jwtToken.Valid { // token秘钥/时间等校验未通过
		return 0, nil
	}

	// 白名单
	tokenAtoms := strings.Split(token, ".")
	key := fmt.Sprintf(consts.JwtLogin, userType, claims["jti"], tokenAtoms[2])
	if err := di.JwtRedis().Get(context.Background(), key).Err(); err != nil {
		if err != redis.Nil { // redis服务异常
			zap.L().Error(err.Error())
			return 0, err
		}
		// 不在白名单内
		return 0, nil
	}

	return cast.ToInt64(claims["jti"]), nil
}

// JwtLogout JWT登出
//	从redis白名单删除
//  @receiver auth
//  @param userType string
//  @param token string
//  @param id int64
//  @return error
func (auth) JwtLogout(userType, token string, id int64) error {
	tokenAtoms := strings.Split(token, ".")
	key := fmt.Sprintf(consts.JwtLogin, userType, id, tokenAtoms[2])
	if err := di.JwtRedis().Del(context.Background(), key).Err(); err != nil {
		zap.L().Error(err.Error())
		return err
	}

	return nil
}
