package token

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

const (
	tokenJwtSecret     = "LqlH101211132414"
	tokenValidDuration = 10 //token有效时间，单位小时
	TokenIssuer        = "lq-chat"
)

type Claims struct {
	Uid int64 `json:"uid"`
	jwt.StandardClaims
}


func GenerateForeverToken(uid int64) (string,error){
	nowTime := time.Now()
	expireTime := nowTime.Add(20 * 365 * 24 * time.Hour)
	claims := Claims{
		uid,
		jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    TokenIssuer,
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString([]byte(tokenJwtSecret))
	return token, err
}

func GenerateToken(uid int64) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(tokenValidDuration * time.Hour)
	claims := Claims{
		uid,
		jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    TokenIssuer,
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString([]byte(tokenJwtSecret))
	return token, err
}

func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenJwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if tokenClaims == nil {
		return nil, errors.New("tokenClaims is nil")
	}
	if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
		return claims, nil
	}
	return nil, errors.New("parse to *Claims fail or tokenClaims is invalid")
}
