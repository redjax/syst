package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
	"strings"
)

// GenerateOpenSSLBase64 generates base64 encoded random bytes
func GenerateOpenSSLBase64(length int) (string, error) {
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// GenerateOpenSSLHex generates random bytes and returns hex string
func GenerateOpenSSLHex(length int) (string, error) {
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateOpenSSLSafe generates a base64 string safe for env vars by replacing + / = chars
func GenerateOpenSSLSafe(length int) (string, error) {
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	s := base64.StdEncoding.EncodeToString(b)
	s = strings.NewReplacer("+", "A", "/", "B", "=", "C").Replace(s)
	return s, nil
}
