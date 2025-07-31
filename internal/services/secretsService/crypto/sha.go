package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
)

func GenerateSHA256(length int) (string, error) {
	return generateHash(sha256.New, length)
}

func generateHash(hashFunc func() hash.Hash, length int) (string, error) {
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	h := hashFunc()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil)), nil
}
