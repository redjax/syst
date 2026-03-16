package sshservice

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateKey_Ed25519(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_ed25519")

	privPath, pubPath, err := GenerateKey(KeyGenOptions{
		FilePath:  keyPath,
		Algorithm: "ed25519",
		Comment:   "test@test",
		Force:     true,
	})
	if err != nil {
		t.Fatalf("GenerateKey ed25519 error: %v", err)
	}

	if privPath != keyPath {
		t.Errorf("privPath = %q, want %q", privPath, keyPath)
	}
	if pubPath != keyPath+".pub" {
		t.Errorf("pubPath = %q, want %q", pubPath, keyPath+".pub")
	}

	// Verify files exist
	if _, err := os.Stat(privPath); err != nil {
		t.Errorf("private key file not found: %v", err)
	}
	if _, err := os.Stat(pubPath); err != nil {
		t.Errorf("public key file not found: %v", err)
	}

	// Verify private key content looks like a PEM key
	privContent, _ := os.ReadFile(privPath)
	if !strings.Contains(string(privContent), "PRIVATE KEY") {
		t.Error("private key doesn't look like a PEM key")
	}

	// Verify public key content looks like an SSH public key
	pubContent, _ := os.ReadFile(pubPath)
	if !strings.HasPrefix(string(pubContent), "ssh-ed25519 ") {
		t.Error("public key doesn't start with ssh-ed25519")
	}

	// Verify permissions (owner read/write only)
	info, _ := os.Stat(privPath)
	if info.Mode().Perm() != 0600 {
		t.Errorf("private key permissions = %o, want 0600", info.Mode().Perm())
	}
}

func TestGenerateKey_RSA(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_rsa")

	privPath, pubPath, err := GenerateKey(KeyGenOptions{
		FilePath:  keyPath,
		Algorithm: "rsa",
		Bits:      2048, // minimum secure size
		Comment:   "test@rsa",
		Force:     true,
	})
	if err != nil {
		t.Fatalf("GenerateKey rsa error: %v", err)
	}

	// Verify files exist
	if _, err := os.Stat(privPath); err != nil {
		t.Errorf("private key file not found: %v", err)
	}
	if _, err := os.Stat(pubPath); err != nil {
		t.Errorf("public key file not found: %v", err)
	}

	pubContent, _ := os.ReadFile(pubPath)
	if !strings.HasPrefix(string(pubContent), "ssh-rsa ") {
		t.Error("public key doesn't start with ssh-rsa")
	}
}

func TestGenerateKey_InsecureRSA(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_weak_rsa")

	_, _, err := GenerateKey(KeyGenOptions{
		FilePath:  keyPath,
		Algorithm: "rsa",
		Bits:      1024, // too small
		Force:     true,
	})
	if err == nil {
		t.Error("GenerateKey with 1024 bit RSA should return error")
	}
}

func TestGenerateKey_UnsupportedAlgorithm(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_bad")

	_, _, err := GenerateKey(KeyGenOptions{
		FilePath:  keyPath,
		Algorithm: "dsa",
		Force:     true,
	})
	if err == nil {
		t.Error("GenerateKey with unsupported algorithm should return error")
	}
}

func TestGenerateKey_NoOverwriteWithoutForce(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "existing_key")

	// Create an existing file
	os.WriteFile(keyPath, []byte("existing"), 0600)

	// Without Force, it would normally prompt (which we can't do in tests)
	// But since stdin is not a terminal, promptOverwrite should error or return false
	_, _, err := GenerateKey(KeyGenOptions{
		FilePath:  keyPath,
		Algorithm: "ed25519",
		Force:     false,
	})
	// Should get an error since we can't interactively confirm
	if err == nil {
		t.Error("GenerateKey without Force on existing file should prompt/error")
	}
}

func TestEnsureOverwriteSafety_Force(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "existing")
	os.WriteFile(filePath, []byte("data"), 0600)

	err := ensureOverwriteSafety(filePath, true)
	if err != nil {
		t.Errorf("ensureOverwriteSafety with force=true should not error: %v", err)
	}
}

func TestEnsureOverwriteSafety_NoFile(t *testing.T) {
	err := ensureOverwriteSafety("/nonexistent/path/file", false)
	if err != nil {
		t.Errorf("ensureOverwriteSafety on nonexistent file should not error: %v", err)
	}
}
