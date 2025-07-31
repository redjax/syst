package encodeservice

import "encoding/hex"

func EncodeStringToUTF8Bytes(input string) []byte {
	// Returns utf-8 encoded string
	return []byte(input)
}

// EncodeStringUTF8Hex returns hex string of UTF-8 bytes of input string
func EncodeStringUTF8Hex(input string) string {
	return hex.EncodeToString([]byte(input))
}
