package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type BCryptHasher struct{ Cost int }

func (h BCryptHasher) Hash(password string) (string, error) {
	cost := h.Cost
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	b, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(b), err
}

func (h BCryptHasher) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

type Argon2Hasher struct{}

func (h Argon2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("argon2id$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash)), nil
}

func (h Argon2Hasher) Compare(encoded, password string) error {
	parts := strings.Split(encoded, "$")
	if len(parts) != 3 || parts[0] != "argon2id" {
		return fmt.Errorf("invalid argon2 hash format")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return err
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return err
	}
	actual := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	if subtle.ConstantTimeCompare(actual, expected) != 1 {
		return fmt.Errorf("password mismatch")
	}
	return nil
}
