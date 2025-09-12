package middleware

import (
	"errors"
	"fmt"
	"github.com/sporfi/sport-common/pkg/constants"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims 自定义JWT声明
type CustomClaims struct {
	UserID uint64 `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint64, email string, expire int64, secret string) (string, error) {
	// 设置过期时间
	expireTime := time.Now().Add(time.Duration(expire) * time.Second)

	// 创建声明
	claims := CustomClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "sport-user",
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(constants.SigningMethodHS256, claims)

	// 签名令牌
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString, secret string) (*CustomClaims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法是否与预期一致
		if token.Method != constants.SigningMethodHS256 {
			return nil, fmt.Errorf("Unsupported signature algorithm: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证令牌并获取声明
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
