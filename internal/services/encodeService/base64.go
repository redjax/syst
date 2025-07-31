package encodeservice

import (
	"encoding/base64"
)

// EncodeStringBase64 encodes the provided string input into a Base64 encoded string.
func EncodeStringBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}
