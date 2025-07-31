package secretsservice

import (
	"errors"
	"strings"

	"github.com/redjax/syst/internal/services/secretsService/crypto"
)

type SecretsService struct {
	// add config fields if needed in future
}

// Constructor
func NewSecretsService() *SecretsService {
	return &SecretsService{}
}

// GenerateSecret takes a method and length, and delegates to appropriate implementation
func (s *SecretsService) GenerateSecret(method string, length int) (string, error) {
	method = strings.ToLower(method)
	switch method {
	case "sha256":
		return crypto.GenerateSHA256(length)
	case "md5":
		return crypto.GenerateMD5(length)
	case "openssl":
		return crypto.GenerateOpenSSLBase64(length)
	case "openssl-hex":
		return crypto.GenerateOpenSSLHex(length)
	case "openssl-safe":
		return crypto.GenerateOpenSSLSafe(length)
	default:
		return "", errors.New("unknown secret generation method: " + method)
	}
}
