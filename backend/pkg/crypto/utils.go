package crypto

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateToken() (string, error) {
	bytes := make([]byte, 32) // 256 бит
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(bytes), nil
}