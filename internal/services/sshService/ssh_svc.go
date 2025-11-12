package sshservice

import (
	"errors"
	"fmt"
	"strings"

	"github.com/redjax/syst/internal/services/sshService/keygen"
)

type KeyGenOptions struct {
	FilePath  string
	Algorithm string
	Bits      int    // Only applies for RSA
	Password  string // Optional
	Comment   string // Optional
}

func GenerateKey(opts KeyGenOptions) (privKey, pubKey string, err error) {
	algo := strings.ToLower(opts.Algorithm)
	switch algo {
	case "rsa":
		bits := opts.Bits
		if bits == 0 {
			bits = 4096
		}
		if bits < 2048 {
			return "", "", errors.New("insecure RSA key size, must be at least 2048 bits")
		}
		return keygen.GenerateRSAKey(bits, opts.Comment, opts.Password)
	case "ed25519":
		return keygen.GenerateEd25519Key(opts.Comment, opts.Password)
	default:
		return "", "", fmt.Errorf("unsupported key algorithm: %s", opts.Algorithm)
	}
}
