package encodeservice

import (
	"encoding/base32"
)

// EncodeStringBase32 encodes the input string into Base32 string.
func EncodeStringBase32(input string) string {
	return base32.StdEncoding.EncodeToString([]byte(input))
}
