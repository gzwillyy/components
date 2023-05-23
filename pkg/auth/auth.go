// Package auth 加密和比较密码字符串
package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// Encrypt 使用 bcrypt 加密纯文本.
func Encrypt(source string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(source), bcrypt.DefaultCost)
	return string(hashedBytes), err
}

// Compare 比较加密文本和明文是否相同.
func Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Sign 根据 secretID、secretKey、iss 和 aud 签发一个 jwt token.
func Sign(secretID string, secretKey string, iss, aud string) string {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Minute).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Add(0).Unix(),
		"aud": aud,
		"iss": iss,
	}

	// create a new token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = secretID

	// Sign the token with the specified secret.
	tokenString, _ := token.SignedString([]byte(secretKey))

	return tokenString
}
