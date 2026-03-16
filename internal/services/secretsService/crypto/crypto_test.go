package crypto

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
)

func TestGenerateSHA256(t *testing.T) {
	result, err := GenerateSHA256(32)
	if err != nil {
		t.Fatalf("GenerateSHA256(32) error: %v", err)
	}
	// SHA-256 always produces 64 hex chars
	if len(result) != 64 {
		t.Errorf("GenerateSHA256(32) length = %d, want 64", len(result))
	}
	// Verify it's valid hex
	if _, err := hex.DecodeString(result); err != nil {
		t.Errorf("GenerateSHA256(32) produced invalid hex: %v", err)
	}
}

func TestGenerateSHA256_Uniqueness(t *testing.T) {
	a, _ := GenerateSHA256(32)
	b, _ := GenerateSHA256(32)
	if a == b {
		t.Error("two GenerateSHA256 calls produced identical output")
	}
}

func TestGenerateMD5(t *testing.T) {
	result, err := GenerateMD5(32)
	if err != nil {
		t.Fatalf("GenerateMD5(32) error: %v", err)
	}
	// MD5 always produces 32 hex chars
	if len(result) != 32 {
		t.Errorf("GenerateMD5(32) length = %d, want 32", len(result))
	}
	if _, err := hex.DecodeString(result); err != nil {
		t.Errorf("GenerateMD5(32) produced invalid hex: %v", err)
	}
}

func TestGenerateOpenSSLBase64(t *testing.T) {
	result, err := GenerateOpenSSLBase64(32)
	if err != nil {
		t.Fatalf("GenerateOpenSSLBase64(32) error: %v", err)
	}
	if result == "" {
		t.Error("GenerateOpenSSLBase64(32) returned empty string")
	}
	// Verify it's valid base64
	if _, err := base64.StdEncoding.DecodeString(result); err != nil {
		t.Errorf("GenerateOpenSSLBase64(32) produced invalid base64: %v", err)
	}
}

func TestGenerateOpenSSLHex(t *testing.T) {
	result, err := GenerateOpenSSLHex(32)
	if err != nil {
		t.Fatalf("GenerateOpenSSLHex(32) error: %v", err)
	}
	// 32 bytes -> 64 hex chars
	if len(result) != 64 {
		t.Errorf("GenerateOpenSSLHex(32) length = %d, want 64", len(result))
	}
	if _, err := hex.DecodeString(result); err != nil {
		t.Errorf("GenerateOpenSSLHex(32) produced invalid hex: %v", err)
	}
}

func TestGenerateOpenSSLSafe(t *testing.T) {
	result, err := GenerateOpenSSLSafe(32)
	if err != nil {
		t.Fatalf("GenerateOpenSSLSafe(32) error: %v", err)
	}
	if result == "" {
		t.Error("GenerateOpenSSLSafe(32) returned empty string")
	}
	// Verify no unsafe chars remain (+ / = should be replaced)
	if strings.ContainsAny(result, "+/=") {
		t.Errorf("GenerateOpenSSLSafe(32) contains unsafe chars: %q", result)
	}
}

func TestGenerateBase64(t *testing.T) {
	tests := []struct {
		name   string
		length int
		wantOk bool
	}{
		{"normal length", 32, true},
		{"small length", 1, true},
		{"zero length", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateBase64(tt.length)
			if err != nil {
				t.Fatalf("GenerateBase64(%d) error: %v", tt.length, err)
			}
			if tt.length == 0 {
				if result != "" {
					t.Errorf("GenerateBase64(0) = %q, want empty", result)
				}
				return
			}
			if result == "" {
				t.Errorf("GenerateBase64(%d) returned empty string", tt.length)
			}
			if _, err := base64.StdEncoding.DecodeString(result); err != nil {
				t.Errorf("GenerateBase64(%d) invalid base64: %v", tt.length, err)
			}
		})
	}
}

func TestEncodeStringBase64_Crypto(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello", "aGVsbG8="},
		{"", ""},
		{"test", "dGVzdA=="},
	}
	for _, tt := range tests {
		got := EncodeStringBase64(tt.input)
		if got != tt.want {
			t.Errorf("EncodeStringBase64(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
