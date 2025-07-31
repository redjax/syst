package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

// GenerateBase64 generates a secure random byte slice of the given length,
// and returns it encoded in standard Base64.
func GenerateBase64(length int) (string, error) {
	if length <= 0 {
		return "", nil // or optionally return an error for invalid length
	}

	b := make([]byte, length)
	n, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", err
	}
	if n != length {
		return "", io.ErrUnexpectedEOF
	}

	encoded := base64.StdEncoding.EncodeToString(b)
	return encoded, nil
}

// EncodeStringBase64 encodes the provided string input into a Base64 encoded string.
func EncodeStringBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}
