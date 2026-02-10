package token

import (
	"crypto/sha256"
	"encoding/hex"
)

func NewOpaqueTokenFromPlain(plain string) (string, string, error) {
	h := sha256.Sum256([]byte(plain))
	return plain, hex.EncodeToString(h[:]), nil
}
