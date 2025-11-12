package crypto

import (
	"crypto/md5" // #nosec G501 - MD5 used for non-security purposes (checksums, legacy compatibility)
	"crypto/rand"
	"encoding/hex"
	"io"
)

// GenerateMD5 generates a random MD5 hash.
// Note: MD5 is not cryptographically secure. Only use for non-security purposes
// like checksums, cache keys, or legacy system compatibility.
func GenerateMD5(length int) (string, error) {
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	sum := md5.Sum(b) // #nosec G401 - MD5 used for non-security purposes
	return hex.EncodeToString(sum[:]), nil
}
