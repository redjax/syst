package secretsservice

import (
	"encoding/hex"
	"testing"
)

func TestNewSecretsService(t *testing.T) {
	svc := NewSecretsService()
	if svc == nil {
		t.Fatal("NewSecretsService() returned nil")
	}
}

func TestSecretsService_GenerateSecret(t *testing.T) {
	svc := NewSecretsService()

	tests := []struct {
		name    string
		method  string
		length  int
		wantErr bool
	}{
		{"sha256", "sha256", 32, false},
		{"sha256 uppercase", "SHA256", 32, false},
		{"md5", "md5", 32, false},
		{"openssl", "openssl", 32, false},
		{"openssl-hex", "openssl-hex", 32, false},
		{"openssl-safe", "openssl-safe", 32, false},
		{"unknown method", "rot13", 32, true},
		{"empty method", "", 32, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.GenerateSecret(tt.method, tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSecret(%q, %d) error = %v, wantErr %v", tt.method, tt.length, err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" {
				t.Errorf("GenerateSecret(%q, %d) returned empty string", tt.method, tt.length)
			}
		})
	}
}

func TestSecretsService_SHA256Format(t *testing.T) {
	svc := NewSecretsService()
	result, err := svc.GenerateSecret("sha256", 32)
	if err != nil {
		t.Fatalf("GenerateSecret sha256 error: %v", err)
	}
	// SHA-256 outputs 64 hex chars
	if len(result) != 64 {
		t.Errorf("sha256 result length = %d, want 64", len(result))
	}
	if _, err := hex.DecodeString(result); err != nil {
		t.Errorf("sha256 result is not valid hex: %v", err)
	}
}

func TestSecretsService_Uniqueness(t *testing.T) {
	svc := NewSecretsService()
	methods := []string{"sha256", "md5", "openssl", "openssl-hex", "openssl-safe"}
	for _, method := range methods {
		a, err := svc.GenerateSecret(method, 32)
		if err != nil {
			t.Fatalf("GenerateSecret(%q) error: %v", method, err)
		}
		b, err := svc.GenerateSecret(method, 32)
		if err != nil {
			t.Fatalf("GenerateSecret(%q) error: %v", method, err)
		}
		if a == b {
			t.Errorf("GenerateSecret(%q) produced identical results: %q", method, a)
		}
	}
}
