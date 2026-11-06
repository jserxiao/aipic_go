package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret string

// SetJWTSecret 设置 JWT 密钥
func SetJWTSecret(secret string) {
	jwtSecret = secret
}

// TokenType token 类型
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Claims JWT 声明
type Claims struct {
	UserID string    `json:"user_id"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 access token
func GenerateToken(userID string, expireHours int) (string, error) {
	claims := Claims{
		UserID: userID,
		Type:   TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "aipic-go",
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// ParseToken 解析 access token（只接受 access 类型）
func ParseToken(tokenString string) (*Claims, error) {
	claims, err := parseTokenClaims(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Type != TokenTypeAccess {
		return nil, errors.New("invalid token type: expected access token")
	}
	return claims, nil
}

// GenerateRefreshToken 生成 refresh token
func GenerateRefreshToken(userID string, expireHours int) (string, error) {
	claims := Claims{
		UserID: userID,
		Type:   TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "aipic-go",
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// ParseRefreshToken 解析 refresh token（只接受 refresh 类型）
func ParseRefreshToken(tokenString string) (*Claims, error) {
	claims, err := parseTokenClaims(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Type != TokenTypeRefresh {
		return nil, errors.New("invalid token type: expected refresh token")
	}
	return claims, nil
}

// parseTokenClaims 内部解析，不校验类型
func parseTokenClaims(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 确保使用预期的签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
