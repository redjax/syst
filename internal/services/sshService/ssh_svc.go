package sshservice

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/redjax/syst/internal/services/sshService/keygen"
	"github.com/redjax/syst/internal/utils/path"
)

var ErrUserAborted = errors.New("user aborted")

type KeyGenOptions struct {
	FilePath  string
	Algorithm string
	Bits      int    // Only applies for RSA
	Password  string // Optional
	Comment   string // Optional
	Force     bool
}

// GenerateKey handles the entire SSH key generation lifecycle:
//   - Path expansion (~ → /home/user)
//   - Directory creation
//   - Existence checks + confirmation prompts
//   - Dispatch to keygen functions
//   - Writing key files
func GenerateKey(opts KeyGenOptions) (privKeyPath, pubKeyPath string, err error) {
	// Set default output path based on algorithm if not provided
	if opts.FilePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", "", fmt.Errorf("failed to get home directory: %w", err)
		}

		switch strings.ToLower(opts.Algorithm) {
		case "rsa":
			opts.FilePath = filepath.Join(home, ".ssh", "id_rsa")
		case "ed25519":
			opts.FilePath = filepath.Join(home, ".ssh", "id_ed25519")
		default:
			return "", "", fmt.Errorf("unsupported algorithm: %s", opts.Algorithm)
		}
	}

	// Expand ~ and other shell-like paths
	expanded, err := path.ExpandPath(opts.FilePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve path: %w", err)
	}
	opts.FilePath = expanded

	// Ensure parent directory exists
	dir := filepath.Dir(opts.FilePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Safety checks before overwriting existing keys
	if err := ensureOverwriteSafety(opts.FilePath, opts.Force); err != nil {
		return "", "", err
	}
	if err := ensureOverwriteSafety(opts.FilePath+".pub", opts.Force); err != nil {
		return "", "", err
	}

	// Generate keypair
	var privKey, pubKey string
	switch strings.ToLower(opts.Algorithm) {
	case "rsa":
		bits := opts.Bits
		if bits == 0 {
			bits = 4096
		}
		if bits < 2048 {
			return "", "", errors.New("insecure RSA key size, must be at least 2048 bits")
		}
		privKey, pubKey, err = keygen.GenerateRSAKey(bits, opts.Comment, opts.Password)

	case "ed25519":
		privKey, pubKey, err = keygen.GenerateEd25519Key(opts.Comment, opts.Password)

	default:
		return "", "", fmt.Errorf("unsupported algorithm: %s", opts.Algorithm)
	}
	if err != nil {
		return "", "", fmt.Errorf("key generation failed: %w", err)
	}

	// Write keys to disk
	if err := os.WriteFile(opts.FilePath, []byte(privKey), 0600); err != nil {
		return "", "", fmt.Errorf("failed to save private key: %w", err)
	}

	pubFile := opts.FilePath + ".pub"
	if err := os.WriteFile(pubFile, []byte(pubKey), 0644); err != nil {
		return "", "", fmt.Errorf("failed to save public key: %w", err)
	}

	return opts.FilePath, pubFile, nil
}

// ensureOverwriteSafety checks whether a file exists and prompts for overwrite if needed.
func ensureOverwriteSafety(path string, force bool) error {
	if _, err := os.Stat(path); err == nil {
		if force {
			return nil
		}
		ok, err := promptOverwrite(path)
		if err != nil {
			return err
		}
		if !ok {
			// Return the sentinel error instead of a generic error
			return ErrUserAborted
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check key file: %w", err)
	}
	return nil
}

func promptOverwrite(path string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("File %s already exists. Overwrite? [y/N]: ", path)
	resp, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}
	resp = strings.TrimSpace(strings.ToLower(resp))
	return resp == "y" || resp == "yes", nil
}
