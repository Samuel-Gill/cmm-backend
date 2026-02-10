package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	secret []byte
}

func NewManager(secret string) Manager { return Manager{secret: []byte(secret)} }

func (m Manager) CreateAccessToken(userID string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(ttl).Unix(),
		"iat": time.Now().Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(m.secret)
}

func (m Manager) ParseAccessToken(tokenStr string) (string, error) {
	t, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil || !t.Valid {
		return "", err
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return "", jwt.ErrTokenInvalidClaims
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", jwt.ErrTokenInvalidClaims
	}
	return sub, nil
}

func NewOpaqueToken() (plain string, hash string, err error) {
	buf := make([]byte, 32)
	_, err = rand.Read(buf)
	if err != nil {
		return "", "", err
	}
	plain = base64.RawURLEncoding.EncodeToString(buf)
	h := sha256.Sum256([]byte(plain))
	hash = hex.EncodeToString(h[:])
	return plain, hash, nil
}
