package encodeservice

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"testing"
)

func TestEncodeStringBase64(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello", base64.StdEncoding.EncodeToString([]byte("hello"))},
		{"", ""},
		{"hello world!", base64.StdEncoding.EncodeToString([]byte("hello world!"))},
		{"特殊字符", base64.StdEncoding.EncodeToString([]byte("特殊字符"))},
	}
	for _, tt := range tests {
		got := EncodeStringBase64(tt.input)
		if got != tt.want {
			t.Errorf("EncodeStringBase64(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEncodeStringBase32(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello", base32.StdEncoding.EncodeToString([]byte("hello"))},
		{"", ""},
		{"test123", base32.StdEncoding.EncodeToString([]byte("test123"))},
	}
	for _, tt := range tests {
		got := EncodeStringBase32(tt.input)
		if got != tt.want {
			t.Errorf("EncodeStringBase32(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEncodeStringHex(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello", hex.EncodeToString([]byte("hello"))},
		{"", ""},
		{"abc", "616263"},
	}
	for _, tt := range tests {
		got := EncodeStringHex(tt.input)
		if got != tt.want {
			t.Errorf("EncodeStringHex(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEncodeStringUrl(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello world", url.QueryEscape("hello world")},
		{"no+encoding", url.QueryEscape("no+encoding")},
		{"already%20encoded", url.QueryEscape("already%20encoded")},
		{"special/chars?key=val&foo=bar", url.QueryEscape("special/chars?key=val&foo=bar")},
		{"", ""},
	}
	for _, tt := range tests {
		got := EncodeStringUrl(tt.input)
		if got != tt.want {
			t.Errorf("EncodeStringUrl(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEncodeStringToUTF8Bytes(t *testing.T) {
	tests := []struct {
		input string
		want  []byte
	}{
		{"hello", []byte("hello")},
		{"", []byte("")},
		{"日本語", []byte("日本語")},
	}
	for _, tt := range tests {
		got := EncodeStringToUTF8Bytes(tt.input)
		if string(got) != string(tt.want) {
			t.Errorf("EncodeStringToUTF8Bytes(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestEncodeStringUTF8Hex(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello", hex.EncodeToString([]byte("hello"))},
		{"", ""},
		{"A", "41"},
	}
	for _, tt := range tests {
		got := EncodeStringUTF8Hex(tt.input)
		if got != tt.want {
			t.Errorf("EncodeStringUTF8Hex(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEncodeService_Encode(t *testing.T) {
	svc := NewEncodeService()

	tests := []struct {
		name    string
		method  string
		input   string
		wantErr bool
	}{
		{"base64", "base64", "hello", false},
		{"b64 alias", "b64", "hello", false},
		{"base32", "base32", "hello", false},
		{"b32 alias", "b32", "hello", false},
		{"hex", "hex", "hello", false},
		{"url", "url", "hello world", false},
		{"urlencode alias", "urlencode", "hello world", false},
		{"utf8", "utf8", "hello", false},
		{"utf alias", "utf", "hello", false},
		{"utf8-hex", "utf8-hex", "hello", false},
		{"utf-hex alias", "utf-hex", "hello", false},
		{"case insensitive", "BASE64", "hello", false},
		{"unknown method", "rot13", "hello", true},
		{"empty method", "", "hello", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Encode(tt.method, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode(%q, %q) error = %v, wantErr %v", tt.method, tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" && tt.input != "" {
				t.Errorf("Encode(%q, %q) returned empty string", tt.method, tt.input)
			}
		})
	}
}

// TestEncodeService_Roundtrip verifies base64/hex can roundtrip
func TestEncodeService_Roundtrip(t *testing.T) {
	original := "The quick brown fox jumps over the lazy dog"

	// Base64 roundtrip
	encoded := EncodeStringBase64(original)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("base64 decode error: %v", err)
	}
	if string(decoded) != original {
		t.Errorf("base64 roundtrip failed: got %q", string(decoded))
	}

	// Hex roundtrip
	hexEncoded := EncodeStringHex(original)
	hexDecoded, err := hex.DecodeString(hexEncoded)
	if err != nil {
		t.Fatalf("hex decode error: %v", err)
	}
	if string(hexDecoded) != original {
		t.Errorf("hex roundtrip failed: got %q", string(hexDecoded))
	}
}
