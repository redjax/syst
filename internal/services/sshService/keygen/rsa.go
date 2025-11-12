package keygen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"golang.org/x/crypto/ssh"
)

// GenerateRSAKey generates an RSA keypair and returns the PEM-encoded private key
// and SSH public key. If password is non-empty, the private key is encrypted using AES-256.
func GenerateRSAKey(bits int, comment string, password string) (privKeyPEM string, pubKey string, err error) {
	if bits < 2048 {
		return "", "", errors.New("insecure RSA key size, use 2048 or higher")
	}

	// Generate RSA private key
	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Marshal private key to DER format
	privDER := x509.MarshalPKCS1PrivateKey(privKey)

	var block *pem.Block
	if password != "" {
		// Use modern x509.EncryptPEMBlock instead of deprecated pem.EncryptPEMBlock
		block, err = x509.EncryptPEMBlock(rand.Reader, "RSA PRIVATE KEY", privDER, []byte(password), x509.PEMCipherAES256)
		if err != nil {
			return "", "", fmt.Errorf("failed to encrypt private key: %w", err)
		}
	} else {
		block = &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privDER,
		}
	}

	// Encode PEM
	privKeyPEM = string(pem.EncodeToMemory(block))

	// Generate SSH public key
	pubKeySSH, err := ssh.NewPublicKey(&privKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate public key: %w", err)
	}
	pubKey = string(ssh.MarshalAuthorizedKey(pubKeySSH))
	if comment != "" {
		pubKey = pubKey[:len(pubKey)-1] + " " + comment + "\n"
	}

	return privKeyPEM, pubKey, nil
}
