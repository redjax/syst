package encodeservice

import (
	"errors"
	"strings"
)

// EncodeService provides encoding operations.
type EncodeService struct{}

// NewEncodeService creates and returns a new EncodeService instance.
func NewEncodeService() *EncodeService {
	return &EncodeService{}
}

// Encode takes a method name and input string, then returns the encoded string.
// Returns error if method is unsupported.
func (s *EncodeService) Encode(method string, input string) (string, error) {
	method = strings.ToLower(method)

	switch method {
	case "base32", "b32":
		return EncodeStringBase32(input), nil
	case "base64", "b64":
		return EncodeStringBase64(input), nil
	case "hex":
		return EncodeStringHex(input), nil
	case "url", "urlencode":
		return EncodeStringUrl(input), nil
	case "utf8", "utf":
		return string(EncodeStringToUTF8Bytes(input)), nil
	case "utf8-hex", "utf-hex":
		return EncodeStringUTF8Hex(input), nil
	default:
		return "", errors.New("unknown encode method: " + method)
	}
}
