package keygen

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ssh"
)

func GenerateEd25519Key(comment string, password string) (privKeyPEM string, pubKey string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate ed25519 key: %w", err)
	}

	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", "", fmt.Errorf("unable to marshal ed25519 private key: %w", err)
	}

	var block *pem.Block
	if password != "" {
		block, err = x509.EncryptPEMBlock(rand.Reader, "ENCRYPTED PRIVATE KEY", privDER, []byte(password), x509.PEMCipherAES256)
		if err != nil {
			return "", "", fmt.Errorf("failed to encrypt private key: %w", err)
		}
	} else {
		block = &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privDER,
		}
	}

	privKeyPEM = string(pem.EncodeToMemory(block))

	pubKeySSH, err := ssh.NewPublicKey(pub)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate public key: %w", err)
	}
	pubKey = string(ssh.MarshalAuthorizedKey(pubKeySSH))
	if comment != "" {
		pubKey = pubKey[:len(pubKey)-1] + " " + comment + "\n"
	}

	return privKeyPEM, pubKey, nil
}
