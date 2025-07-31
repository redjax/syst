// internal/services/encodeservice/hex.go
package encodeservice

import (
	"encoding/hex"
)

// EncodeStringHex encodes the provided string input into a hexadecimal encoded string.
func EncodeStringHex(input string) string {
	return hex.EncodeToString([]byte(input))
}
