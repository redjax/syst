package crypto

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
)

func GenerateMD5(length int) (string, error) {
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	sum := md5.Sum(b)
	return hex.EncodeToString(sum[:]), nil
}
